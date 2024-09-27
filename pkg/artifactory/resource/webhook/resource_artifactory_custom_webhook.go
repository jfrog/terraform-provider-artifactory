package webhook

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"

	"golang.org/x/exp/slices"
)

func baseCustomWebhookBaseSchema(webhookType string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"key": {
			Type:     schema.TypeString,
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
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 1000)),
			Description:      "Description of webhook. Max length 1000 characters.",
		},
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Status of webhook. Default to 'true'",
		},
		"event_types": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Description: fmt.Sprintf("List of Events in Artifactory, Distribution, Release Bundle that function as the event trigger for the Webhook.\n"+
				"Allow values: %v", strings.Trim(strings.Join(DomainEventTypesSupported[webhookType], ", "), "[]")),
		},
		"handler": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:     schema.TypeString,
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
						Type:     schema.TypeMap,
						Optional: true,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringMatch(regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$"), "Secret name must match '^[a-zA-Z_][a-zA-Z0-9_]*$'\"")),
						},
						Description: "A set of sensitive values that will be injected in the request (headers and/or payload), comprise of key/value pair.",
					},
					"proxy": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateDiagFunc: validator.All(
							validator.StringIsNotEmpty,
							validator.StringIsNotURL,
						),
						Description: "Proxy key from Artifactory UI (Administration -> Proxies -> Configuration)",
					},
					"http_headers": {
						Type:        schema.TypeMap,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "HTTP headers you wish to use to invoke the Webhook, comprise of key/value pair. Used in custom webhooks.",
					},
					"payload": {
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
						Description:      "This attribute is used to build the request body. Used in custom webhooks",
					},
				},
			},
		},
	}
}

type CustomBaseParams struct {
	Key                 string              `json:"key"`
	Description         string              `json:"description"`
	Enabled             bool                `json:"enabled"`
	EventFilterAPIModel EventFilterAPIModel `json:"event_filter"`
	Handlers            []CustomHandler     `json:"handlers"`
}

func (w CustomBaseParams) Id() string {
	return w.Key
}

type CustomHandler struct {
	HandlerType string                 `json:"handler_type"`
	Url         string                 `json:"url"`
	Secrets     []KeyValuePairAPIModel `json:"secrets"`
	Proxy       string                 `json:"proxy"`
	HttpHeaders []KeyValuePairAPIModel `json:"http_headers"`
	Payload     string                 `json:"payload,omitempty"`
}

type SecretName struct {
	Name string `json:"name"`
}

