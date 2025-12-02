// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

const (
	webhooksURL = "/event/api/v1/subscriptions"
	WebhookURL  = "/event/api/v1/subscriptions/{webhookKey}"

	ArtifactLifecycleDomain        = "artifact_lifecycle"
	ArtifactPropertyDomain         = "artifact_property"
	ArtifactDomain                 = "artifact"
	ArtifactoryReleaseBundleDomain = "artifactory_release_bundle"
	BuildDomain                    = "build"
	DestinationDomain              = "destination"
	DistributionDomain             = "distribution"
	DockerDomain                   = "docker"
	ReleaseBundleDomain            = "release_bundle"
	ReleaseBundleV2Domain          = "release_bundle_v2"
	ReleaseBundleV2PromotionDomain = "release_bundle_v2_promotion"
	UserDomain                     = "user"
)

const currentSchemaVersion = 2

var DomainEventTypesSupported = map[string][]string{
	ArtifactDomain:                 {"deployed", "deleted", "moved", "copied", "cached"},
	ArtifactPropertyDomain:         {"added", "deleted"},
	DockerDomain:                   {"pushed", "deleted", "promoted"},
	BuildDomain:                    {"uploaded", "deleted", "promoted"},
	ReleaseBundleDomain:            {"created", "signed", "deleted"},
	DistributionDomain:             {"distribute_started", "distribute_completed", "distribute_aborted", "distribute_failed", "delete_started", "delete_completed", "delete_failed"},
	ArtifactoryReleaseBundleDomain: {"received", "delete_started", "delete_completed", "delete_failed"},
	DestinationDomain:              {"received", "delete_started", "delete_completed", "delete_failed"},
	UserDomain:                     {"locked"},
	ReleaseBundleV2Domain:          {"release_bundle_v2_started", "release_bundle_v2_failed", "release_bundle_v2_completed"},
	ReleaseBundleV2PromotionDomain: {"release_bundle_v2_promotion_completed", "release_bundle_v2_promotion_failed", "release_bundle_v2_promotion_started"},
	ArtifactLifecycleDomain:        {"archive", "restore"},
}

type WebhookResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
	Domain       string
	Description  string
}

var patternsSchemaAttributes = func(description string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"include_patterns": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: description,
		},
		"exclude_patterns": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: description,
		},
	}
}

var handlerBlock = schema.SetNestedBlock{
	NestedObject: schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validatorfw_string.IsURLHttpOrHttps(),
				},
				Description: "Specifies the URL that the Webhook invokes. This will be the URL that Artifactory will send an HTTP POST request to.",
			},
			"secret": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Secret authentication token that will be sent to the configured URL.",
			},
			"use_secret_for_signing": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "When set to `true`, the secret will be used to sign the event payload, allowing the target to validate that the payload content has not been changed and will not be passed as part of the event. If left unset or set to `false`, the secret is passed through the `X-JFrog-Event-Auth` HTTP header.",
			},
			"proxy": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validatorfw_string.RegexNotMatches(regexp.MustCompile(`^http.+`), "expected \"proxy\" not to be a valid url"),
				},
				Description: "Proxy key from Artifactory Proxies setting",
			},
			"custom_http_headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Custom HTTP headers you wish to use to invoke the Webhook, comprise of key/value pair.",
			},
		},
	},
	Validators: []validator.Set{
		setvalidator.IsRequired(),
		setvalidator.SizeAtLeast(1),
	},
}

func (r *WebhookResource) CreateSchema(domain string, criteriaBlock *schema.SetNestedBlock, handlerBlock schema.SetNestedBlock) schema.Schema {
	blocks := map[string]schema.Block{
		"handler": handlerBlock,
	}

	if criteriaBlock != nil {
		blocks["criteria"] = *criteriaBlock
	}

	return schema.Schema{
		Version: currentSchemaVersion,
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 200),
					stringvalidator.NoneOf(" "),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Key of webhook. Must be between 2 and 200 characters. Cannot contain spaces.",
			},
			"description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
				},
				Description: "Description of webhook. Max length 1000 characters.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Status of webhook. Default to `true`",
			},
			"event_types": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(DomainEventTypesSupported[domain]...),
					),
				},
				Description: fmt.Sprintf("List of Events in Artifactory, Distribution, Release Bundle that function as the event trigger for the Webhook.\n"+
					"Allow values: %v", strings.Trim(strings.Join(DomainEventTypesSupported[ArtifactDomain], ", "), "[]")),
			},
		},
		Blocks:              blocks,
		MarkdownDescription: r.Description,
	}
}

