package webhook

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
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
	sdkv2_diag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	util_validator "github.com/jfrog/terraform-provider-shared/validator"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"

	"golang.org/x/exp/slices"
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

var DomainSupported = []string{
	ReleaseBundleV2PromotionDomain,
	UserDomain,
}

func baseCustomWebhookBaseSchema(webhookType string) map[string]*sdkv2_schema.Schema {
	return map[string]*sdkv2_schema.Schema{
		"key": {
			Type:     sdkv2_schema.TypeString,
			Required: true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.All(
					validation.StringLenBetween(2, 200),
					validation.StringDoesNotContainAny(" "),
				),
			),
			Description: "Key of webhook. Must be between 2 and 200 characters. Cannot contain spaces.",
		},
		"description": {
			Type:             sdkv2_schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 1000)),
			Description:      "Description of webhook. Max length 1000 characters.",
		},
		"enabled": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Status of webhook. Default to 'true'",
		},
		"event_types": {
			Type:     sdkv2_schema.TypeSet,
			Required: true,
			MinItems: 1,
			Elem:     &sdkv2_schema.Schema{Type: sdkv2_schema.TypeString},
			Description: fmt.Sprintf("List of Events in Artifactory, Distribution, Release Bundle that function as the event trigger for the Webhook.\n"+
				"Allow values: %v", strings.Trim(strings.Join(DomainEventTypesSupported[webhookType], ", "), "[]")),
		},
		"handler": {
			Type:     sdkv2_schema.TypeSet,
			Required: true,
			MinItems: 1,
			Elem: &sdkv2_schema.Resource{
				Schema: map[string]*sdkv2_schema.Schema{
					"url": {
						Type:     sdkv2_schema.TypeString,
						Required: true,
						ValidateDiagFunc: validation.ToDiagFunc(
							validation.All(
								validation.IsURLWithHTTPorHTTPS,
								validation.StringIsNotEmpty,
							),
						),
						Description: "Specifies the URL that the Webhook invokes. This will be the URL that Artifactory will send an HTTP POST request to.",
					},
					"secrets": {
						Type:     sdkv2_schema.TypeMap,
						Optional: true,
						Elem: &sdkv2_schema.Schema{
							Type:             sdkv2_schema.TypeString,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringMatch(regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$"), "Secret name must match '^[a-zA-Z_][a-zA-Z0-9_]*$'\"")),
						},
						Description: "A set of sensitive values that will be injected in the request (headers and/or payload), comprise of key/value pair.",
					},
					"proxy": {
						Type:     sdkv2_schema.TypeString,
						Optional: true,
						ValidateDiagFunc: util_validator.All(
							util_validator.StringIsNotEmpty,
							util_validator.StringIsNotURL,
						),
						Description: "Proxy key from Artifactory UI (Administration -> Proxies -> Configuration)",
					},
					"http_headers": {
						Type:        sdkv2_schema.TypeMap,
						Optional:    true,
						Elem:        &sdkv2_schema.Schema{Type: sdkv2_schema.TypeString},
						Description: "HTTP headers you wish to use to invoke the Webhook, comprise of key/value pair. Used in custom webhooks.",
					},
					"payload": {
						Type:             sdkv2_schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
						Description:      "This attribute is used to build the request body. Used in custom webhooks",
					},
				},
			},
		},
	}
}

