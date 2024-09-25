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
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"

	"golang.org/x/exp/slices"
)

const (
	webhooksURL = "/event/api/v1/subscriptions"
	WebhookURL  = "/event/api/v1/subscriptions/{webhookKey}"

	ArtifactLifecycleType        = "artifact_lifecycle"
	ArtifactPropertyType         = "artifact_property"
	ArtifactType                 = "artifact"
	ArtifactoryReleaseBundleType = "artifactory_release_bundle"
	BuildType                    = "build"
	DestinationType              = "destination"
	DistributionType             = "distribution"
	DockerType                   = "docker"
	ReleaseBundleType            = "release_bundle"
	ReleaseBundleV2Type          = "release_bundle_v2"
	ReleaseBundleV2PromotionType = "release_bundle_v2_promotion"
	UserType                     = "user"
)

const currentSchemaVersion = 2

var TypesSupported = []string{
	ArtifactLifecycleType,
	ArtifactPropertyType,
	// ArtifactType,
	ArtifactoryReleaseBundleType,
	BuildType,
	DestinationType,
	DistributionType,
	DockerType,
	ReleaseBundleType,
	ReleaseBundleV2Type,
	ReleaseBundleV2PromotionType,
	UserType,
}

var DomainEventTypesSupported = map[string][]string{
	ArtifactType:                 {"deployed", "deleted", "moved", "copied", "cached"},
	ArtifactPropertyType:         {"added", "deleted"},
	DockerType:                   {"pushed", "deleted", "promoted"},
	BuildType:                    {"uploaded", "deleted", "promoted"},
	ReleaseBundleType:            {"created", "signed", "deleted"},
	DistributionType:             {"distribute_started", "distribute_completed", "distribute_aborted", "distribute_failed", "delete_started", "delete_completed", "delete_failed"},
	ArtifactoryReleaseBundleType: {"received", "delete_started", "delete_completed", "delete_failed"},
	DestinationType:              {"received", "delete_started", "delete_completed", "delete_failed"},
	UserType:                     {"locked"},
	ReleaseBundleV2Type:          {"release_bundle_v2_started", "release_bundle_v2_failed", "release_bundle_v2_completed"},
	ReleaseBundleV2PromotionType: {"release_bundle_v2_promotion_completed", "release_bundle_v2_promotion_failed", "release_bundle_v2_promotion_started"},
	ArtifactLifecycleType:        {"archive", "restore"},
}

type BaseParams struct {
	Key         string      `json:"key"`
	Description string      `json:"description"`
	Enabled     bool        `json:"enabled"`
	EventFilter EventFilter `json:"event_filter"`
	Handlers    []Handler   `json:"handlers"`
}

func (w BaseParams) Id() string {
	return w.Key
}

type EventFilter struct {
	Domain     string      `json:"domain"`
	EventTypes []string    `json:"event_types"`
	Criteria   interface{} `json:"criteria"`
}

type Handler struct {
	HandlerType         string         `json:"handler_type"`
	Url                 string         `json:"url"`
	Secret              string         `json:"secret"`
	UseSecretForSigning bool           `json:"use_secret_for_signing"`
	Proxy               string         `json:"proxy"`
	CustomHttpHeaders   []KeyValuePair `json:"custom_http_headers"`
}

type KeyValuePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var unpackKeyValuePair = func(keyValuePairs map[string]interface{}) []KeyValuePair {
	var kvPairs []KeyValuePair
	for key, value := range keyValuePairs {
		keyValuePair := KeyValuePair{
			Name:  key,
			Value: value.(string),
		}
		kvPairs = append(kvPairs, keyValuePair)
	}

	return kvPairs
}

var packKeyValuePair = func(keyValuePairs []KeyValuePair) map[string]interface{} {
	kvPairs := make(map[string]interface{})
	for _, keyValuePair := range keyValuePairs {
		kvPairs[keyValuePair.Name] = keyValuePair.Value
	}

	return kvPairs
}

var domainCriteriaLookup = map[string]interface{}{
	ArtifactType:                 RepoWebhookCriteria{},
	ArtifactPropertyType:         RepoWebhookCriteria{},
	DockerType:                   RepoWebhookCriteria{},
	BuildType:                    BuildWebhookCriteria{},
	ReleaseBundleType:            ReleaseBundleWebhookCriteria{},
	DistributionType:             ReleaseBundleWebhookCriteria{},
	ArtifactoryReleaseBundleType: ReleaseBundleWebhookCriteria{},
	DestinationType:              ReleaseBundleWebhookCriteria{},
	UserType:                     EmptyWebhookCriteria{},
	ReleaseBundleV2Type:          ReleaseBundleV2WebhookCriteria{},
	ReleaseBundleV2PromotionType: ReleaseBundleV2PromotionWebhookCriteria{},
	ArtifactLifecycleType:        EmptyWebhookCriteria{},
}

var domainPackLookup = map[string]func(map[string]interface{}) map[string]interface{}{
	ArtifactType:                 packRepoCriteria,
	ArtifactPropertyType:         packRepoCriteria,
	DockerType:                   packRepoCriteria,
	BuildType:                    packBuildCriteria,
	ReleaseBundleType:            packReleaseBundleCriteria,
	DistributionType:             packReleaseBundleCriteria,
	ArtifactoryReleaseBundleType: packReleaseBundleCriteria,
	DestinationType:              packReleaseBundleCriteria,
	UserType:                     packEmptyCriteria,
	ReleaseBundleV2Type:          packReleaseBundleV2Criteria,
	ReleaseBundleV2PromotionType: packReleaseBundleV2PromotionCriteria,
	ArtifactLifecycleType:        packEmptyCriteria,
}

