package artifactory

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

var webhookTypesSupported = []string{
	"artifact",
	"artifact_property",
	"docker",
	"build",
	"release_bundle",
	"distribution",
	"artifactory_release_bundle",
}

var domainEventTypesSupported = map[string][]string{
	"artifact": []string{"deployed", "deleted", "moved", "copied"},
	"artifact_property": []string{"added", "deleted"},
	"docker": []string{"pushed", "deleted", "promoted"},
	"build": []string{"uploaded", "deleted", "promoted"},
	"release_bundle": []string{"created", "signed", "deleted"},
	"distribution": []string{"distribute_started", "distribute_completed", "distribute_aborted", "distribute_failed", "delete_started", "delete_completed", "delete_failed"},
	"artifactory_release_bundle": []string{"received", "delete_started", "delete_completed", "delete_failed"},
}

type WebhookBaseParams struct {
	Key         string             `json:"key"`
	Description string             `json:"description"`
	Enabled     bool               `json:"enabled"`
	EventFilter WebhookEventFilter `json:"event_filter"`
	Handlers    []WebhookHandler   `json:"handlers"`
}

func (w WebhookBaseParams) Id() string {
	return w.Key
}

type WebhookEventFilter struct {
	Domain     string      `json:"domain"`
	EventTypes []string    `json:"event_types"`
	Criteria   interface{} `json:"criteria"`
}

type WebhookHandler struct {
	HandlerType       string                    `json:"handler_type"`
	Url               string                    `json:"url"`
	Secret            string                    `json:"secret"`
	Proxy             string                    `json:"proxy"`
	CustomHttpHeaders []WebhookCustomHttpHeader `json:"custom_http_headers"`
}

type WebhookCustomHttpHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

const webhooksUrl = "/event/api/v1/subscriptions"

const webhookUrl = webhooksUrl + "/{webhookKey}"

