package webhook

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/validator"
)

type BaseWebhookCriteria struct {
	IncludePatterns []string `json:"includePatterns"`
	ExcludePatterns []string `json:"excludePatterns"`
}

var baseCriteriaSchema = map[string]*schema.Schema{
	"include_patterns": {
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: `Simple comma separated wildcard patterns for repository artifact paths (with no leading slash).\nAnt-style path expressions are supported (*, **, ?).\nFor example: "org/apache/**"`,
	},
	"exclude_patterns": {
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: `Simple comma separated wildcard patterns for repository artifact paths (with no leading slash).\nAnt-style path expressions are supported (*, **, ?).\nFor example: "org/apache/**"`,
	},
}

func getBaseSchemaByVersion(webhookType string, version int) map[string]*schema.Schema {
	if version == 1 {
		return baseWebhookBaseSchemaV1(webhookType)
	}
	return baseWebhookBaseSchemaV2(webhookType)
}

func baseWebhookBaseSchemaV1(webhookType string) map[string]*schema.Schema {
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
		"secret": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Secret authentication token that will be sent to the configured URL.",
		},
		"proxy": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Proxy key from Artifactory Proxies setting",
		},
		"custom_http_headers": {
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Custom HTTP headers you wish to use to invoke the Webhook, comprise of key/value pair.",
		},
	}
}

func baseWebhookBaseSchemaV2(webhookType string) map[string]*schema.Schema {
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
					"secret": {
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
						Description:      "Secret authentication token that will be sent to the configured URL. The value will be sent as `x-jfrog-event-auth` header.",
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
					"custom_http_headers": {
						Type:        schema.TypeMap,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Custom HTTP headers you wish to use to invoke the Webhook, comprise of key/value pair.",
					},
				},
			},
		},
	}
}
