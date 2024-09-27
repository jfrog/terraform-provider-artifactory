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
	sdkv2_diag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"

	"golang.org/x/exp/slices"
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

var DomainSupported = []string{
	ReleaseBundleV2Domain,
	ReleaseBundleV2PromotionDomain,
	UserDomain,
}

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

func (r *WebhookResource) schema(domain string, criteriaBlock *schema.SetNestedBlock) schema.Schema {
	blocks := map[string]schema.Block{
		"handler": schema.SetNestedBlock{
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
						Optional: true,
						// Sensitive: true,
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
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
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
		},
	}

	if criteriaBlock != nil {
		blocks = lo.Assign(
			blocks,
			map[string]schema.Block{
				"criteria": *criteriaBlock,
			},
		)
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

func (r *WebhookResource) Create(ctx context.Context, webhook WebhookAPIModel, req resource.CreateRequest, resp *resource.CreateResponse) {
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.ProviderData.Client.R().
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

func (r *WebhookResource) Read(ctx context.Context, key string, webhook *WebhookAPIModel, req resource.ReadRequest, resp *resource.ReadResponse) (found bool) {
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.ProviderData.Client.R().
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

func (r *WebhookResource) Update(ctx context.Context, key string, webhook WebhookAPIModel, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.ProviderData.Client.R().
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

func (r *WebhookResource) Delete(ctx context.Context, key string, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

type WebhookNoCriteriaResourceModel struct {
	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	EventTypes  types.Set    `tfsdk:"event_types"`
	Handlers    types.Set    `tfsdk:"handler"`
}

type WebhookResourceModel struct {
	WebhookNoCriteriaResourceModel
	Criteria types.Set `tfsdk:"criteria"`
}

func (m WebhookNoCriteriaResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *WebhookAPIModel) (diags diag.Diagnostics) {
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

func (m WebhookResourceModel) toAPIModel(ctx context.Context, domain string, criteriaAPIModel interface{}, apiModel *WebhookAPIModel) (diags diag.Diagnostics) {
	d := m.WebhookNoCriteriaResourceModel.toAPIModel(ctx, domain, apiModel)

	apiModel.EventFilter.Criteria = criteriaAPIModel

	return d
}

func (m *WebhookResourceModel) toBaseCriteriaAPIModel(ctx context.Context, criteriaAttrs map[string]attr.Value) (BaseCriteriaAPIModel, diag.Diagnostics) {
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

func (m *WebhookResourceModel) fromBaseCriteriaAPIModel(ctx context.Context, criteriaAPIModel map[string]interface{}) (map[string]attr.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	includePatterns := types.SetNull(types.StringType)
	if v, ok := criteriaAPIModel["includePatterns"]; ok && v != nil {
		ps, d := types.SetValueFrom(ctx, types.StringType, v)
		if d.HasError() {
			diags.Append(d...)
		}

		includePatterns = ps
	}

	excludePatterns := types.SetNull(types.StringType)
	if v, ok := criteriaAPIModel["excludePatterns"]; ok && v != nil {
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

func (m *WebhookNoCriteriaResourceModel) fromAPIModel(ctx context.Context, apiModel WebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
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

			secret := types.StringNull()
			useSecretForSigning := types.BoolPointerValue(handler.UseSecretForSigning)

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

				// API doesn't include 'use_secret_for_signing' if set to 'false'
				// so need set state to null if attribute is defined in config and set to 'false'
				u := attrs["use_secret_for_signing"].(types.Bool)
				if handler.UseSecretForSigning == nil && !u.IsNull() && !u.ValueBool() {
					useSecretForSigning = types.BoolNull()
				}
			}

			h, d := types.ObjectValue(
				handlerSetResourceModelAttributeTypes,
				map[string]attr.Value{
					"url":                    types.StringValue(handler.Url),
					"secret":                 secret,
					"use_secret_for_signing": useSecretForSigning,
					"proxy":                  types.StringPointerValue(handler.Proxy),
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

func (m *WebhookResourceModel) fromAPIModel(ctx context.Context, apiModel WebhookAPIModel, stateHandlers basetypes.SetValue, criteriaSet *basetypes.SetValue) diag.Diagnostics {
	if criteriaSet != nil {
		m.Criteria = *criteriaSet
	}

	return m.WebhookNoCriteriaResourceModel.fromAPIModel(ctx, apiModel, stateHandlers)
}

type WebhookAPIModel struct {
	Key         string              `json:"key"`
	Description string              `json:"description"`
	Enabled     bool                `json:"enabled"`
	EventFilter EventFilterAPIModel `json:"event_filter"`
	Handlers    []HandlerAPIModel   `json:"handlers"`
}

func (w WebhookAPIModel) Id() string {
	return w.Key
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

var unpackKeyValuePair = func(keyValuePairs map[string]interface{}) []KeyValuePairAPIModel {
	var kvPairs []KeyValuePairAPIModel
	for key, value := range keyValuePairs {
		keyValuePair := KeyValuePairAPIModel{
			Name:  key,
			Value: value.(string),
		}
		kvPairs = append(kvPairs, keyValuePair)
	}

	return kvPairs
}

var packKeyValuePair = func(keyValuePairs []KeyValuePairAPIModel) map[string]interface{} {
	kvPairs := make(map[string]interface{})
	for _, keyValuePair := range keyValuePairs {
		kvPairs[keyValuePair.Name] = keyValuePair.Value
	}

	return kvPairs
}

var domainCriteriaLookup = map[string]interface{}{
	UserDomain:                     EmptyWebhookCriteria{},
	ReleaseBundleV2Domain:          ReleaseBundleV2WebhookCriteria{},
	ReleaseBundleV2PromotionDomain: ReleaseBundleV2PromotionWebhookCriteria{},
}

var domainPackLookup = map[string]func(map[string]interface{}) map[string]interface{}{
	UserDomain:                     packEmptyCriteria,
	ReleaseBundleV2Domain:          packReleaseBundleV2Criteria,
	ReleaseBundleV2PromotionDomain: packReleaseBundleV2PromotionCriteria,
}

var domainUnpackLookup = map[string]func(map[string]interface{}, BaseCriteriaAPIModel) interface{}{
	UserDomain:                     unpackEmptyCriteria,
	ReleaseBundleV2Domain:          unpackReleaseBundleV2Criteria,
	ReleaseBundleV2PromotionDomain: unpackReleaseBundleV2PromotionCriteria,
}

var domainSchemaLookup = func(version int, isCustom bool, webhookType string) map[string]map[string]*sdkv2_schema.Schema {
	return map[string]map[string]*sdkv2_schema.Schema{
		UserDomain:                     userWebhookSchema(webhookType, version, isCustom),
		ReleaseBundleV2Domain:          releaseBundleV2WebhookSchema(webhookType, version, isCustom),
		ReleaseBundleV2PromotionDomain: releaseBundleV2PromotionWebhookSchema(webhookType, version, isCustom),
	}
}

var unpackCriteria = func(d *utilsdk.ResourceData, webhookType string) interface{} {
	var webhookCriteria interface{}

	if v, ok := d.GetOk("criteria"); ok {
		criteria := v.(*sdkv2_schema.Set).List()
		if len(criteria) == 1 {
			id := criteria[0].(map[string]interface{})

			baseCriteria := BaseCriteriaAPIModel{
				IncludePatterns: utilsdk.CastToStringArr(id["include_patterns"].(*sdkv2_schema.Set).List()),
				ExcludePatterns: utilsdk.CastToStringArr(id["exclude_patterns"].(*sdkv2_schema.Set).List()),
			}

			webhookCriteria = domainUnpackLookup[webhookType](id, baseCriteria)
		}
	}

	return webhookCriteria
}

var packCriteria = func(d *sdkv2_schema.ResourceData, webhookType string, criteria map[string]interface{}) []error {
	setValue := utilsdk.MkLens(d)

	resource := domainSchemaLookup(currentSchemaVersion, false, webhookType)[webhookType]["criteria"].Elem.(*sdkv2_schema.Resource)
	packedCriteria := domainPackLookup[webhookType](criteria)

	includePatterns := []interface{}{}
	if v, ok := criteria["includePatterns"]; ok && v != nil {
		includePatterns = v.([]interface{})
	}
	packedCriteria["include_patterns"] = sdkv2_schema.NewSet(sdkv2_schema.HashString, includePatterns)

	excludePatterns := []interface{}{}
	if v, ok := criteria["excludePatterns"]; ok && v != nil {
		excludePatterns = v.([]interface{})
	}
	packedCriteria["exclude_patterns"] = sdkv2_schema.NewSet(sdkv2_schema.HashString, excludePatterns)

	return setValue("criteria", sdkv2_schema.NewSet(sdkv2_schema.HashResource(resource), []interface{}{packedCriteria}))
}

var domainCriteriaValidationLookup = map[string]func(context.Context, map[string]interface{}) error{
	UserDomain:                     emptyCriteriaValidation,
	ReleaseBundleV2Domain:          releaseBundleV2CriteriaValidation,
	ReleaseBundleV2PromotionDomain: emptyCriteriaValidation,
}

var emptyCriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
	return nil
}

var packSecret = func(d *sdkv2_schema.ResourceData, url string) string {
	// Get secret from TF state
	var secret string
	if v, ok := d.GetOk("handler"); ok {
		handlers := v.(*sdkv2_schema.Set).List()
		for _, handler := range handlers {
			h := handler.(map[string]interface{})
			// if urls match, assign the secret value from the state
			if h["url"].(string) == url {
				secret = h["secret"].(string)
			}
		}
	}

	return secret
}

var retryOnProxyError = func(response *resty.Response, _r error) bool {
	var proxyNotFoundRegex = regexp.MustCompile("proxy with key '.*' not found")

	return proxyNotFoundRegex.MatchString(string(response.Body()[:]))
}

func ResourceArtifactoryWebhook(webhookType string) *sdkv2_schema.Resource {

	var unpackWebhook = func(data *sdkv2_schema.ResourceData) (WebhookAPIModel, error) {
		d := &utilsdk.ResourceData{ResourceData: data}

		var unpackHandlers = func(d *utilsdk.ResourceData) []HandlerAPIModel {
			var webhookHandlers []HandlerAPIModel

			if v, ok := d.GetOk("handler"); ok {
				handlers := v.(*sdkv2_schema.Set).List()
				for _, handler := range handlers {
					h := handler.(map[string]interface{})
					// use this to filter out weirdness with terraform adding an extra blank webhook in a set
					// https://discuss.hashicorp.com/t/using-typeset-in-provider-always-adds-an-empty-element-on-update/18566/2
					if h["url"].(string) != "" {
						webhookHandler := HandlerAPIModel{
							HandlerType: "webhook",
							Url:         h["url"].(string),
						}

						if v, ok := h["secret"]; ok {
							if s, ok := v.(string); ok {
								webhookHandler.Secret = &s
							}
						}

						if v, ok := h["use_secret_for_signing"]; ok {
							if b, ok := v.(bool); ok {
								webhookHandler.UseSecretForSigning = &b
							}
						}

						if v, ok := h["proxy"]; ok {
							if s, ok := v.(string); ok {
								webhookHandler.Proxy = &s
							}
						}

						if v, ok := h["custom_http_headers"]; ok {
							webhookHandler.CustomHttpHeaders = unpackKeyValuePair(v.(map[string]interface{}))
						}

						webhookHandlers = append(webhookHandlers, webhookHandler)
					}
				}
			}

			return webhookHandlers
		}

		webhook := WebhookAPIModel{
			Key:         d.GetString("key", false),
			Description: d.GetString("description", false),
			Enabled:     d.GetBool("enabled", false),
			EventFilter: EventFilterAPIModel{
				Domain:     webhookType,
				EventTypes: d.GetSet("event_types"),
				Criteria:   unpackCriteria(d, webhookType),
			},
			Handlers: unpackHandlers(d),
		}

		return webhook, nil
	}

	var packHandlers = func(d *sdkv2_schema.ResourceData, handlers []HandlerAPIModel) []error {
		setValue := utilsdk.MkLens(d)
		resource := domainSchemaLookup(currentSchemaVersion, false, webhookType)[webhookType]["handler"].Elem.(*sdkv2_schema.Resource)
		var packedHandlers []interface{}
		for _, handler := range handlers {
			packedHandler := map[string]interface{}{
				"url":    handler.Url,
				"secret": packSecret(d, handler.Url),
			}

			if handler.UseSecretForSigning != nil {
				packedHandler["use_secret_for_signing"] = *handler.UseSecretForSigning
			}

			if handler.Proxy != nil {
				packedHandler["proxy"] = *handler.Proxy
			}

			if handler.CustomHttpHeaders != nil {
				packedHandler["custom_http_headers"] = packKeyValuePair(handler.CustomHttpHeaders)
			}

			packedHandlers = append(packedHandlers, packedHandler)
		}

		return setValue("handler", sdkv2_schema.NewSet(sdkv2_schema.HashResource(resource), packedHandlers))
	}

	var packWebhook = func(d *sdkv2_schema.ResourceData, webhook WebhookAPIModel) sdkv2_diag.Diagnostics {
		setValue := utilsdk.MkLens(d)

		setValue("key", webhook.Key)
		setValue("description", webhook.Description)
		setValue("enabled", webhook.Enabled)
		errors := setValue("event_types", webhook.EventFilter.EventTypes)
		if webhook.EventFilter.Criteria != nil {
			errors = append(errors, packCriteria(d, webhookType, webhook.EventFilter.Criteria.(map[string]interface{}))...)
		}
		errors = append(errors, packHandlers(d, webhook.Handlers)...)

		if len(errors) > 0 {
			return sdkv2_diag.Errorf("failed to pack webhook %q", errors)
		}

		return nil
	}

	var readWebhook = func(ctx context.Context, data *sdkv2_schema.ResourceData, m interface{}) sdkv2_diag.Diagnostics {
		webhook := WebhookAPIModel{}

		webhook.EventFilter.Criteria = domainCriteriaLookup[webhookType]

		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetResult(&webhook).
			SetError(&artifactoryError).
			Get(WebhookURL)

		if err != nil {
			return sdkv2_diag.FromErr(err)
		}

		if resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		if resp.IsError() {
			return sdkv2_diag.Errorf("%s", artifactoryError.String())
		}

		return packWebhook(data, webhook)
	}

	var createWebhook = func(ctx context.Context, data *sdkv2_schema.ResourceData, m interface{}) sdkv2_diag.Diagnostics {
		webhook, err := unpackWebhook(data)
		if err != nil {
			return sdkv2_diag.FromErr(err)
		}

		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetBody(webhook).
			AddRetryCondition(retryOnProxyError).
			SetError(&artifactoryError).
			Post(webhooksURL)
		if err != nil {
			return sdkv2_diag.FromErr(err)
		}

		if resp.IsError() {
			return sdkv2_diag.Errorf("%s", artifactoryError.String())
		}

		data.SetId(webhook.Id())

		return readWebhook(ctx, data, m)
	}

	var updateWebhook = func(ctx context.Context, data *sdkv2_schema.ResourceData, m interface{}) sdkv2_diag.Diagnostics {
		webhook, err := unpackWebhook(data)
		if err != nil {
			return sdkv2_diag.FromErr(err)
		}

		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetBody(webhook).
			AddRetryCondition(retryOnProxyError).
			SetError(&artifactoryError).
			Put(WebhookURL)
		if err != nil {
			return sdkv2_diag.FromErr(err)
		}

		if resp.IsError() {
			return sdkv2_diag.Errorf("%s", artifactoryError.String())
		}

		data.SetId(webhook.Id())

		return readWebhook(ctx, data, m)
	}

	var deleteWebhook = func(ctx context.Context, data *sdkv2_schema.ResourceData, m interface{}) sdkv2_diag.Diagnostics {
		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetError(&artifactoryError).
			Delete(WebhookURL)

		if err != nil {
			return sdkv2_diag.FromErr(err)
		}

		if resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		if resp.IsError() {
			return sdkv2_diag.Errorf("%s", artifactoryError.String())
		}

		return nil
	}

	var eventTypesDiff = func(ctx context.Context, diff *sdkv2_schema.ResourceDiff, v interface{}) error {
		eventTypes := diff.Get("event_types").(*sdkv2_schema.Set).List()
		if len(eventTypes) == 0 {
			return nil
		}

		eventTypesSupported := DomainEventTypesSupported[webhookType]
		for _, eventType := range eventTypes {
			if !slices.Contains(eventTypesSupported, eventType.(string)) {
				return fmt.Errorf("event_type %s not supported for domain %s", eventType, webhookType)
			}
		}
		return nil
	}

	var criteriaDiff = func(ctx context.Context, diff *sdkv2_schema.ResourceDiff, v interface{}) error {
		if resource, ok := diff.GetOk("criteria"); ok {
			criteria := resource.(*sdkv2_schema.Set).List()
			if len(criteria) == 0 {
				return nil
			}
			return domainCriteriaValidationLookup[webhookType](ctx, criteria[0].(map[string]interface{}))
		}

		return nil
	}

	// Previous version of the schema
	// see example in https://www.terraform.io/plugin/sdkv2/resources/state-migration#terraform-v0-12-sdk-state-migrations
	resourceSchemaV1 := &sdkv2_schema.Resource{
		Schema: domainSchemaLookup(1, false, webhookType)[webhookType],
	}

	rs := sdkv2_schema.Resource{
		SchemaVersion: 2,
		CreateContext: createWebhook,
		ReadContext:   readWebhook,
		UpdateContext: updateWebhook,
		DeleteContext: deleteWebhook,

		Importer: &sdkv2_schema.ResourceImporter{
			StateContext: sdkv2_schema.ImportStatePassthroughContext,
		},

		Schema: domainSchemaLookup(currentSchemaVersion, false, webhookType)[webhookType],
		StateUpgraders: []sdkv2_schema.StateUpgrader{
			{
				Type:    resourceSchemaV1.CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceStateUpgradeV1,
				Version: 1,
			},
		},

		CustomizeDiff: customdiff.All(
			eventTypesDiff,
			criteriaDiff,
		),
		Description: "Provides an Artifactory webhook resource",
	}

	if webhookType == "artifactory_release_bundle" {
		rs.DeprecationMessage = "This resource is being deprecated and replaced by artifactory_destination_webhook resource"
	}

	return &rs
}

// ResourceStateUpgradeV1 see the corresponding unit test TestWebhookResourceStateUpgradeV1
// for more details on the schema transformation
func ResourceStateUpgradeV1(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	rawState["handler"] = []map[string]interface{}{
		{
			"url":                 rawState["url"],
			"secret":              rawState["secret"],
			"proxy":               rawState["proxy"],
			"custom_http_headers": rawState["custom_http_headers"],
		},
	}

	delete(rawState, "url")
	delete(rawState, "secret")
	delete(rawState, "proxy")
	delete(rawState, "custom_http_headers")

	return rawState, nil
}
