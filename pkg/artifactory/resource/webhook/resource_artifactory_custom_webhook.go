package webhook

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

type CustomWebhookResource struct {
	WebhookResource
}

var customHandlerBlock = schema.SetNestedBlock{
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
			"method": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("GET", "POST", "PUT", "PATCH", "DELETE"),
				},
				Description: "Specifies the HTTP Method for URL that the Webhook invokes. Allowed values are: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`.",
			},
			"secrets": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Map{
					mapvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$"), "Secret name must match '^[a-zA-Z_][a-zA-Z0-9_]*$'\""),
					),
				},
				Description: "A set of sensitive values that will be injected in the request (headers and/or payload), comprise of key/value pair.",
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
			"http_headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "HTTP headers you wish to use to invoke the Webhook, comprise of key/value pair. Used in custom webhooks.",
			},
			"payload": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "This attribute is used to build the request body. Used in custom webhooks",
			},
		},
	},
	Validators: []validator.Set{
		setvalidator.IsRequired(),
		setvalidator.SizeAtLeast(1),
	},
}

func (r *CustomWebhookResource) CreateSchema(domain string, criteriaBlock *schema.SetNestedBlock) schema.Schema {
	return r.WebhookResource.CreateSchema(domain, criteriaBlock, customHandlerBlock)
}

func (r *CustomWebhookResource) Create(ctx context.Context, webhook CustomWebhookAPIModel, resp *resource.CreateResponse) {
	createWebhook(r.ProviderData.Client, webhook, resp)
}

func (r *CustomWebhookResource) Read(ctx context.Context, key string, webhook *CustomWebhookAPIModel, resp *resource.ReadResponse) (found bool) {
	return readWebhook(ctx, r.ProviderData.Client, key, webhook, resp)
}

func (r *CustomWebhookResource) Update(_ context.Context, key string, webhook CustomWebhookAPIModel, resp *resource.UpdateResponse) {
	updateWebhook(r.ProviderData.Client, key, webhook, resp)
}

type CustomWebhookBaseResourceModel struct {
	WebhookBaseResourceModel
}

func (m CustomWebhookBaseResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *CustomWebhookAPIModel) (diags diag.Diagnostics) {
	var eventTypes []string
	d := m.EventTypes.ElementsAs(ctx, &eventTypes, false)
	if d.HasError() {
		diags.Append(d...)
	}

	handlers := lo.Map(
		m.Handlers.Elements(),
		func(elem attr.Value, _ int) CustomHandlerAPIModel {
			attrs := elem.(types.Object).Attributes()

			secrets := lo.MapToSlice(
				attrs["secrets"].(types.Map).Elements(),
				func(k string, v attr.Value) KeyValuePairAPIModel {
					return KeyValuePairAPIModel{
						Name:  k,
						Value: v.(types.String).ValueString(),
					}
				},
			)

			httpHeaders := lo.MapToSlice(
				attrs["http_headers"].(types.Map).Elements(),
				func(k string, v attr.Value) KeyValuePairAPIModel {
					return KeyValuePairAPIModel{
						Name:  k,
						Value: v.(types.String).ValueString(),
					}
				},
			)

			return CustomHandlerAPIModel{
				HandlerType: "custom-webhook",
				Url:         attrs["url"].(types.String).ValueString(),
				Method:      attrs["method"].(types.String).ValueString(),
				Secrets:     secrets,
				Proxy:       attrs["proxy"].(types.String).ValueStringPointer(),
				HttpHeaders: httpHeaders,
				Payload:     attrs["payload"].(types.String).ValueString(),
			}
		},
	)

	*apiModel = CustomWebhookAPIModel{
		WebhookAPIModel: WebhookAPIModel{
			Key:         m.Key.ValueString(),
			Description: m.Description.ValueString(),
			Enabled:     m.Enabled.ValueBool(),
			EventFilter: EventFilterAPIModel{
				Domain:     domain,
				EventTypes: eventTypes,
			},
		},
		Handlers: handlers,
	}

	return
}

