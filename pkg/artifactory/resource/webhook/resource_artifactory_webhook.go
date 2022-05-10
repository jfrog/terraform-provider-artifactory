package webhook

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/util"
	"golang.org/x/exp/slices"
)

var WebhookTypesSupported = []string{
	"artifact",
	"artifact_property",
	"docker",
	"build",
	"release_bundle",
	"distribution",
	"artifactory_release_bundle",
}

var DomainEventTypesSupported = map[string][]string{
	"artifact":                   []string{"deployed", "deleted", "moved", "copied", "cached"},
	"artifact_property":          []string{"added", "deleted"},
	"docker":                     []string{"pushed", "deleted", "promoted"},
	"build":                      []string{"uploaded", "deleted", "promoted"},
	"release_bundle":             []string{"created", "signed", "deleted"},
	"distribution":               []string{"distribute_started", "distribute_completed", "distribute_aborted", "distribute_failed", "delete_started", "delete_completed", "delete_failed"},
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

const WebhookUrl = webhooksUrl + "/{webhookKey}"

const currentSchemaVersion = 2

func ResourceArtifactoryWebhook(webhookType string) *schema.Resource {

	var domainCriteriaLookup = map[string]interface{}{
		"artifact":                   RepoWebhookCriteria{},
		"artifact_property":          RepoWebhookCriteria{},
		"docker":                     RepoWebhookCriteria{},
		"build":                      BuildWebhookCriteria{},
		"release_bundle":             ReleaseBundleWebhookCriteria{},
		"distribution":               ReleaseBundleWebhookCriteria{},
		"artifactory_release_bundle": ReleaseBundleWebhookCriteria{},
	}

	var domainSchemaLookup = func(version int) map[string]map[string]*schema.Schema {
		return map[string]map[string]*schema.Schema{
			"artifact":                   repoWebhookSchema(webhookType, version),
			"artifact_property":          repoWebhookSchema(webhookType, version),
			"docker":                     repoWebhookSchema(webhookType, version),
			"build":                      buildWebhookSchema(webhookType, version),
			"release_bundle":             releaseBundleWebhookSchema(webhookType, version),
			"distribution":               releaseBundleWebhookSchema(webhookType, version),
			"artifactory_release_bundle": releaseBundleWebhookSchema(webhookType, version),
		}
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
		d := &util.ResourceData{data}

		var unpackCriteria = func(d *util.ResourceData, webhookType string) interface{} {
			var webhookCriteria interface{}

			if v, ok := d.GetOkExists("criteria"); ok {
				criteria := v.(*schema.Set).List()
				if len(criteria) == 1 {
					id := criteria[0].(map[string]interface{})

					baseCriteria := BaseWebhookCriteria{
						IncludePatterns: util.CastToStringArr(id["include_patterns"].(*schema.Set).List()),
						ExcludePatterns: util.CastToStringArr(id["exclude_patterns"].(*schema.Set).List()),
					}

					webhookCriteria = domainUnpackLookup[webhookType](id, baseCriteria)
				}
			}

			return webhookCriteria
		}

		var unpackCustomHttpHeaders = func(customHeaders map[string]interface{}) []WebhookCustomHttpHeader {
			var headers []WebhookCustomHttpHeader

			for key, value := range customHeaders {
				customHeader := WebhookCustomHttpHeader{
					Name:  key,
					Value: value.(string),
				}

				headers = append(headers, customHeader)
			}

			return headers
		}

		var unpackHandlers = func(d *util.ResourceData) []WebhookHandler {
			var webhookHandlers []WebhookHandler

			if v, ok := d.GetOkExists("handler"); ok {
				handlers := v.(*schema.Set).List()
				for _, handler := range handlers {
					h := handler.(map[string]interface{})

					webhookHandler := WebhookHandler{
						HandlerType:       "webhook",
						Url:               h["url"].(string),
						Secret:            h["secret"].(string),
						Proxy:             h["proxy"].(string),
						CustomHttpHeaders: unpackCustomHttpHeaders(h["custom_http_headers"].(map[string]interface{})),
					}
					webhookHandlers = append(webhookHandlers, webhookHandler)
				}
			}

			return webhookHandlers
		}

		webhook := WebhookBaseParams{
			Key:         d.GetString("key", false),
			Description: d.GetString("description", false),
			Enabled:     d.GetBool("enabled", false),
			EventFilter: WebhookEventFilter{
				Domain:     webhookType,
				EventTypes: d.GetSet("event_types"),
				Criteria:   unpackCriteria(d, webhookType),
			},
			Handlers: unpackHandlers(d),
		}

		return webhook, nil
	}

	var packCriteria = func(d *schema.ResourceData, criteria map[string]interface{}) []error {
		setValue := util.MkLens(d)

		resource := domainSchemaLookup(currentSchemaVersion)[webhookType]["criteria"].Elem.(*schema.Resource)
		packedCriteria := domainPackLookup[webhookType](criteria)

		packedCriteria["include_patterns"] = schema.NewSet(schema.HashString, criteria["includePatterns"].([]interface{}))
		packedCriteria["exclude_patterns"] = schema.NewSet(schema.HashString, criteria["excludePatterns"].([]interface{}))

		return setValue("criteria", schema.NewSet(schema.HashResource(resource), []interface{}{packedCriteria}))
	}

	var packCustomHeaders = func(customHeaders []WebhookCustomHttpHeader) map[string]interface{} {
		headers := make(map[string]interface{})
		for _, customHeader := range customHeaders {
			headers[customHeader.Name] = customHeader.Value
		}

		return headers
	}

	var packHandlers = func(d *schema.ResourceData, handlers []WebhookHandler) []error {
		setValue := util.MkLens(d)

		resource := domainSchemaLookup(currentSchemaVersion)[webhookType]["handler"].Elem.(*schema.Resource)

		packedHandlers := make([]interface{}, len(handlers))

		for _, handler := range handlers {
			packedHandler := map[string]interface{}{
				"url": handler.Url,
				"secret": handler.Secret,
				"proxy": handler.Proxy,
				"custom_http_headers": packCustomHeaders(handler.CustomHttpHeaders),
			}
			packedHandlers = append(packedHandlers, packedHandler)
		}

		return setValue("handler", schema.NewSet(schema.HashResource(resource), packedHandlers))
	}

	var packWebhook = func(d *schema.ResourceData, webhook WebhookBaseParams) diag.Diagnostics {
		setValue := util.MkLens(d)

		var errors []error

		errors = append(errors, setValue("key", webhook.Key)...)
		errors = append(errors, setValue("description", webhook.Description)...)
		errors = append(errors, setValue("enabled", webhook.Enabled)...)
		errors = append(errors, setValue("event_types", webhook.EventFilter.EventTypes)...)
		errors = append(errors, packCriteria(d, webhook.EventFilter.Criteria.(map[string]interface{}))...)
		errors = append(errors, packHandlers(d, webhook.Handlers)...)

		if len(errors) > 0 {
			return diag.Errorf("failed to pack webhook %q", errors)
		}

		return nil
	}

	var readWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "tflog.Debug(ctx, \"readWebhook\")")

		webhook := WebhookBaseParams{}

		webhook.EventFilter.Criteria = domainCriteriaLookup[webhookType]

		_, err := m.(*resty.Client).R().
			SetPathParam("webhookKey", data.Id()).
			SetResult(&webhook).
			Get(WebhookUrl)

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
		tflog.Debug(ctx, "createWebhook")

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
		tflog.Debug(ctx, "updateWebhook")

		webhook, err := unpackWebhook(data)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = m.(*resty.Client).R().
			SetPathParam("webhookKey", data.Id()).
			SetBody(webhook).
			AddRetryCondition(retryOnProxyError).
			Put(WebhookUrl)
		if err != nil {
			return diag.FromErr(err)
		}

		data.SetId(webhook.Id())

		return readWebhook(ctx, data, m)
	}

	var deleteWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "deleteWebhook")

		resp, err := m.(*resty.Client).R().
			SetPathParam("webhookKey", data.Id()).
			Delete(WebhookUrl)

		if err != nil && resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return diag.FromErr(err)
		}

		return nil
	}

	var domainCriteriaValidationLookup = map[string]func(context.Context, map[string]interface{}) error{
		"artifact":                   repoCriteriaValidation,
		"artifact_property":          repoCriteriaValidation,
		"docker":                     repoCriteriaValidation,
		"build":                      buildCriteriaValidation,
		"release_bundle":             releaseBundleCriteriaValidation,
		"distribution":               releaseBundleCriteriaValidation,
		"artifactory_release_bundle": releaseBundleCriteriaValidation,
	}

	var eventTypesDiff = func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
		tflog.Debug(ctx, "eventTypesDiff")

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
		tflog.Debug(ctx, "criteriaDiff")

		criteria := diff.Get("criteria").(*schema.Set).List()
		if len(criteria) == 0 {
			return nil
		}

		return domainCriteriaValidationLookup[webhookType](ctx, criteria[0].(map[string]interface{}))
	}

	// taken from example in https://www.terraform.io/plugin/sdkv2/resources/state-migration#terraform-v0-12-sdk-state-migrations
	var resourceSchemaV1 = func() *schema.Resource {
		return &schema.Resource{
			Schema: domainSchemaLookup(1)[webhookType],
		}
	}

	return &schema.Resource{
		SchemaVersion: 2,
		CreateContext: createWebhook,
		ReadContext:   readWebhook,
		UpdateContext: updateWebhook,
		DeleteContext: deleteWebhook,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:         domainSchemaLookup(currentSchemaVersion)[webhookType],
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceSchemaV1().CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceStateUpgradeV1,
				Version: 1,
			},
		},

		CustomizeDiff: customdiff.All(
			eventTypesDiff,
			criteriaDiff,
		),
		Description:   "Provides an Artifactory webhook resource",
	}
}

var ResourceStateUpgradeV1 = func(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	rawState["handler"] = []map[string]interface{}{
		{
			"url": rawState["url"],
			"secret": rawState["secret"],
			"proxy": rawState["proxy"],
			"custom_http_headers": rawState["custom_http_headers"],
		},
	}

	delete(rawState, "url")
	delete(rawState, "secret")
	delete(rawState, "proxy")
	delete(rawState, "custom_http_headers")

	return rawState, nil
}
