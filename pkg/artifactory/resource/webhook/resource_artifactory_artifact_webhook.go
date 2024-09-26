package webhook

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

var _ resource.Resource = &ArtifactWebhookResource{}

func NewArtifactWebhookResource() resource.Resource {
	return &ArtifactWebhookResource{
		TypeName: "artifactory_artifact_webhook",
	}
}

type ArtifactWebhookResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type ArtifactWebhookResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	EventTypes  types.Set    `tfsdk:"event_types"`
	Criteria    types.Set    `tfsdk:"criteria"`
	Handlers    types.Set    `tfsdk:"handler"`
}

func (m ArtifactWebhookResourceModel) toAPIModel(ctx context.Context, apiModel *ArtifactWebhookAPIModel) (diags diag.Diagnostics) {
	critieriaObj := m.Criteria.Elements()[0].(types.Object)
	critieriaAttrs := critieriaObj.Attributes()

	var includePatterns []string
	d := critieriaAttrs["include_patterns"].(types.Set).ElementsAs(ctx, &includePatterns, false)
	if d.HasError() {
		diags.Append(d...)
	}

	var excludePatterns []string
	d = critieriaAttrs["exclude_patterns"].(types.Set).ElementsAs(ctx, &excludePatterns, false)
	if d.HasError() {
		diags.Append(d...)
	}

	var repoKeys []string
	d = critieriaAttrs["repo_keys"].(types.Set).ElementsAs(ctx, &repoKeys, false)
	if d.HasError() {
		diags.Append(d...)
	}

	var eventTypes []string
	d = m.EventTypes.ElementsAs(ctx, &eventTypes, false)
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
				Secret:              attrs["secret"].(types.String).ValueString(),
				UseSecretForSigning: attrs["use_secret_for_signing"].(types.Bool).ValueBool(),
				Proxy:               attrs["proxy"].(types.String).ValueString(),
				CustomHttpHeaders:   customHttpHeaders,
			}
		},
	)

	*apiModel = ArtifactWebhookAPIModel{
		BaseAPIModel: BaseAPIModel{
			Key:         m.Key.ValueString(),
			Description: m.Description.ValueString(),
			Enabled:     m.Enabled.ValueBool(),
			EventFilter: EventFilterAPIModel{
				Domain:     ArtifactType,
				EventTypes: eventTypes,
				Criteria: RepoCriteriaAPIModel{
					BaseCriteriaAPIModel: BaseCriteriaAPIModel{
						IncludePatterns: includePatterns,
						ExcludePatterns: excludePatterns,
					},
					AnyLocal:     critieriaAttrs["any_local"].(types.Bool).ValueBool(),
					AnyRemote:    critieriaAttrs["any_remote"].(types.Bool).ValueBool(),
					AnyFederated: critieriaAttrs["any_federated"].(types.Bool).ValueBool(),
					RepoKeys:     repoKeys,
				},
			},
			Handlers: handlers,
		},
	}

	return
}

var criteriaSetResourceModelAttributeTypes = map[string]attr.Type{
	"include_patterns": types.SetType{ElemType: types.StringType},
	"exclude_patterns": types.SetType{ElemType: types.StringType},
	"any_local":        types.BoolType,
	"any_remote":       types.BoolType,
	"any_federated":    types.BoolType,
	"repo_keys":        types.SetType{ElemType: types.StringType},
}