var packSecretsCustom = func(keyValuePairs []KeyValuePairAPIModel, d *schema.ResourceData, url string) map[string]interface{} {
	KVPairs := make(map[string]interface{})
	// Get secrets from TF state
	var secrets map[string]interface{}
	if v, ok := d.GetOk("handler"); ok {
		handlers := v.(*schema.Set).List()
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

func ResourceArtifactoryCustomWebhook(webhookType string) *schema.Resource {

	var unpackWebhook = func(data *schema.ResourceData) (CustomBaseParams, error) {
		d := &utilsdk.ResourceData{ResourceData: data}

		var unpackHandlers = func(d *utilsdk.ResourceData) []CustomHandler {
			var webhookHandlers []CustomHandler

			if v, ok := d.GetOk("handler"); ok {
				handlers := v.(*schema.Set).List()
				for _, handler := range handlers {
					h := handler.(map[string]interface{})
					// use this to filter out weirdness with terraform adding an extra blank webhook in a set
					// https://discuss.hashicorp.com/t/using-typeset-in-provider-always-adds-an-empty-element-on-update/18566/2
					if h["url"].(string) != "" {
						webhookHandler := CustomHandler{
							HandlerType: "custom-webhook",
							Url:         h["url"].(string),
						}

						if v, ok := h["secrets"]; ok {
							webhookHandler.Secrets = unpackKeyValuePair(v.(map[string]interface{}))
						}

						if v, ok := h["proxy"]; ok {
							webhookHandler.Proxy = v.(string)
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

		webhook := CustomBaseParams{
			Key:         d.GetString("key", false),
			Description: d.GetString("description", false),
			Enabled:     d.GetBool("enabled", false),
			EventFilterAPIModel: EventFilterAPIModel{
				Domain:     webhookType,
				EventTypes: d.GetSet("event_types"),
				Criteria:   unpackCriteria(d, webhookType),
			},
			Handlers: unpackHandlers(d),
		}

		return webhook, nil
	}

	var packHandlers = func(d *schema.ResourceData, handlers []CustomHandler) []error {
		setValue := utilsdk.MkLens(d)
		resource := domainSchemaLookup(currentSchemaVersion, true, webhookType)[webhookType]["handler"].Elem.(*schema.Resource)
		packedHandlers := make([]interface{}, len(handlers))
		for _, handler := range handlers {
			packedHandler := map[string]interface{}{
				"url":     handler.Url,
				"proxy":   handler.Proxy,
				"payload": handler.Payload,
			}

			if handler.Secrets != nil {
				packedHandler["secrets"] = packSecretsCustom(handler.Secrets, d, handler.Url)
			}

			if handler.HttpHeaders != nil {
				packedHandler["http_headers"] = packKeyValuePair(handler.HttpHeaders)
			}

			packedHandlers = append(packedHandlers, packedHandler)
		}

		return setValue("handler", schema.NewSet(schema.HashResource(resource), packedHandlers))
	}

	var packWebhook = func(d *schema.ResourceData, webhook CustomBaseParams) diag.Diagnostics {
		setValue := utilsdk.MkLens(d)

		setValue("key", webhook.Key)
		setValue("description", webhook.Description)
		setValue("enabled", webhook.Enabled)
		errors := setValue("event_types", webhook.EventFilterAPIModel.EventTypes)
		if webhook.EventFilterAPIModel.Criteria != nil {
			errors = append(errors, packCriteria(d, webhookType, webhook.EventFilterAPIModel.Criteria.(map[string]interface{}))...)
		}
		errors = append(errors, packHandlers(d, webhook.Handlers)...)

		if len(errors) > 0 {
			return diag.Errorf("failed to pack webhook %q", errors)
		}

		return nil
	}

	var readWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		webhook := CustomBaseParams{}

		webhook.EventFilterAPIModel.Criteria = domainCriteriaLookup[webhookType]

		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetResult(&webhook).
			SetError(&artifactoryError).
			Get(WebhookURL)

		if err != nil {
			return diag.FromErr(err)
		}

		if resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		if resp.IsError() {
			return diag.Errorf("%s", artifactoryError.String())
		}

		return packWebhook(data, webhook)
	}

	var retryOnProxyError = func(response *resty.Response, _r error) bool {
		var proxyNotFoundRegex = regexp.MustCompile("proxy with key '.*' not found")

		return proxyNotFoundRegex.MatchString(string(response.Body()[:]))
	}

	var createWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		webhook, err := unpackWebhook(data)
		if err != nil {
			return diag.FromErr(err)
		}

		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetBody(webhook).
			SetError(&artifactoryError).
			AddRetryCondition(retryOnProxyError).
			Post(webhooksURL)
		if err != nil {
			return diag.FromErr(err)
		}

		if resp.IsError() {
			return diag.Errorf("%s", artifactoryError.String())
		}

		data.SetId(webhook.Id())

		return readWebhook(ctx, data, m)
	}

	var updateWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		webhook, err := unpackWebhook(data)
		if err != nil {
			return diag.FromErr(err)
		}

		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetBody(webhook).
			SetError(&artifactoryError).
			AddRetryCondition(retryOnProxyError).
			Put(WebhookURL)
		if err != nil {
			return diag.FromErr(err)
		}

		if resp.IsError() {
			return diag.Errorf("%s", artifactoryError.String())
		}

		data.SetId(webhook.Id())

		return readWebhook(ctx, data, m)
	}

	var deleteWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetError(&artifactoryError).
			Delete(WebhookURL)

		if err != nil {
			return diag.FromErr(err)
		}

		if resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		if resp.IsError() {
			return diag.Errorf("%s", artifactoryError.String())
		}

		return nil
	}

	var eventTypesDiff = func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
		eventTypes := diff.Get("event_types").(*schema.Set).List()
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

	var criteriaDiff = func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
		if resource, ok := diff.GetOk("criteria"); ok {
			criteria := resource.(*schema.Set).List()
			if len(criteria) == 0 {
				return nil
			}
			return domainCriteriaValidationLookup[webhookType](ctx, criteria[0].(map[string]interface{}))
		}

		return nil
	}

	rs := schema.Resource{
		SchemaVersion: 2,
		CreateContext: createWebhook,
		ReadContext:   readWebhook,
		UpdateContext: updateWebhook,
		DeleteContext: deleteWebhook,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: domainSchemaLookup(currentSchemaVersion, true, webhookType)[webhookType],

		CustomizeDiff: customdiff.All(
			eventTypesDiff,
			criteriaDiff,
		),
		Description: "Provides an Artifactory webhook resource",
	}

	if webhookType == "artifactory_release_bundle" {
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

var domainCriteriaLookup = map[string]interface{}{}

var domainPackLookup = map[string]func(map[string]interface{}) map[string]interface{}{}

var domainUnpackLookup = map[string]func(map[string]interface{}, BaseCriteriaAPIModel) interface{}{}

var domainSchemaLookup = func(version int, isCustom bool, webhookType string) map[string]map[string]*schema.Schema {
	return map[string]map[string]*schema.Schema{}
}

var unpackCriteria = func(d *utilsdk.ResourceData, webhookType string) interface{} {
	var webhookCriteria interface{}

	if v, ok := d.GetOk("criteria"); ok {
		criteria := v.(*schema.Set).List()
		if len(criteria) == 1 {
			id := criteria[0].(map[string]interface{})

			baseCriteria := BaseCriteriaAPIModel{
				IncludePatterns: utilsdk.CastToStringArr(id["include_patterns"].(*schema.Set).List()),
				ExcludePatterns: utilsdk.CastToStringArr(id["exclude_patterns"].(*schema.Set).List()),
			}

			webhookCriteria = domainUnpackLookup[webhookType](id, baseCriteria)
		}
	}

	return webhookCriteria
}

var packCriteria = func(d *schema.ResourceData, webhookType string, criteria map[string]interface{}) []error {
	setValue := utilsdk.MkLens(d)

	resource := domainSchemaLookup(currentSchemaVersion, false, webhookType)[webhookType]["criteria"].Elem.(*schema.Resource)
	packedCriteria := domainPackLookup[webhookType](criteria)

	includePatterns := []interface{}{}
	if v, ok := criteria["includePatterns"]; ok && v != nil {
		includePatterns = v.([]interface{})
	}
	packedCriteria["include_patterns"] = schema.NewSet(schema.HashString, includePatterns)

	excludePatterns := []interface{}{}
	if v, ok := criteria["excludePatterns"]; ok && v != nil {
		excludePatterns = v.([]interface{})
	}
	packedCriteria["exclude_patterns"] = schema.NewSet(schema.HashString, excludePatterns)

	return setValue("criteria", schema.NewSet(schema.HashResource(resource), []interface{}{packedCriteria}))
}

var domainCriteriaValidationLookup = map[string]func(context.Context, map[string]interface{}) error{
	UserDomain: emptyCriteriaValidation,
}

var emptyCriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
	return nil
}

var packSecret = func(d *schema.ResourceData, url string) string {
	// Get secret from TF state
	var secret string
	if v, ok := d.GetOk("handler"); ok {
		handlers := v.(*schema.Set).List()
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

func ResourceArtifactoryWebhook(webhookType string) *schema.Resource {

	var unpackWebhook = func(data *schema.ResourceData) (WebhookAPIModel, error) {
		d := &utilsdk.ResourceData{ResourceData: data}

		var unpackHandlers = func(d *utilsdk.ResourceData) []HandlerAPIModel {
			var webhookHandlers []HandlerAPIModel

			if v, ok := d.GetOk("handler"); ok {
				handlers := v.(*schema.Set).List()
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

	var packHandlers = func(d *schema.ResourceData, handlers []HandlerAPIModel) []error {
		setValue := utilsdk.MkLens(d)
		resource := domainSchemaLookup(currentSchemaVersion, false, webhookType)[webhookType]["handler"].Elem.(*schema.Resource)
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

		return setValue("handler", schema.NewSet(schema.HashResource(resource), packedHandlers))
	}

	var packWebhook = func(d *schema.ResourceData, webhook WebhookAPIModel) diag.Diagnostics {
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
			return diag.Errorf("failed to pack webhook %q", errors)
		}

		return nil
	}

	var readWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		webhook := WebhookAPIModel{}

		webhook.EventFilter.Criteria = domainCriteriaLookup[webhookType]

		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetResult(&webhook).
			SetError(&artifactoryError).
			Get(WebhookURL)

		if err != nil {
			return diag.FromErr(err)
		}

		if resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		if resp.IsError() {
			return diag.Errorf("%s", artifactoryError.String())
		}

		return packWebhook(data, webhook)
	}

	var createWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		webhook, err := unpackWebhook(data)
		if err != nil {
			return diag.FromErr(err)
		}

		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetBody(webhook).
			AddRetryCondition(retryOnProxyError).
			SetError(&artifactoryError).
			Post(webhooksURL)
		if err != nil {
			return diag.FromErr(err)
		}

		if resp.IsError() {
			return diag.Errorf("%s", artifactoryError.String())
		}

		data.SetId(webhook.Id())

		return readWebhook(ctx, data, m)
	}

	var updateWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		webhook, err := unpackWebhook(data)
		if err != nil {
			return diag.FromErr(err)
		}

		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetBody(webhook).
			AddRetryCondition(retryOnProxyError).
			SetError(&artifactoryError).
			Put(WebhookURL)
		if err != nil {
			return diag.FromErr(err)
		}

		if resp.IsError() {
			return diag.Errorf("%s", artifactoryError.String())
		}

		data.SetId(webhook.Id())

		return readWebhook(ctx, data, m)
	}

	var deleteWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		var artifactoryError artifactory.ArtifactoryErrorsResponse
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetError(&artifactoryError).
			Delete(WebhookURL)

		if err != nil {
			return diag.FromErr(err)
		}

		if resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		if resp.IsError() {
			return diag.Errorf("%s", artifactoryError.String())
		}

		return nil
	}

	var eventTypesDiff = func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
		eventTypes := diff.Get("event_types").(*schema.Set).List()
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

	var criteriaDiff = func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
		if resource, ok := diff.GetOk("criteria"); ok {
			criteria := resource.(*schema.Set).List()
			if len(criteria) == 0 {
				return nil
			}
			return domainCriteriaValidationLookup[webhookType](ctx, criteria[0].(map[string]interface{}))
		}

		return nil
	}

	// Previous version of the schema
	// see example in https://www.terraform.io/plugin/sdkv2/resources/state-migration#terraform-v0-12-sdk-state-migrations
	resourceSchemaV1 := &schema.Resource{
		Schema: domainSchemaLookup(1, false, webhookType)[webhookType],
	}

	rs := schema.Resource{
		SchemaVersion: 2,
		CreateContext: createWebhook,
		ReadContext:   readWebhook,
		UpdateContext: updateWebhook,
		DeleteContext: deleteWebhook,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: domainSchemaLookup(currentSchemaVersion, false, webhookType)[webhookType],
		StateUpgraders: []schema.StateUpgrader{
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
