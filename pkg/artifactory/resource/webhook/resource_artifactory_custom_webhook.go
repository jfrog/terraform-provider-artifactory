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

var DomainSupported = []string{
	ArtifactLifecycleDomain,
	ArtifactPropertyDomain,
	ArtifactDomain,
	ArtifactoryReleaseBundleDomain,
	BuildDomain,
	DestinationDomain,
	DistributionDomain,
	DockerDomain,
	ReleaseBundleDomain,
	ReleaseBundleV2Domain,
	ReleaseBundleV2PromotionDomain,
	UserDomain,
}

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

var repoWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(getBaseSchemaByVersion(webhookType, version, isCustom), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: utilsdk.MergeMaps(baseCriteriaSchema, map[string]*schema.Schema{
					"any_local": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any local repositories",
					},
					"any_remote": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any remote repositories",
					},
					"any_federated": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any federated repositories",
					},
					"repo_keys": {
						Type:        schema.TypeSet,
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Trigger on this list of repository keys",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied on which repositories.",
		},
	})
}

var buildWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(getBaseSchemaByVersion(webhookType, version, isCustom), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: utilsdk.MergeMaps(baseCriteriaSchema, map[string]*schema.Schema{
					"any_build": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any builds",
					},
					"selected_builds": {
						Type:        schema.TypeSet,
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Trigger on this list of build IDs",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied on which builds.",
		},
	})
}

var releaseBundleWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(getBaseSchemaByVersion(webhookType, version, isCustom), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: utilsdk.MergeMaps(baseCriteriaSchema, map[string]*schema.Schema{
					"any_release_bundle": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any release bundles or distributions",
					},
					"registered_release_bundle_names": {
						Type:        schema.TypeSet,
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Trigger on this list of release bundle names",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied, on which release bundles or distributions.",
		},
	})
}

var releaseBundleV2WebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(getBaseSchemaByVersion(webhookType, version, isCustom), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: utilsdk.MergeMaps(baseCriteriaSchema, map[string]*schema.Schema{
					"any_release_bundle": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any release bundles or distributions",
					},
					"selected_release_bundles": {
						Type:        schema.TypeSet,
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Trigger on this list of release bundle names",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied, on which release bundles or distributions.",
		},
	})
}

var releaseBundleV2PromotionWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(getBaseSchemaByVersion(webhookType, version, isCustom), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: utilsdk.MergeMaps(baseCriteriaSchema, map[string]*schema.Schema{
					"selected_environments": {
						Type:        schema.TypeSet,
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Trigger on this list of environments",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied, on which release bundles promotion.",
		},
	})
}

var userWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return getBaseSchemaByVersion(webhookType, version, isCustom)
}

var artifactLifecycleWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return getBaseSchemaByVersion(webhookType, version, isCustom)
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

type EmptyWebhookCriteria struct{}

var domainCriteriaLookup = map[string]interface{}{
	"artifact":                    RepoCriteriaAPIModel{},
	"artifact_property":           RepoCriteriaAPIModel{},
	"docker":                      RepoCriteriaAPIModel{},
	"build":                       BuildCriteriaAPIModel{},
	"release_bundle":              ReleaseBundleCriteriaAPIModel{},
	"distribution":                ReleaseBundleCriteriaAPIModel{},
	"artifactory_release_bundle":  ReleaseBundleCriteriaAPIModel{},
	"destination":                 ReleaseBundleCriteriaAPIModel{},
	"user":                        EmptyWebhookCriteria{},
	"release_bundle_v2":           ReleaseBundleV2CriteriaAPIModel{},
	"release_bundle_v2_promotion": ReleaseBundleV2PromotionCriteriaAPIModel{},
	"artifact_lifecycle":          EmptyWebhookCriteria{},
}

var domainPackLookup = map[string]func(map[string]interface{}) map[string]interface{}{
	"artifact":                    packRepoCriteria,
	"artifact_property":           packRepoCriteria,
	"docker":                      packRepoCriteria,
	"build":                       packBuildCriteria,
	"release_bundle":              packReleaseBundleCriteria,
	"distribution":                packReleaseBundleCriteria,
	"artifactory_release_bundle":  packReleaseBundleCriteria,
	"destination":                 packReleaseBundleCriteria,
	"user":                        packEmptyCriteria,
	"release_bundle_v2":           packReleaseBundleV2Criteria,
	"release_bundle_v2_promotion": packReleaseBundleV2PromotionCriteria,
	"artifact_lifecycle":          packEmptyCriteria,
}