var customHandlerSetResourceModelAttributeTypes = map[string]attr.Type{
	"url":          types.StringType,
	"method":       types.StringType,
	"secrets":      types.MapType{ElemType: types.StringType},
	"proxy":        types.StringType,
	"http_headers": types.MapType{ElemType: types.StringType},
	"payload":      types.StringType,
}

var customHandlerSetResourceModelElementTypes = types.ObjectType{
	AttrTypes: customHandlerSetResourceModelAttributeTypes,
}

func (m *CustomWebhookBaseResourceModel) fromAPIModel(ctx context.Context, apiModel CustomWebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
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
		func(handler CustomHandlerAPIModel, _ int) attr.Value {
			secrets := types.MapNull(types.StringType)

			matchedHandler, found := lo.Find(
				stateHandlers.Elements(),
				func(elem attr.Value) bool {
					attrs := elem.(types.Object).Attributes()
					return attrs["url"].(types.String).ValueString() == handler.Url
				},
			)
			if found {
				attrs := matchedHandler.(types.Object).Attributes()
				s := attrs["secrets"].(types.Map)
				if !s.IsNull() && len(s.Elements()) > 0 {
					secrets = s
				}
			}

			httpHeaders := types.MapNull(types.StringType)
			if len(handler.HttpHeaders) > 0 {
				headerElems := lo.Associate(
					handler.HttpHeaders,
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

				httpHeaders = h
			}

			h, d := types.ObjectValue(
				customHandlerSetResourceModelAttributeTypes,
				map[string]attr.Value{
					"url":          types.StringValue(handler.Url),
					"method":       types.StringValue(handler.Method),
					"secrets":      secrets,
					"proxy":        types.StringPointerValue(handler.Proxy),
					"http_headers": httpHeaders,
					"payload":      types.StringValue(handler.Payload),
				},
			)
			if d.HasError() {
				diags.Append(d...)
			}

			return h
		},
	)

	handlersSet, d := types.SetValue(
		customHandlerSetResourceModelElementTypes,
		handlers,
	)
	if d.HasError() {
		diags.Append(d...)
	}
	m.Handlers = handlersSet

	return diags
}

type CustomWebhookResourceModel struct {
	CustomWebhookBaseResourceModel
	WebhookCriteriaResourceModel
}

func (m CustomWebhookResourceModel) toAPIModel(ctx context.Context, domain string, criteriaAPIModel interface{}, apiModel *CustomWebhookAPIModel) (diags diag.Diagnostics) {
	d := m.CustomWebhookBaseResourceModel.toAPIModel(ctx, domain, apiModel)

	apiModel.EventFilter.Criteria = criteriaAPIModel

	return d
}

func (m *CustomWebhookResourceModel) fromAPIModel(ctx context.Context, apiModel CustomWebhookAPIModel, stateHandlers basetypes.SetValue, criteriaSet *basetypes.SetValue) diag.Diagnostics {
	if criteriaSet != nil {
		m.Criteria = *criteriaSet
	}

	return m.CustomWebhookBaseResourceModel.fromAPIModel(ctx, apiModel, stateHandlers)
}

type CustomWebhookAPIModel struct {
	WebhookAPIModel
	Handlers []CustomHandlerAPIModel `json:"handlers"`
}

func (w CustomWebhookAPIModel) Id() string {
	return w.Key
}

type CustomHandlerAPIModel struct {
	HandlerType string                 `json:"handler_type"`
	Url         string                 `json:"url"`
	Method      string                 `json:"method"`
	Secrets     []KeyValuePairAPIModel `json:"secrets"`
	Proxy       *string                `json:"proxy"`
	HttpHeaders []KeyValuePairAPIModel `json:"http_headers"`
	Payload     string                 `json:"payload,omitempty"`
}