var criteriaSetResourceModelElementTypes = types.ObjectType{
	AttrTypes: criteriaSetResourceModelAttributeTypes,
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

func (m *ArtifactWebhookResourceModel) fromAPIModel(ctx context.Context, apiModel ArtifactWebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	m.ID = types.StringValue(apiModel.Key)
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

	criteriaAPIModel := apiModel.EventFilter.Criteria.(map[string]interface{})

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

	repoKeys := types.SetNull(types.StringType)
	if v, ok := criteriaAPIModel["repoKeys"]; ok && v != nil {
		ks, d := types.SetValueFrom(ctx, types.StringType, v)
		if d.HasError() {
			diags.Append(d...)
		}

		repoKeys = ks
	}

	criteria, d := types.ObjectValue(
		criteriaSetResourceModelAttributeTypes,
		map[string]attr.Value{
			"include_patterns": includePatterns,
			"exclude_patterns": excludePatterns,
			"any_local":        types.BoolValue(criteriaAPIModel["anyLocal"].(bool)),
			"any_remote":       types.BoolValue(criteriaAPIModel["anyRemote"].(bool)),
			"any_federated":    types.BoolValue(criteriaAPIModel["anyFederated"].(bool)),
			"repo_keys":        repoKeys,
		},
	)
	if d.HasError() {
		diags.Append(d...)
	}
	criteriaSet, d := types.SetValue(
		criteriaSetResourceModelElementTypes,
		[]attr.Value{criteria},
	)
	if d.HasError() {
		diags.Append(d...)
	}
	m.Criteria = criteriaSet

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

			secret := types.StringValue("")
			matchedHandler, found := lo.Find(
				stateHandlers.Elements(),
				func(elem attr.Value) bool {
					attrs := elem.(types.Object).Attributes()
					return attrs["url"].(types.String).ValueString() == handler.Url
				},
			)
			if found {
				attrs := matchedHandler.(types.Object).Attributes()
				if !attrs["secret"].(types.String).IsNull() {
					secret = attrs["secret"].(types.String)
				}
			}

			proxy := types.StringNull()
			if handler.Proxy != "" {
				proxy = types.StringValue(handler.Proxy)
			}

			h, d := types.ObjectValue(
				handlerSetResourceModelAttributeTypes,
				map[string]attr.Value{
					"url":                    types.StringValue(handler.Url),
					"secret":                 secret,
					"use_secret_for_signing": types.BoolValue(handler.UseSecretForSigning),
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

type ArtifactWebhookAPIModel struct {
	BaseAPIModel
}

func (r *ArtifactWebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *ArtifactWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 2,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{ // for backward compatability
				Computed: true,
			},
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
						stringvalidator.OneOf(DomainEventTypesSupported[ArtifactType]...),
					),
				},
				Description: fmt.Sprintf("List of Events in Artifactory, Distribution, Release Bundle that function as the event trigger for the Webhook.\n"+
					"Allow values: %v", strings.Trim(strings.Join(DomainEventTypesSupported[ArtifactType], ", "), "[]")),
			},
		},
		Blocks: map[string]schema.Block{
			"criteria": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"include_patterns": schema.SetAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							MarkdownDescription: "Simple comma separated wildcard patterns for repository artifact paths (with no leading slash).\nAnt-style path expressions are supported (*, **, ?).\nFor example: `org/apache/**`",
						},
						"exclude_patterns": schema.SetAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							MarkdownDescription: "Simple comma separated wildcard patterns for repository artifact paths (with no leading slash).\nAnt-style path expressions are supported (*, **, ?).\nFor example: `org/apache/**`",
						},
						"any_local": schema.BoolAttribute{
							Required:    true,
							Description: "Trigger on any local repositories",
						},
						"any_remote": schema.BoolAttribute{
							Required:    true,
							Description: "Trigger on any remote repositories",
						},
						"any_federated": schema.BoolAttribute{
							Required:    true,
							Description: "Trigger on any federated repositories",
						},
						"repo_keys": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "Trigger on this list of repository keys",
						},
					},
				},
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 1),
					setvalidator.IsRequired(),
				},
				Description: "Specifies where the webhook will be applied on which repositories.",
			},
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
							Computed: true,
							Default:  stringdefault.StaticString(""), // for backward compatability
							// Sensitive: true,
							Description: "Secret authentication token that will be sent to the configured URL.",
						},
						"use_secret_for_signing": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
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
			},
		},
		MarkdownDescription: "This resource enables you to creates a new Release Bundle v2, uniquely identified by a combination of repository key, name, and version. For more information, see [Understanding Release Bundles v2](https://jfrog.com/help/r/jfrog-artifactory-documentation/understanding-release-bundles-v2) and [REST API](https://jfrog.com/help/r/jfrog-rest-apis/create-release-bundle-v2-version).",
	}
}

func (r *ArtifactWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r ArtifactWebhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ArtifactWebhookResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	criteriaObj := data.Criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	anyLocal := criteriaAttrs["any_local"].(types.Bool).ValueBool()
	anyRemote := criteriaAttrs["any_remote"].(types.Bool).ValueBool()
	anyFederated := criteriaAttrs["any_federated"].(types.Bool).ValueBool()

	if (!anyLocal && !anyRemote && !anyFederated) && len(criteriaAttrs["repo_keys"].(types.Set).Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("repo_keys"),
			"Invalid Attribute Configuration",
			"repo_keys cannot be empty when any_local, any_remote, and any_federated are false",
		)
	}
}

func (r *ArtifactWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ArtifactWebhookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook ArtifactWebhookAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &webhook)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	plan.ID = plan.Key

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ArtifactWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ArtifactWebhookResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook ArtifactWebhookAPIModel
	var artifactoryError artifactory.ArtifactoryErrorsResponse

	response, err := r.ProviderData.Client.R().
		SetPathParam("webhookKey", state.Key.ValueString()).
		SetResult(&webhook).
		SetError(&artifactoryError).
		Get(WebhookURL)
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, artifactoryError.String())
		return
	}

	resp.Diagnostics.Append(state.fromAPIModel(ctx, webhook, state.Handlers)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ArtifactWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ArtifactWebhookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook ArtifactWebhookAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &webhook)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.ProviderData.Client.R().
		SetPathParam("webhookKey", plan.Key.ValueString()).
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

	plan.ID = plan.Key

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ArtifactWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ArtifactWebhookResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.ProviderData.Client.R().
		SetPathParam("webhookKey", state.Key.ValueString()).
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

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ArtifactWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
