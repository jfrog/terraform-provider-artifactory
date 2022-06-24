package webhook_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/webhook"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/test"
)

var domainRepoTypeLookup = map[string]string{
	"artifact":          "generic",
	"artifact_property": "generic",
	"docker":            "docker_v2",
}

var domainValidationErrorMessageLookup = map[string]string{
	"artifact":                   "repo_keys cannot be empty when both any_local and any_remote are false",
	"artifact_property":          "repo_keys cannot be empty when both any_local and any_remote are false",
	"docker":                     "repo_keys cannot be empty when both any_local and any_remote are false",
	"build":                      "selected_builds cannot be empty when any_build is false",
	"release_bundle":             "registered_release_bundle_names cannot be empty when any_release_bundle is false",
	"distribution":               "registered_release_bundle_names cannot be empty when any_release_bundle is false",
	"artifactory_release_bundle": "registered_release_bundle_names cannot be empty when any_release_bundle is false",
}

var repoTemplate = `
	resource "artifactory_{{ .webhookType }}_webhook" "{{ .webhookName }}" {
		key         = "{{ .webhookName }}"
		description = "test description"
		event_types = [{{ range $index, $eventType := .eventTypes}}{{if $index}},{{end}}"{{$eventType}}"{{end}}]
		criteria {
			any_local = false
			any_remote = false
			repo_keys = []
		}
		handler {
			url = "http://tempurl.org"
		}
	}
`

var buildTemplate = `
	resource "artifactory_{{ .webhookType }}_webhook" "{{ .webhookName }}" {
		key         = "{{ .webhookName }}"
		description = "test description"
		event_types = [{{ range $index, $eventType := .eventTypes}}{{if $index}},{{end}}"{{$eventType}}"{{end}}]
		criteria {
			any_build = false
			selected_builds = []
		}
		handler {
			url = "http://tempurl.org"
		}
	}
`

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
			url = "http://tempurl.org"
		}
	}
`

func testAccWebhookCriteriaValidation(t *testing.T) {
	for _, webhookType := range webhook.WebhookTypesSupported {
		t.Run(fmt.Sprintf("TestWebhook%sCriteriaValidation", strings.Title(strings.ToLower(webhookType))), func(t *testing.T) {
			resource.Test(webhookCriteriaValidationTestCase(webhookType, t))
		})
	}
}

func webhookCriteriaValidationTestCase(webhookType string, t *testing.T) (*testing.T, resource.TestCase) {
	id := test.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_%s_webhook.%s", webhookType, name)

	var template string
	switch webhookType {
	case "artifact", "artifact_property", "docker":
		template = repoTemplate
	case "build":
		template = buildTemplate
	case "release_bundle", "distribution", "artifactory_release_bundle":
		template = releaseBundleTemplate
	}

	params := map[string]interface{}{
		"webhookType": webhookType,
		"webhookName": name,
		"eventTypes":  webhook.DomainEventTypesSupported[webhookType],
	}
	webhookConfig := acctest.ExecuteTemplate("TestAccWebhookCriteriaValidation", template, params)

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config:      webhookConfig,
				ExpectError: regexp.MustCompile(domainValidationErrorMessageLookup[webhookType]),
			},
		},
	}
}

func testAccWebhookEventTypesValidation(t *testing.T) {
	id := test.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_artifact_webhook.%s", name)

	wrongEventType := "wrong-event-type"

	params := map[string]interface{}{
		"webhookName": name,
		"eventType":   wrongEventType,
	}
	webhookConfig := acctest.ExecuteTemplate("TestAccWebhookEventTypesValidation", `
		resource "artifactory_artifact_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["{{ .eventType }}"]
			criteria {
				any_local  = true
				any_remote = true
				repo_keys  = []
			}
			handler {
				url = "http://tempurl.org"
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config:      webhookConfig,
				ExpectError: regexp.MustCompile(fmt.Sprintf("event_type %s not supported for domain artifact", wrongEventType)),
			},
		},
	})
}

func testAccWebhookHandlerValidation_EmptyProxy(t *testing.T) {
	id := test.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_artifact_webhook.%s", name)

	params := map[string]interface{}{
		"webhookName": name,
	}
	webhookConfig := acctest.ExecuteTemplate("TestAccWebhookEventTypesValidation", `
		resource "artifactory_artifact_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["deployed"]
			criteria {
				any_local  = true
				any_remote = true
				repo_keys  = []
			}
			handler {
				url   = "http://tempurl.org"
				proxy = ""
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config:      webhookConfig,
				ExpectError: regexp.MustCompile(`expected "proxy" to not be an empty string`),
			},
		},
	})
}

func testAccWebhookHandlerValidation_ProxyWithURL(t *testing.T) {
	id := test.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_artifact_webhook.%s", name)

	params := map[string]interface{}{
		"webhookName": name,
	}
	webhookConfig := acctest.ExecuteTemplate("TestAccWebhookEventTypesValidation", `
		resource "artifactory_artifact_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["deployed"]
			criteria {
				any_local  = true
				any_remote = true
				repo_keys  = []
			}
			handler {
				url   = "http://tempurl.org"
				proxy = "http://tempurl.org"
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config:      webhookConfig,
				ExpectError: regexp.MustCompile(`expected "proxy" not to be a valid url, got http://tempurl.org`),
			},
		},
	})
}