var domainUnpackLookup = map[string]func(map[string]interface{}, BaseWebhookCriteria) interface{}{
	ArtifactType:                 unpackRepoCriteria,
	ArtifactPropertyType:         unpackRepoCriteria,
	DockerType:                   unpackRepoCriteria,
	BuildType:                    unpackBuildCriteria,
	ReleaseBundleType:            unpackReleaseBundleCriteria,
	DistributionType:             unpackReleaseBundleCriteria,
	ArtifactoryReleaseBundleType: unpackReleaseBundleCriteria,
	DestinationType:              unpackReleaseBundleCriteria,
	UserType:                     unpackEmptyCriteria,
	ReleaseBundleV2Type:          unpackReleaseBundleV2Criteria,
	ReleaseBundleV2PromotionType: unpackReleaseBundleV2PromotionCriteria,
	ArtifactLifecycleType:        unpackEmptyCriteria,
}

var domainSchemaLookup = func(version int, isCustom bool, webhookType string) map[string]map[string]*schema.Schema {
	return map[string]map[string]*schema.Schema{
		ArtifactType:                 repoWebhookSchema(webhookType, version, isCustom),
		ArtifactPropertyType:         repoWebhookSchema(webhookType, version, isCustom),
		DockerType:                   repoWebhookSchema(webhookType, version, isCustom),
		BuildType:                    buildWebhookSchema(webhookType, version, isCustom),
		ReleaseBundleType:            releaseBundleWebhookSchema(webhookType, version, isCustom),
		DistributionType:             releaseBundleWebhookSchema(webhookType, version, isCustom),
		ArtifactoryReleaseBundleType: releaseBundleWebhookSchema(webhookType, version, isCustom),
		DestinationType:              releaseBundleWebhookSchema(webhookType, version, isCustom),
		UserType:                     userWebhookSchema(webhookType, version, isCustom),
		ReleaseBundleV2Type:          releaseBundleV2WebhookSchema(webhookType, version, isCustom),
		ReleaseBundleV2PromotionType: releaseBundleV2PromotionWebhookSchema(webhookType, version, isCustom),
		ArtifactLifecycleType:        artifactLifecycleWebhookSchema(webhookType, version, isCustom),
	}
}

var unpackCriteria = func(d *utilsdk.ResourceData, webhookType string) interface{} {
	var webhookCriteria interface{}

	if v, ok := d.GetOk("criteria"); ok {
		criteria := v.(*schema.Set).List()
		if len(criteria) == 1 {
			id := criteria[0].(map[string]interface{})

			baseCriteria := BaseWebhookCriteria{
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
	ArtifactType:                 repoCriteriaValidation,
	ArtifactPropertyType:         repoCriteriaValidation,
	DockerType:                   repoCriteriaValidation,
	BuildType:                    buildCriteriaValidation,
	ReleaseBundleType:            releaseBundleCriteriaValidation,
	DistributionType:             releaseBundleCriteriaValidation,
	ArtifactoryReleaseBundleType: releaseBundleCriteriaValidation,
	DestinationType:              releaseBundleCriteriaValidation,
	UserType:                     emptyCriteriaValidation,
	ReleaseBundleV2Type:          releaseBundleV2CriteriaValidation,
	ReleaseBundleV2PromotionType: emptyCriteriaValidation,
	ArtifactLifecycleType:        emptyCriteriaValidation,
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

	var unpackWebhook = func(data *schema.ResourceData) (BaseParams, error) {
		d := &utilsdk.ResourceData{ResourceData: data}

		var unpackHandlers = func(d *utilsdk.ResourceData) []Handler {
			var webhookHandlers []Handler

			if v, ok := d.GetOk("handler"); ok {
				handlers := v.(*schema.Set).List()
				for _, handler := range handlers {
					h := handler.(map[string]interface{})
					// use this to filter out weirdness with terraform adding an extra blank webhook in a set
					// https://discuss.hashicorp.com/t/using-typeset-in-provider-always-adds-an-empty-element-on-update/18566/2
					if h["url"].(string) != "" {
						webhookHandler := Handler{
							HandlerType: "webhook",
							Url:         h["url"].(string),
						}

						if v, ok := h["secret"]; ok {
							webhookHandler.Secret = v.(string)
						}

						if v, ok := h["use_secret_for_signing"]; ok {
							webhookHandler.UseSecretForSigning = v.(bool)
						}

						if v, ok := h["proxy"]; ok {
							webhookHandler.Proxy = v.(string)
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

		webhook := BaseParams{
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

	var packHandlers = func(d *schema.ResourceData, handlers []Handler) []error {
		setValue := utilsdk.MkLens(d)
		resource := domainSchemaLookup(currentSchemaVersion, false, webhookType)[webhookType]["handler"].Elem.(*schema.Resource)
		var packedHandlers []interface{}
		for _, handler := range handlers {
			packedHandler := map[string]interface{}{
				"url":                    handler.Url,
				"secret":                 packSecret(d, handler.Url),
				"use_secret_for_signing": handler.UseSecretForSigning,
				"proxy":                  handler.Proxy,
			}

			if handler.CustomHttpHeaders != nil {
				packedHandler["custom_http_headers"] = packKeyValuePair(handler.CustomHttpHeaders)
			}

			packedHandlers = append(packedHandlers, packedHandler)
		}

		return setValue("handler", schema.NewSet(schema.HashResource(resource), packedHandlers))
	}

	var packWebhook = func(d *schema.ResourceData, webhook BaseParams) diag.Diagnostics {
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
		tflog.Debug(ctx, "tflog.Debug(ctx, \"readWebhook\")")

		webhook := BaseParams{}

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
		tflog.Debug(ctx, "updateWebhook")

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
		tflog.Debug(ctx, "deleteWebhook")

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