var domainUnpackLookup = map[string]func(map[string]interface{}, BaseCriteriaAPIModel) interface{}{
	"artifact":                    unpackRepoCriteria,
	"artifact_property":           unpackRepoCriteria,
	"docker":                      unpackRepoCriteria,
	"build":                       unpackBuildCriteria,
	"release_bundle":              unpackReleaseBundleCriteria,
	"distribution":                unpackReleaseBundleCriteria,
	"artifactory_release_bundle":  unpackReleaseBundleCriteria,
	"destination":                 unpackReleaseBundleCriteria,
	"user":                        unpackEmptyCriteria,
	"release_bundle_v2":           unpackReleaseBundleV2Criteria,
	"release_bundle_v2_promotion": unpackReleaseBundleV2PromotionCriteria,
	"artifact_lifecycle":          unpackEmptyCriteria,
}

var domainSchemaLookup = func(version int, isCustom bool, webhookType string) map[string]map[string]*schema.Schema {
	return map[string]map[string]*schema.Schema{
		"artifact":                    repoWebhookSchema(webhookType, version, isCustom),
		"artifact_property":           repoWebhookSchema(webhookType, version, isCustom),
		"docker":                      repoWebhookSchema(webhookType, version, isCustom),
		"build":                       buildWebhookSchema(webhookType, version, isCustom),
		"release_bundle":              releaseBundleWebhookSchema(webhookType, version, isCustom),
		"distribution":                releaseBundleWebhookSchema(webhookType, version, isCustom),
		"artifactory_release_bundle":  releaseBundleWebhookSchema(webhookType, version, isCustom),
		"destination":                 releaseBundleWebhookSchema(webhookType, version, isCustom),
		"user":                        userWebhookSchema(webhookType, version, isCustom),
		"release_bundle_v2":           releaseBundleV2WebhookSchema(webhookType, version, isCustom),
		"release_bundle_v2_promotion": releaseBundleV2PromotionWebhookSchema(webhookType, version, isCustom),
		"artifact_lifecycle":          artifactLifecycleWebhookSchema(webhookType, version, isCustom),
	}
}

var packRepoCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	criteria := map[string]interface{}{
		"any_local":     artifactoryCriteria["anyLocal"].(bool),
		"any_remote":    artifactoryCriteria["anyRemote"].(bool),
		"any_federated": false,
		"repo_keys":     schema.NewSet(schema.HashString, artifactoryCriteria["repoKeys"].([]interface{})),
	}

	if v, ok := artifactoryCriteria["anyFederated"]; ok {
		criteria["any_federated"] = v.(bool)
	}

	return criteria
}

var packReleaseBundleCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"any_release_bundle":              artifactoryCriteria["anyReleaseBundle"].(bool),
		"registered_release_bundle_names": schema.NewSet(schema.HashString, artifactoryCriteria["registeredReleaseBundlesNames"].([]interface{})),
	}
}

var unpackReleaseBundleCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseCriteriaAPIModel) interface{} {
	return ReleaseBundleCriteriaAPIModel{
		AnyReleaseBundle:              terraformCriteria["any_release_bundle"].(bool),
		RegisteredReleaseBundlesNames: utilsdk.CastToStringArr(terraformCriteria["registered_release_bundle_names"].(*schema.Set).List()),
		BaseCriteriaAPIModel:          baseCriteria,
	}
}

var unpackRepoCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseCriteriaAPIModel) interface{} {
	return RepoCriteriaAPIModel{
		AnyLocal:             terraformCriteria["any_local"].(bool),
		AnyRemote:            terraformCriteria["any_remote"].(bool),
		AnyFederated:         terraformCriteria["any_federated"].(bool),
		RepoKeys:             utilsdk.CastToStringArr(terraformCriteria["repo_keys"].(*schema.Set).List()),
		BaseCriteriaAPIModel: baseCriteria,
	}
}

var packBuildCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"any_build":       artifactoryCriteria["anyBuild"].(bool),
		"selected_builds": schema.NewSet(schema.HashString, artifactoryCriteria["selectedBuilds"].([]interface{})),
	}
}

var unpackBuildCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseCriteriaAPIModel) interface{} {
	return BuildCriteriaAPIModel{
		AnyBuild:             terraformCriteria["any_build"].(bool),
		SelectedBuilds:       utilsdk.CastToStringArr(terraformCriteria["selected_builds"].(*schema.Set).List()),
		BaseCriteriaAPIModel: baseCriteria,
	}
}

var packReleaseBundleV2Criteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"any_release_bundle":       artifactoryCriteria["anyReleaseBundle"].(bool),
		"selected_release_bundles": schema.NewSet(schema.HashString, artifactoryCriteria["selectedReleaseBundles"].([]interface{})),
	}
}

var unpackReleaseBundleV2Criteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseCriteriaAPIModel) interface{} {
	return ReleaseBundleV2CriteriaAPIModel{
		AnyReleaseBundle:       terraformCriteria["any_release_bundle"].(bool),
		SelectedReleaseBundles: utilsdk.CastToStringArr(terraformCriteria["selected_release_bundles"].(*schema.Set).List()),
		BaseCriteriaAPIModel:   baseCriteria,
	}
}

var packReleaseBundleV2PromotionCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"selected_environments": schema.NewSet(schema.HashString, artifactoryCriteria["selectedEnvironments"].([]interface{})),
	}
}

var unpackReleaseBundleV2PromotionCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseCriteriaAPIModel) interface{} {
	return ReleaseBundleV2PromotionCriteriaAPIModel{
		SelectedEnvironments: utilsdk.CastToStringArr(terraformCriteria["selected_environments"].(*schema.Set).List()),
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
	"artifact":                    repoCriteriaValidation,
	"artifact_property":           repoCriteriaValidation,
	"docker":                      repoCriteriaValidation,
	"build":                       buildCriteriaValidation,
	"release_bundle":              releaseBundleCriteriaValidation,
	"distribution":                releaseBundleCriteriaValidation,
	"artifactory_release_bundle":  releaseBundleCriteriaValidation,
	"destination":                 releaseBundleCriteriaValidation,
	"user":                        emptyCriteriaValidation,
	"release_bundle_v2":           releaseBundleV2CriteriaValidation,
	"release_bundle_v2_promotion": emptyCriteriaValidation,
	"artifact_lifecycle":          emptyCriteriaValidation,
}

var repoCriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
	anyLocal := criteria["any_local"].(bool)
	anyRemote := criteria["any_remote"].(bool)
	anyFederated := criteria["any_federated"].(bool)
	repoKeys := criteria["repo_keys"].(*schema.Set).List()

	if (!anyLocal && !anyRemote && !anyFederated) && len(repoKeys) == 0 {
		return fmt.Errorf("repo_keys cannot be empty when any_local, any_remote, and any_federated are false")
	}

	return nil
}

var buildCriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
	anyBuild := criteria["any_build"].(bool)
	selectedBuilds := criteria["selected_builds"].(*schema.Set).List()
	includePatterns := criteria["include_patterns"].(*schema.Set).List()

	if !anyBuild && (len(selectedBuilds) == 0 && len(includePatterns) == 0) {
		return fmt.Errorf("selected_builds or include_patterns cannot be empty when any_build is false")
	}

	return nil
}

var releaseBundleCriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
	anyReleaseBundle := criteria["any_release_bundle"].(bool)
	registeredReleaseBundlesNames := criteria["registered_release_bundle_names"].(*schema.Set).List()

	if !anyReleaseBundle && len(registeredReleaseBundlesNames) == 0 {
		return fmt.Errorf("registered_release_bundle_names cannot be empty when any_release_bundle is false")
	}

	return nil
}

var releaseBundleV2CriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
	anyReleaseBundle := criteria["any_release_bundle"].(bool)
	selectedReleaseBundles := criteria["selected_release_bundles"].(*schema.Set).List()

	if !anyReleaseBundle && len(selectedReleaseBundles) == 0 {
		return fmt.Errorf("selected_release_bundles cannot be empty when any_release_bundle is false")
	}

	return nil
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