func testAccWebhookAllTypes(t *testing.T) {
	// Can only realistically test these 3 types of webhook since creating
	// build, release_bundle, or distribution in test environment is almost impossible
	for _, webhookType := range []string{"artifact", "artifact_property", "docker"} {
		t.Run(fmt.Sprintf("TestWebhook%s", strings.Title(strings.ToLower(webhookType))), func(t *testing.T) {
			resource.Test(webhookTestCase(webhookType, t))
		})
	}
}

func webhookTestCase(webhookType string, t *testing.T) (*testing.T, resource.TestCase) {
	id := test.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_%s_webhook.%s", webhookType, name)

	repoType := domainRepoTypeLookup[webhookType]
	repoName := fmt.Sprintf("%s-local-%d", webhookType, id)
	eventTypes := webhook.DomainEventTypesSupported[webhookType]

	params := map[string]interface{}{
		"repoName":    repoName,
		"repoType":    repoType,
		"webhookType": webhookType,
		"webhookName": name,
		"eventTypes":  eventTypes,
		"anyLocal":    test.RandBool(),
		"anyRemote":   test.RandBool(),
	}
	webhookConfig := acctest.ExecuteTemplate("TestAccWebhook{{ .webhookType }}Type", `
		resource "artifactory_local_{{ .repoType }}_repository" "{{ .repoName }}" {
			key = "{{ .repoName }}"
		}

		resource "artifactory_{{ .webhookType }}_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = [{{ range $index, $eventType := .eventTypes}}{{if $index}},{{end}}"{{$eventType}}"{{end}}]
			criteria {
				any_local  = {{ .anyLocal }}
				any_remote = {{ .anyRemote }}
				repo_keys  = ["{{ .repoName }}"]
				include_patterns = ["foo/**"]
				exclude_patterns = ["bar/**"]
			}
			handler {
				url                 = "http://tempurl.org"
				secret              = "fake-secret"
				custom_http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
			}
			handler {
				url                 = "http://tempurl.org"
				secret              = "fake-secret-2"
				custom_http_headers = {
					header-3 = "value-3"
					header-4 = "value-4"
				}
			}

			depends_on = [artifactory_local_{{ .repoType }}_repository.{{ .repoName }}]
		}
	`, params)

	testChecks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(fqrn, "key", name),
		resource.TestCheckResourceAttr(fqrn, "event_types.#", fmt.Sprintf("%d", len(eventTypes))),
		resource.TestCheckResourceAttr(fqrn, "criteria.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_local", fmt.Sprintf("%t", params["anyLocal"])),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_remote", fmt.Sprintf("%t", params["anyRemote"])),
		resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.repo_keys.*", repoName),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.0", "foo/**"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.0", "bar/**"),
		resource.TestCheckResourceAttr(fqrn, "handler.#", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.url", "http://tempurl.org"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.secret", "fake-secret"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-1", "value-1"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-2", "value-2"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.url", "http://tempurl.org"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.secret", "fake-secret-2"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.custom_http_headers.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.custom_http_headers.header-3", "value-3"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.custom_http_headers.header-4", "value-4"),
	}

	for _, eventType := range eventTypes {
		eventTypeCheck := resource.TestCheckTypeSetElemAttr(fqrn, "event_types.*", eventType)
		testChecks = append(testChecks, eventTypeCheck)
	}

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, testCheckWebhook),

		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check:  resource.ComposeTestCheckFunc(testChecks...),
			},
		},
	}
}

func testCheckWebhook(id string, request *resty.Request) (*resty.Response, error) {
	return request.
		SetPathParam("webhookKey", id).
		AddRetryCondition(client.NeverRetry).
		Get(webhook.WebhookUrl)
}

// Unit tests for state migration func
func testWebhookResourceStateUpgradeV1(t *testing.T) {
	v1Data := map[string]interface{}{
		"url":    "http://tempurl.org",
		"secret": "fake-secret",
		"proxy":  "fake-proxy-key",
		"custom_http_headers": map[string]interface{}{
			"header-1": "fake-value-1",
			"header-2": "fake-value-2",
		},
	}
	v2Data := map[string]interface{}{
		"handler": []map[string]interface{}{
			{
				"url":    "http://tempurl.org",
				"secret": "fake-secret",
				"proxy":  "fake-proxy-key",
				"custom_http_headers": map[string]interface{}{
					"header-1": "fake-value-1",
					"header-2": "fake-value-2",
				},
			},
		},
	}

	actual, err := webhook.ResourceStateUpgradeV1(context.Background(), v1Data, nil)

	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(v2Data, actual) {
		t.Fatalf("expected: %v\n\ngot: %v", v2Data, actual)
	}
}