func (r *WebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *WebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func createWebhook[V WebhookAPIModel | CustomWebhookAPIModel](client *resty.Client, webhook V, resp *resource.CreateResponse) {
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := client.R().
		SetBody(webhook).
		SetError(&artifactoryError).
		AddRetryCondition(retryOnProxyError).
		Post(webhooksURL)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, artifactoryError.String())
		return
	}
}

func (r *WebhookResource) Create(_ context.Context, webhook WebhookAPIModel, resp *resource.CreateResponse) {
	createWebhook(r.ProviderData.Client, webhook, resp)
}

func readWebhook[V WebhookAPIModel | CustomWebhookAPIModel](ctx context.Context, client *resty.Client, key string, webhook *V, resp *resource.ReadResponse) (found bool) {
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := client.R().
		SetPathParam("webhookKey", key).
		SetResult(&webhook).
		SetError(&artifactoryError).
		Get(WebhookURL)
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return false
	}

	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return false
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, artifactoryError.String())
		return false
	}

	return true
}

func (r *WebhookResource) Read(ctx context.Context, key string, webhook *WebhookAPIModel, resp *resource.ReadResponse) (found bool) {
	return readWebhook(ctx, r.ProviderData.Client, key, webhook, resp)
}

func updateWebhook[V WebhookAPIModel | CustomWebhookAPIModel](client *resty.Client, key string, webhook V, resp *resource.UpdateResponse) {
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := client.R().
		SetPathParam("webhookKey", key).
		SetBody(webhook).
		AddRetryCondition(retryOnProxyError).
		SetError(&artifactoryError).
		Put(WebhookURL)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, artifactoryError.String())
		return
	}
}

func (r *WebhookResource) Update(_ context.Context, key string, webhook WebhookAPIModel, resp *resource.UpdateResponse) {
	updateWebhook(r.ProviderData.Client, key, webhook, resp)
}

func (r *WebhookResource) Delete(ctx context.Context, key string, resp *resource.DeleteResponse) {
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.ProviderData.Client.R().
		SetPathParam("webhookKey", key).
		SetError(&artifactoryError).
		Delete(WebhookURL)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, artifactoryError.String())
		return
	}
}

// ImportState imports the resource into the Terraform state.
func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

type WebhookBaseResourceModel struct {
	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	EventTypes  types.Set    `tfsdk:"event_types"`
	Handlers    types.Set    `tfsdk:"handler"`
}

func (m WebhookBaseResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *WebhookAPIModel) (diags diag.Diagnostics) {
	var eventTypes []string
	d := m.EventTypes.ElementsAs(ctx, &eventTypes, false)
	if d.HasError() {
		diags.Append(d...)
	}

	handlers := lo.Map(
		m.Handlers.Elements(),
		func(elem attr.Value, _ int) HandlerAPIModel {
			attrs := elem.(types.Object).Attributes()

			customHttpHeaders := lo.MapToSlice(
				attrs["custom_http_headers"].(types.Map).Elements(),
				func(k string, v attr.Value) KeyValuePairAPIModel {
					return KeyValuePairAPIModel{
						Name:  k,
						Value: v.(types.String).ValueString(),
					}
				},
			)

			return HandlerAPIModel{
				HandlerType:         "webhook",
				Url:                 attrs["url"].(types.String).ValueString(),
				Secret:              attrs["secret"].(types.String).ValueStringPointer(),
				UseSecretForSigning: attrs["use_secret_for_signing"].(types.Bool).ValueBoolPointer(),
				Proxy:               attrs["proxy"].(types.String).ValueStringPointer(),
				CustomHttpHeaders:   customHttpHeaders,
			}
		},
	)

	*apiModel = WebhookAPIModel{
		Key:         m.Key.ValueString(),
		Description: m.Description.ValueString(),
		Enabled:     m.Enabled.ValueBool(),
		EventFilter: EventFilterAPIModel{
			Domain:     domain,
			EventTypes: eventTypes,
		},
		Handlers: handlers,
	}

	return
}