func resourceArtifactoryWebhook(webhookType string) *schema.Resource {

	var domainCriteriaLookup = map[string]interface{}{
		"artifact":                   RepoWebhookCriteria{},
		"artifact_property":          RepoWebhookCriteria{},
		"docker":                     RepoWebhookCriteria{},
		"build":                      BuildWebhookCriteria{},
		"release_bundle":             ReleaseBundleWebhookCriteria{},
		"distribution":               ReleaseBundleWebhookCriteria{},
		"artifactory_release_bundle": ReleaseBundleWebhookCriteria{},
	}

	var domainSchemaLookup = map[string]map[string]*schema.Schema{
		"artifact":                   repoWebhookSchema,
		"artifact_property":          repoWebhookSchema,
		"docker":                     repoWebhookSchema,
		"build":                      buildWebhookSchema,
		"release_bundle":             releaseBundleWebhookSchema,
		"distribution":               releaseBundleWebhookSchema,
		"artifactory_release_bundle": releaseBundleWebhookSchema,
	}

	var domainPackLookup = map[string]func(map[string]interface{}) map[string]interface{}{
		"artifact":                   packRepoCriteria,
		"artifact_property":          packRepoCriteria,
		"docker":                     packRepoCriteria,
		"build":                      packBuildCriteria,
		"release_bundle":             packReleaseBundleCriteria,
		"distribution":               packReleaseBundleCriteria,
		"artifactory_release_bundle": packReleaseBundleCriteria,
	}

	var domainUnpackLookup = map[string]func(map[string]interface{}, BaseWebhookCriteria) interface{}{
		"artifact":                   unpackRepoCriteria,
		"artifact_property":          unpackRepoCriteria,
		"docker":                     unpackRepoCriteria,
		"build":                      unpackBuildCriteria,
		"release_bundle":             unpackReleaseBundleCriteria,
		"distribution":               unpackReleaseBundleCriteria,
		"artifactory_release_bundle": unpackReleaseBundleCriteria,
	}

	var unpackWebhook = func(data *schema.ResourceData) (WebhookBaseParams, error) {
		d := &ResourceData{data}

		var unpackCriteria = func(d *ResourceData, webhookType string) interface{} {
			var webhookCriteria interface{}

			if v, ok := d.GetOkExists("criteria"); ok {
				criteria := v.(*schema.Set).List()
				if len(criteria) == 1 {
					id := criteria[0].(map[string]interface{})

					baseCriteria := BaseWebhookCriteria{
						IncludePatterns: castToStringArr(id["include_patterns"].(*schema.Set).List()),
						ExcludePatterns: castToStringArr(id["exclude_patterns"].(*schema.Set).List()),
					}

					webhookCriteria = domainUnpackLookup[webhookType](id, baseCriteria)
				}
			}

			return webhookCriteria
		}

		var unpackCustomHttpHeaders = func(d *ResourceData) []WebhookCustomHttpHeader {
			var customHeaders []WebhookCustomHttpHeader

			if v, ok := d.GetOkExists("custom_http_headers"); ok {
				headers := v.(map[string]interface{})
				for key, value := range headers {
					customHeader := WebhookCustomHttpHeader{
						Name:  key,
						Value: value.(string),
					}

					customHeaders = append(customHeaders, customHeader)
				}
			}

			return customHeaders
		}

		webhook := WebhookBaseParams{
			Key:         d.getString("key", false),
			Description: d.getString("description", false),
			Enabled:     d.getBool("enabled", false),
			EventFilter: WebhookEventFilter{
				Domain:     webhookType,
				EventTypes: d.getSet("event_types"),
				Criteria:   unpackCriteria(d, webhookType),
			},
			Handlers: []WebhookHandler{
				WebhookHandler{
					HandlerType:       "webhook",
					Url:               d.getString("url", false),
					Secret:            d.getString("secret", false),
					Proxy:             d.getString("proxy", false),
					CustomHttpHeaders: unpackCustomHttpHeaders(d),
				},
			},
		}

		return webhook, nil
	}

	var packCriteria = func(d *schema.ResourceData, criteria map[string]interface{}) []error {
		setValue := mkLens(d)

		resource := domainSchemaLookup[webhookType]["criteria"].Elem.(*schema.Resource)
		packedCriteria := domainPackLookup[webhookType](criteria)

		packedCriteria["include_patterns"] = schema.NewSet(schema.HashString, criteria["includePatterns"].([]interface{}))
		packedCriteria["exclude_patterns"] = schema.NewSet(schema.HashString, criteria["excludePatterns"].([]interface{}))

		return setValue("criteria", schema.NewSet(schema.HashResource(resource), []interface{}{packedCriteria}))
	}

	var packCustomHeaders = func(d *schema.ResourceData, customHeaders []WebhookCustomHttpHeader) []error {
		setValue := mkLens(d)

		headers := make(map[string]interface{})
		for _, customHeader := range customHeaders {
			headers[customHeader.Name] = customHeader.Value
		}

		return setValue("custom_http_headers", headers)
	}

	var packWebhook = func(d *schema.ResourceData, webhook WebhookBaseParams) diag.Diagnostics {
		setValue := mkLens(d)

		var errors []error

		errors = append(errors, setValue("key", webhook.Key)...)
		errors = append(errors, setValue("description", webhook.Description)...)
		errors = append(errors, setValue("enabled", webhook.Enabled)...)
		errors = append(errors, setValue("event_types", webhook.EventFilter.EventTypes)...)

		log.Printf("webhook.EventFilter.Criteria %v", webhook.EventFilter.Criteria)
		errors = append(errors, packCriteria(d, webhook.EventFilter.Criteria.(map[string]interface{}))...)

		handler := webhook.Handlers[0]
		errors = append(errors, setValue("url", handler.Url)...)
		errors = append(errors, setValue("secret", handler.Secret)...)
		errors = append(errors, setValue("proxy", handler.Proxy)...)

		errors = append(errors, packCustomHeaders(d, handler.CustomHttpHeaders)...)

		if len(errors) > 0 {
			return diag.Errorf("failed to pack webhook %q", errors)
		}

		return nil
	}

	var readWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] readWebhook")

		webhook := WebhookBaseParams{}

		webhook.EventFilter.Criteria = domainCriteriaLookup[webhookType]

		_, err := m.(*resty.Client).R().
			SetPathParam("webhookKey", data.Id()).
			SetResult(&webhook).
			Get(webhookUrl)

		if err != nil {
			return diag.FromErr(err)
		}

		return packWebhook(data, webhook)
	}

	var retryOnProxyError = func(response *resty.Response, _r error) bool {
		var proxyNotFoundRegex = regexp.MustCompile("proxy with key '.*' not found")

		return proxyNotFoundRegex.MatchString(string(response.Body()[:]))
	}

	var createWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] createWebhook")

		webhook, err := unpackWebhook(data)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = m.(*resty.Client).R().
			SetBody(webhook).
			AddRetryCondition(retryOnProxyError).
			Post(webhooksUrl)
		if err != nil {
			return diag.FromErr(err)
		}

		data.SetId(webhook.Id())

		return readWebhook(ctx, data, m)
	}

	var updateWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] updateWebhook")

		webhook, err := unpackWebhook(data)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = m.(*resty.Client).R().
			SetPathParam("webhookKey", data.Id()).
			SetBody(webhook).
			AddRetryCondition(retryOnProxyError).
			Put(webhookUrl)
		if err != nil {
			return diag.FromErr(err)
		}

		data.SetId(webhook.Id())

		return readWebhook(ctx, data, m)
	}

	var deleteWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] deleteWebhook")

		resp, err := m.(*resty.Client).R().
			SetPathParam("webhookKey", data.Id()).
			Delete(webhookUrl)

		if err != nil && resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return diag.FromErr(err)
		}

		return nil
	}

	var domainCriteriaValidationLookup = map[string]func(map[string]interface{}) error{
		"artifact":                   repoCriteriaValidation,
		"artifact_property":          repoCriteriaValidation,
		"docker":                     repoCriteriaValidation,
		"build":                      buildCriteriaValidation,
		"release_bundle":             releaseBundleCriteriaValidation,
		"distribution":               releaseBundleCriteriaValidation,
		"artifactory_release_bundle": releaseBundleCriteriaValidation,
	}

	var eventTypesDiff = func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
		log.Print("[DEBUG] eventTypesDiff")

		eventTypes := diff.Get("event_types").(*schema.Set).List()
		if len(eventTypes) == 0 {
			return nil
		}

		eventTypesSupported := domainEventTypesSupported[webhookType]
		for _, eventType := range eventTypes {
			if !contains(eventTypesSupported, eventType.(string)) {
				return fmt.Errorf("event_type %s not supported for domain %s", eventType, webhookType)
			}
		}
		return nil
	}

	var criteriaDiff = func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
		log.Print("[DEBUG] criteriaDiff")

		criteria := diff.Get("criteria").(*schema.Set).List()
		if len(criteria) == 0 {
			return nil
		}

		return domainCriteriaValidationLookup[webhookType](criteria[0].(map[string]interface{}))
	}

	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: createWebhook,
		ReadContext:   readWebhook,
		UpdateContext: updateWebhook,
		DeleteContext: deleteWebhook,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        domainSchemaLookup[webhookType],
		CustomizeDiff: customdiff.All(
			eventTypesDiff,
			criteriaDiff,
		),
		Description:   "Provides an Artifactory webhook resource",
	}
}
