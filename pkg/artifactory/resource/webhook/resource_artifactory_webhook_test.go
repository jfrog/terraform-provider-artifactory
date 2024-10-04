package webhook_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/webhook"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

var domainValidationErrorMessageLookup = map[string]string{
	"artifact":                   `repo_keys cannot be empty when any_local, any_remote, and any_federated are\s*false`,
	"artifact_property":          `repo_keys cannot be empty when any_local, any_remote, and any_federated are\s*false`,
	"docker":                     `repo_keys cannot be empty when any_local, any_remote, and any_federated are\s*false`,
	"build":                      `selected_builds or include_patterns cannot be empty when any_build is false`,
	"release_bundle":             `registered_release_bundle_names cannot be empty when any_release_bundle is\s*false`,
	"distribution":               `registered_release_bundle_names cannot be empty when any_release_bundle is\s*false`,
	"artifactory_release_bundle": `registered_release_bundle_names cannot be empty when any_release_bundle is\s*false`,
}

var releaseBundleTemplate = `
	resource "artifactory_{{ .webhookType }}_webhook" "{{ .webhookName }}" {
		key         = "{{ .webhookName }}"
		description = "test description"
		event_types = [{{ range $index, $eventType := .eventTypes}}{{if $index}},{{end}}"{{$eventType}}"{{end}}]
		criteria {
			any_release_bundle = false
			registered_release_bundle_names = []
		}
		handler {
			url = "https://google.com"
		}
	}
`

var releaseBundleV2Template = `
	resource "artifactory_{{ .webhookType }}_webhook" "{{ .webhookName }}" {
		key         = "{{ .webhookName }}"
		description = "test description"
		event_types = [{{ range $index, $eventType := .eventTypes}}{{if $index}},{{end}}"{{$eventType}}"{{end}}]
		criteria {
			any_release_bundle = false
			selected_release_bundles = []
		}
		handler {
			url = "https://google.com"
		}
	}
`

func TestAccWebhook_CriteriaValidation(t *testing.T) {
	for _, webhookType := range []string{webhook.ArtifactDomain, webhook.ArtifactPropertyDomain, webhook.ArtifactoryReleaseBundleDomain, webhook.BuildDomain, webhook.DestinationDomain, webhook.DistributionDomain, webhook.DockerDomain, webhook.ReleaseBundleDomain, webhook.ReleaseBundleV2Domain} {
		t.Run(webhookType, func(t *testing.T) {
			resource.Test(webhookCriteriaValidationTestCase(webhookType, t))
		})
	}
}

func webhookCriteriaValidationTestCase(webhookType string, t *testing.T) (*testing.T, resource.TestCase) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_%s_webhook.%s", webhookType, name)

	var template string
	switch webhookType {
	case "artifact", "artifact_property", "docker":
		template = repoTemplate
	case "build":
		template = buildTemplate
	case "release_bundle", "distribution", "artifactory_release_bundle", "destination":
		template = releaseBundleTemplate
	case "release_bundle_v2":
		template = releaseBundleV2Template
	}

	params := map[string]interface{}{
		"webhookType": webhookType,
		"webhookName": name,
		"eventTypes":  webhook.DomainEventTypesSupported[webhookType],
	}
	webhookConfig := util.ExecuteTemplate("TestAccWebhookCriteriaValidation", template, params)

	return t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:      webhookConfig,
				ExpectError: regexp.MustCompile(domainValidationErrorMessageLookup[webhookType]),
			},
		},
	}
}

func TestAccWebhook_EventTypesValidation(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_artifact_webhook.%s", name)

	wrongEventType := "wrong-event-type"

	params := map[string]interface{}{
		"webhookName": name,
		"eventType":   wrongEventType,
	}
	webhookConfig := util.ExecuteTemplate("TestAccWebhookEventTypesValidation", `
		resource "artifactory_artifact_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["{{ .eventType }}"]
			criteria {
				any_local  = true
				any_remote = true
				any_federated = true
				repo_keys  = []
			}
			handler {
				url = "https://google.com"
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:      webhookConfig,
				ExpectError: regexp.MustCompile(fmt.Sprintf(`value must be one of:\s*\["deployed" "deleted" "moved" "copied" "cached"\], got: "%s"`, wrongEventType)),
			},
		},
	})
}

func TestAccWebhook_HandlerValidation_EmptyProxy(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_artifact_webhook.%s", name)

	params := map[string]interface{}{
		"webhookName": name,
	}
	webhookConfig := util.ExecuteTemplate("TestAccWebhookEventTypesValidation", `
		resource "artifactory_artifact_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["deployed"]
			criteria {
				any_local  = true
				any_remote = true
				any_federated = true
				repo_keys  = []
			}
			handler {
				url   = "https://google.com"
				proxy = ""
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:      webhookConfig,
				ExpectError: regexp.MustCompile(`proxy\s*string length must be at least 1, got: 0`),
			},
		},
	})
}

func TestAccWebhook_HandlerValidation_ProxyWithURL(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_artifact_webhook.%s", name)

	params := map[string]interface{}{
		"webhookName": name,
	}

	webhookConfig := util.ExecuteTemplate("TestAccWebhookEventTypesValidation", `
		resource "artifactory_artifact_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["deployed"]
			criteria {
				any_local  = true
				any_remote = true
				any_federated = true
				repo_keys  = []
			}
			handler {
				url   = "https://google.com"
				proxy = "https://google.com"
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:      webhookConfig,
				ExpectError: regexp.MustCompile(`.*expected "proxy" not to be a valid url.*`),
			},
		},
	})
}

func testCheckWebhook(id string, request *resty.Request) (*resty.Response, error) {
	return request.
		SetPathParam("webhookKey", id).
		AddRetryCondition(client.NeverRetry).
		Get(webhook.WebhookURL)
}

func TestAccWebhook_GH476WebHookChangeBearerSet0(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-webhook", "artifactory_artifact_webhook")

	temp := `
		resource "artifactory_artifact_webhook" "{{ .webhookName }}" {
		  key = "{{ .webhookName }}"
		
		  event_types = ["deployed"]
		
		  criteria {
			any_local  = true
			any_remote = false
			any_federated = false
		
			repo_keys = []
		  }
		
		  handler {
			custom_http_headers = {
			  "Authorization" = "Bearer {{ .token }}"
			}
		
			url = "https://example.com"
		  }
		}
	`
	firstToken := testutil.RandomInt()
	config1 := util.ExecuteTemplate(
		"TestAccWebhook{{ .webhookName }}",
		temp,
		map[string]interface{}{
			"webhookName": name,
			"token":       firstToken,
		},
	)
	secondToken := testutil.RandomInt()
	config2 := util.ExecuteTemplate(
		"TestAccWebhook{{ .webhookName }}",
		temp,
		map[string]interface{}{
			"webhookName": name,
			"token":       secondToken,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "key", testCheckWebhook),
		Steps: []resource.TestStep{
			{
				Config:           config1,
				Check:            resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.Authorization", fmt.Sprintf("Bearer %d", firstToken)),
				ConfigPlanChecks: testutil.ConfigPlanChecks(fqrn),
			},
			{
				Config: config2,
				Check:  resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.Authorization", fmt.Sprintf("Bearer %d", secondToken)),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}