var handlerSetResourceModelAttributeTypes = map[string]attr.Type{
	"url":                    types.StringType,
	"secret":                 types.StringType,
	"use_secret_for_signing": types.BoolType,
	"proxy":                  types.StringType,
	"custom_http_headers":    types.MapType{ElemType: types.StringType},
}

var handlerSetResourceModelElementTypes = types.ObjectType{
	AttrTypes: handlerSetResourceModelAttributeTypes,
}

func (m *WebhookBaseResourceModel) fromAPIModel(ctx context.Context, apiModel WebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	m.Key = types.StringValue(apiModel.Key)

	description := types.StringNull()
	if apiModel.Description != "" {
		description = types.StringValue(apiModel.Description)
	}
	m.Description = description

	m.Enabled = types.BoolValue(apiModel.Enabled)

	eventTypes, d := types.SetValueFrom(ctx, types.StringType, apiModel.EventFilter.EventTypes)
	if d.HasError() {
		diags.Append(d...)
	}
	m.EventTypes = eventTypes

	handlers := lo.Map(
		apiModel.Handlers,
		func(handler HandlerAPIModel, _ int) attr.Value {
			customHttpHeaders := types.MapNull(types.StringType)
			if len(handler.CustomHttpHeaders) > 0 {
				headerElems := lo.Associate(
					handler.CustomHttpHeaders,
					func(kvPair KeyValuePairAPIModel) (string, attr.Value) {
						return kvPair.Name, types.StringValue(kvPair.Value)
					},
				)
				h, d := types.MapValue(
					types.StringType,
					headerElems,
				)
				if d.HasError() {
					diags.Append(d...)
				}

				customHttpHeaders = h
			}

			useSecretForSigning := types.BoolNull()
			if handler.UseSecretForSigning != nil {
				useSecretForSigning = types.BoolPointerValue(handler.UseSecretForSigning)
			}

			secret := types.StringNull()
			matchedHandler, found := lo.Find(
				stateHandlers.Elements(),
				func(elem attr.Value) bool {
					attrs := elem.(types.Object).Attributes()
					return attrs["url"].(types.String).ValueString() == handler.Url
				},
			)
			if found {
				attrs := matchedHandler.(types.Object).Attributes()
				s := attrs["secret"].(types.String)
				if !s.IsNull() && s.ValueString() != "" {
					secret = s
				}
			}

			proxy := types.StringNull()
			if handler.Proxy != nil {
				proxy = types.StringPointerValue(handler.Proxy)
			}

			h, d := types.ObjectValue(
				handlerSetResourceModelAttributeTypes,
				map[string]attr.Value{
					"url":                    types.StringValue(handler.Url),
					"secret":                 secret,
					"use_secret_for_signing": useSecretForSigning,
					"proxy":                  proxy,
					"custom_http_headers":    customHttpHeaders,
				},
			)
			if d.HasError() {
				diags.Append(d...)
			}

			return h
		},
	)

	handlersSet, d := types.SetValue(
		handlerSetResourceModelElementTypes,
		handlers,
	)
	if d.HasError() {
		diags.Append(d...)
	}
	m.Handlers = handlersSet

	return diags
}

type WebhookCriteriaResourceModel struct {
	Criteria types.Set `tfsdk:"criteria"`
}