var releaseBundleV2PromotionWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*sdkv2_schema.Schema {
	return utilsdk.MergeMaps(getBaseSchemaByVersion(webhookType, version, isCustom), map[string]*sdkv2_schema.Schema{
		"criteria": {
			Type:     sdkv2_schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &sdkv2_schema.Resource{
				Schema: utilsdk.MergeMaps(baseCriteriaSchema, map[string]*sdkv2_schema.Schema{
					"selected_environments": {
						Type:        sdkv2_schema.TypeSet,
						Required:    true,
						Elem:        &sdkv2_schema.Schema{Type: sdkv2_schema.TypeString},
						Description: "Trigger on this list of environments",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied, on which release bundles promotion.",
		},
	})
}

var userWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*sdkv2_schema.Schema {
	return getBaseSchemaByVersion(webhookType, version, isCustom)
}

var artifactLifecycleWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*sdkv2_schema.Schema {
	return getBaseSchemaByVersion(webhookType, version, isCustom)
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
	Secrets     []KeyValuePairAPIModel `json:"secrets"`
	Proxy       *string                `json:"proxy"`
	HttpHeaders []KeyValuePairAPIModel `json:"http_headers"`
	Payload     string                 `json:"payload,omitempty"`
}

type SecretName struct {
	Name string `json:"name"`
}

var packSecretsCustom = func(keyValuePairs []KeyValuePairAPIModel, d *sdkv2_schema.ResourceData, url string) map[string]interface{} {
	KVPairs := make(map[string]interface{})
	// Get secrets from TF state
	var secrets map[string]interface{}
	if v, ok := d.GetOk("handler"); ok {
		handlers := v.(*sdkv2_schema.Set).List()
		for _, handler := range handlers {
			h := handler.(map[string]interface{})
			// if url match, merge secret maps
			if h["url"] == url {
				secrets = utilsdk.MergeMaps(secrets, h["secrets"].(map[string]interface{}))
			}
		}
	}
	// We assign secret the value from the state, because it's not returned in the API body response
	for _, keyValuePair := range keyValuePairs {
		if v, ok := secrets[keyValuePair.Name]; ok {
			KVPairs[keyValuePair.Name] = v.(string)
		}
	}

	return KVPairs
}

func ResourceArtifactoryCustomWebhook(webhookType string) *sdkv2_schema.Resource {

	var unpackWebhook = func(data *sdkv2_schema.ResourceData) (CustomWebhookAPIModel, error) {
		d := &utilsdk.ResourceData{ResourceData: data}

		var unpackHandlers = func(d *utilsdk.ResourceData) []CustomHandlerAPIModel {
			var webhookHandlers []CustomHandlerAPIModel

			if v, ok := d.GetOk("handler"); ok {
				handlers := v.(*sdkv2_schema.Set).List()
				for _, handler := range handlers {
					h := handler.(map[string]interface{})
					// use this to filter out weirdness with terraform adding an extra blank webhook in a set
					// https://discuss.hashicorp.com/t/using-typeset-in-provider-always-adds-an-empty-element-on-update/18566/2
					if h["url"].(string) != "" {
						webhookHandler := CustomHandlerAPIModel{
							HandlerType: "custom-webhook",
							Url:         h["url"].(string),
						}

						if v, ok := h["secrets"]; ok {
							webhookHandler.Secrets = unpackKeyValuePair(v.(map[string]interface{}))
						}

						if v, ok := h["proxy"]; ok {
							if s, ok := v.(string); ok {
								webhookHandler.Proxy = &s
							}
						}

						if v, ok := h["http_headers"]; ok {
							webhookHandler.HttpHeaders = unpackKeyValuePair(v.(map[string]interface{}))
						}

						if v, ok := h["payload"]; ok {
							webhookHandler.Payload = v.(string)
						}

						webhookHandlers = append(webhookHandlers, webhookHandler)
					}
				}
			}

			return webhookHandlers
		}

		webhook := CustomWebhookAPIModel{
			WebhookAPIModel: WebhookAPIModel{
				Key:         d.GetString("key", false),
				Description: d.GetString("description", false),
				Enabled:     d.GetBool("enabled", false),
				EventFilter: EventFilterAPIModel{
					Domain:     webhookType,
					EventTypes: d.GetSet("event_types"),
					Criteria:   unpackCriteria(d, webhookType),
				},
			},
			Handlers: unpackHandlers(d),
		}

		return webhook, nil
	}

	var packHandlers = func(d *sdkv2_schema.ResourceData, handlers []CustomHandlerAPIModel) []error {
		setValue := utilsdk.MkLens(d)
		resource := domainSchemaLookup(currentSchemaVersion, true, webhookType)[webhookType]["handler"].Elem.(*sdkv2_schema.Resource)
		packedHandlers := make([]interface{}, len(handlers))
		for _, handler := range handlers {
			packedHandler := map[string]interface{}{
				"url":     handler.Url,
				"payload": handler.Payload,
			}

			if handler.Proxy != nil {
				packedHandler["proxy"] = *handler.Proxy
			}

			if handler.Secrets != nil {
				packedHandler["secrets"] = packSecretsCustom(handler.Secrets, d, handler.Url)
			}

			if handler.HttpHeaders != nil {
				packedHandler["http_headers"] = packKeyValuePair(handler.HttpHeaders)
			}

			packedHandlers = append(packedHandlers, packedHandler)
		}

		return setValue("handler", sdkv2_schema.NewSet(sdkv2_schema.HashResource(resource), packedHandlers))
	}

	var packWebhook = func(d *sdkv2_schema.ResourceData, webhook CustomWebhookAPIModel) sdkv2_diag.Diagnostics {
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
		webhook := CustomWebhookAPIModel{}

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

	var retryOnProxyError = func(response *resty.Response, _r error) bool {
		var proxyNotFoundRegex = regexp.MustCompile("proxy with key '.*' not found")

		return proxyNotFoundRegex.MatchString(string(response.Body()[:]))
	}

	var createWebhook = func(ctx context.Context, data *sdkv2_schema.ResourceData, m interface{}) sdkv2_diag.Diagnostics {
		webhook, err := unpackWebhook(data)
		if err != nil {
			return sdkv2_diag.FromErr(err)
		}

		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetBody(webhook).
			SetError(&artifactoryError).
			AddRetryCondition(retryOnProxyError).
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
			SetError(&artifactoryError).
			AddRetryCondition(retryOnProxyError).
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

	rs := sdkv2_schema.Resource{
		SchemaVersion: 2,
		CreateContext: createWebhook,
		ReadContext:   readWebhook,
		UpdateContext: updateWebhook,
		DeleteContext: deleteWebhook,

		Importer: &sdkv2_schema.ResourceImporter{
			StateContext: sdkv2_schema.ImportStatePassthroughContext,
		},

		Schema: domainSchemaLookup(currentSchemaVersion, true, webhookType)[webhookType],

		CustomizeDiff: customdiff.All(
			eventTypesDiff,
			criteriaDiff,
		),
		Description: "Provides an Artifactory webhook resource",
	}

	if webhookType == ReleaseBundleDomain {
		rs.DeprecationMessage = "This resource is being deprecated and replaced by artifactory_destination_custom_webhook resource"
	}

	return &rs
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

type EmptyWebhookCriteria struct{}

var domainCriteriaLookup = map[string]interface{}{
	"user":                        EmptyWebhookCriteria{},
	"release_bundle_v2_promotion": ReleaseBundleV2PromotionCriteriaAPIModel{},
}

var domainPackLookup = map[string]func(map[string]interface{}) map[string]interface{}{
	"user":                        packEmptyCriteria,
	"release_bundle_v2_promotion": packReleaseBundleV2PromotionCriteria,
}

var domainUnpackLookup = map[string]func(map[string]interface{}, BaseCriteriaAPIModel) interface{}{
	"user":                        unpackEmptyCriteria,
	"release_bundle_v2_promotion": unpackReleaseBundleV2PromotionCriteria,
}

var domainSchemaLookup = func(version int, isCustom bool, webhookType string) map[string]map[string]*sdkv2_schema.Schema {
	return map[string]map[string]*sdkv2_schema.Schema{
		"user":                        userWebhookSchema(webhookType, version, isCustom),
		"release_bundle_v2_promotion": releaseBundleV2PromotionWebhookSchema(webhookType, version, isCustom),
	}
}

var packReleaseBundleV2PromotionCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"selected_environments": sdkv2_schema.NewSet(sdkv2_schema.HashString, artifactoryCriteria["selectedEnvironments"].([]interface{})),
	}
}

var unpackReleaseBundleV2PromotionCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseCriteriaAPIModel) interface{} {
	return ReleaseBundleV2PromotionCriteriaAPIModel{
		SelectedEnvironments: utilsdk.CastToStringArr(terraformCriteria["selected_environments"].(*sdkv2_schema.Set).List()),
	}
}

var packEmptyCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{}
}

var unpackEmptyCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseCriteriaAPIModel) interface{} {
	return EmptyWebhookCriteria{}
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
	"user":                        emptyCriteriaValidation,
	"release_bundle_v2_promotion": emptyCriteriaValidation,
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
