package webhook

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
	Key         string          `json:"key"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	EventFilter EventFilter     `json:"event_filter"`
	Handlers    []CustomHandler `json:"handlers"`
}

func (w CustomBaseParams) Id() string {
	return w.Key
}

type CustomHandler struct {
	HandlerType string         `json:"handler_type"`
	Url         string         `json:"url"`
	Secrets     []KeyValuePair `json:"secrets"`
	Proxy       string         `json:"proxy"`
	HttpHeaders []KeyValuePair `json:"http_headers"`
	Payload     string         `json:"payload,omitempty"`
}

type SecretName struct {
	Name string `json:"name"`
}

var packSecretsCustom = func(keyValuePairs []KeyValuePair, d *schema.ResourceData, url string) map[string]interface{} {
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
							Secrets:     unpackKeyValuePair(h["secrets"].(map[string]interface{})),
							Proxy:       h["proxy"].(string),
							HttpHeaders: unpackKeyValuePair(h["http_headers"].(map[string]interface{})),
							Payload:     h["payload"].(string),
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
			EventFilter: EventFilter{
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
				"url":          handler.Url,
				"secrets":      packSecretsCustom(handler.Secrets, d, handler.Url),
				"proxy":        handler.Proxy,
				"http_headers": packKeyValuePair(handler.HttpHeaders),
				"payload":      handler.Payload,
			}
			packedHandlers = append(packedHandlers, packedHandler)
		}

		return setValue("handler", schema.NewSet(schema.HashResource(resource), packedHandlers))
	}

	var packWebhook = func(d *schema.ResourceData, webhook CustomBaseParams) diag.Diagnostics {
		setValue := utilsdk.MkLens(d)

		var errors []error

		setValue("key", webhook.Key)
		setValue("description", webhook.Description)
		setValue("enabled", webhook.Enabled)
		setValue("event_types", webhook.EventFilter.EventTypes)
		errors = packCriteria(d, webhookType, webhook.EventFilter.Criteria.(map[string]interface{}))
		errors = packHandlers(d, webhook.Handlers)

		if len(errors) > 0 {
			return diag.Errorf("failed to pack webhook %q", errors)
		}

		return nil
	}

	var readWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "tflog.Debug(ctx, \"readWebhook\")")

		webhook := CustomBaseParams{}

		webhook.EventFilter.Criteria = domainCriteriaLookup[webhookType]

		_, err := m.(utilsdk.ProvderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetResult(&webhook).
			Get(WhUrl)

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

		_, err = m.(utilsdk.ProvderMetadata).Client.R().
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

		_, err = m.(utilsdk.ProvderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			SetBody(webhook).
			AddRetryCondition(retryOnProxyError).
			Put(WhUrl)
		if err != nil {
			return diag.FromErr(err)
		}

		data.SetId(webhook.Id())

		return readWebhook(ctx, data, m)
	}

	var deleteWebhook = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "deleteWebhook")

		resp, err := m.(utilsdk.ProvderMetadata).Client.R().
			SetPathParam("webhookKey", data.Id()).
			Delete(WhUrl)

		if err != nil && resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return diag.FromErr(err)
		}

		return nil
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

	return &schema.Resource{
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
}