func (m *WebhookCriteriaResourceModel) toBaseCriteriaAPIModel(ctx context.Context, criteriaAttrs map[string]attr.Value) (BaseCriteriaAPIModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var includePatterns []string
	d := criteriaAttrs["include_patterns"].(types.Set).ElementsAs(ctx, &includePatterns, false)
	if d.HasError() {
		diags.Append(d...)
	}

	var excludePatterns []string
	d = criteriaAttrs["exclude_patterns"].(types.Set).ElementsAs(ctx, &excludePatterns, false)
	if d.HasError() {
		diags.Append(d...)
	}

	return BaseCriteriaAPIModel{
		IncludePatterns: includePatterns,
		ExcludePatterns: excludePatterns,
	}, diags
}

var patternsCriteriaSetResourceModelAttributeTypes = map[string]attr.Type{
	"include_patterns": types.SetType{ElemType: types.StringType},
	"exclude_patterns": types.SetType{ElemType: types.StringType},
}

func (m *WebhookCriteriaResourceModel) fromBaseCriteriaAPIModel(ctx context.Context, criteriaAPIModel map[string]interface{}) (map[string]attr.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	includePatterns := types.SetNull(types.StringType)
	if v, ok := criteriaAPIModel["includePatterns"]; ok && v != nil && len(v.([]interface{})) > 0 {
		ps, d := types.SetValueFrom(ctx, types.StringType, v)
		if d.HasError() {
			diags.Append(d...)
		}

		includePatterns = ps
	}

	excludePatterns := types.SetNull(types.StringType)
	if v, ok := criteriaAPIModel["excludePatterns"]; ok && v != nil && len(v.([]interface{})) > 0 {
		ps, d := types.SetValueFrom(ctx, types.StringType, v)
		if d.HasError() {
			diags.Append(d...)
		}

		excludePatterns = ps
	}

	return map[string]attr.Value{
		"include_patterns": includePatterns,
		"exclude_patterns": excludePatterns,
	}, diags
}

type WebhookResourceModel struct {
	WebhookBaseResourceModel
	WebhookCriteriaResourceModel
}

func (m WebhookResourceModel) toAPIModel(ctx context.Context, domain string, criteriaAPIModel interface{}, apiModel *WebhookAPIModel) (diags diag.Diagnostics) {
	d := m.WebhookBaseResourceModel.toAPIModel(ctx, domain, apiModel)

	apiModel.EventFilter.Criteria = criteriaAPIModel

	return d
}

func (m *WebhookResourceModel) fromAPIModel(ctx context.Context, apiModel WebhookAPIModel, stateHandlers basetypes.SetValue, criteriaSet *basetypes.SetValue) diag.Diagnostics {
	if criteriaSet != nil {
		m.Criteria = *criteriaSet
	}

	return m.WebhookBaseResourceModel.fromAPIModel(ctx, apiModel, stateHandlers)
}

type WebhookAPIModel struct {
	Key         string              `json:"key"`
	Description string              `json:"description"`
	Enabled     bool                `json:"enabled"`
	EventFilter EventFilterAPIModel `json:"event_filter"`
	Handlers    []HandlerAPIModel   `json:"handlers"`
}

type EventFilterAPIModel struct {
	Domain     string      `json:"domain"`
	EventTypes []string    `json:"event_types"`
	Criteria   interface{} `json:"criteria,omitempty"`
}

type BaseCriteriaAPIModel struct {
	IncludePatterns []string `json:"includePatterns"`
	ExcludePatterns []string `json:"excludePatterns"`
}

type HandlerAPIModel struct {
	HandlerType         string                 `json:"handler_type"`
	Url                 string                 `json:"url"`
	Secret              *string                `json:"secret"`
	UseSecretForSigning *bool                  `json:"use_secret_for_signing,omitempty"`
	Proxy               *string                `json:"proxy"`
	CustomHttpHeaders   []KeyValuePairAPIModel `json:"custom_http_headers"`
}

type KeyValuePairAPIModel struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var retryOnProxyError = func(response *resty.Response, _r error) bool {
	var proxyNotFoundRegex = regexp.MustCompile("proxy with key '.*' not found")

	return proxyNotFoundRegex.MatchString(string(response.Body()[:]))
}
