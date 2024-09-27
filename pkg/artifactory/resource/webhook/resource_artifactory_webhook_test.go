package webhook_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/webhook"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

var domainRepoTypeLookup = map[string]string{
	"artifact":          "generic",
	"artifact_property": "generic",
	"docker":            "docker_v2",
}

var domainValidationErrorMessageLookup = map[string]string{
	"artifact":                   `repo_keys cannot be empty when any_local, any_remote, and any_federated are\s*false`,
	"artifact_property":          `repo_keys cannot be empty when any_local, any_remote, and any_federated are\s*false`,
	"docker":                     `repo_keys cannot be empty when any_local, any_remote, and any_federated are\s*false`,
	"build":                      `selected_builds or include_patterns cannot be empty when any_build is false`,
	"release_bundle":             `registered_release_bundle_names cannot be empty when any_release_bundle is\s*false`,
	"distribution":               `registered_release_bundle_names cannot be empty when any_release_bundle is\s*false`,
	"artifactory_release_bundle": `registered_release_bundle_names cannot be empty when any_release_bundle is\s*false`,
}

var repoTemplate = `
	resource "artifactory_{{ .webhookType }}_webhook" "{{ .webhookName }}" {
		key         = "{{ .webhookName }}"
		description = "test description"
		event_types = [{{ range $index, $eventType := .eventTypes}}{{if $index}},{{end}}"{{$eventType}}"{{end}}]
		criteria {
			any_local = false
			any_remote = false
			any_federated = false
			repo_keys = []
		}
		handler {
			url = "https://google.com"
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
			url = "https://google.com"
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

func TestAccWebhook_BuildWithIncludePatterns(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_build_webhook.%s", name)

	params := map[string]interface{}{
		"webhookName": name,
	}
	webhookConfig := util.ExecuteTemplate("TestAccWebhookBuildPatterns", `
		resource "artifactory_build_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["uploaded"]
			criteria {
				any_build  = false
				selected_builds = []
				include_patterns = ["foo"]
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
				Config: webhookConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.0", "foo"),
				),
			},
		},
	})
}

func TestAccWebhook_AllTypes(t *testing.T) {
	// Can only realistically test these 3 types of webhook since creating
	// build, release_bundle, or distribution in test environment is almost impossible
	for _, webhookType := range []string{"artifact", "artifact_property", "docker"} {
		t.Run(webhookType, func(t *testing.T) {
			resource.Test(webhookTestCase(webhookType, t))
		})
	}
}

func webhookTestCase(webhookType string, t *testing.T) (*testing.T, resource.TestCase) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_%s_webhook.%s", webhookType, name)

	repoType := domainRepoTypeLookup[webhookType]
	repoName := fmt.Sprintf("%s-local-%d", webhookType, id)
	eventTypes := webhook.DomainEventTypesSupported[webhookType]

	params := map[string]interface{}{
		"repoName":     repoName,
		"repoType":     repoType,
		"webhookType":  webhookType,
		"webhookName":  name,
		"eventTypes":   eventTypes,
		"anyLocal":     testutil.RandBool(),
		"anyRemote":    testutil.RandBool(),
		"anyFederated": testutil.RandBool(),
	}
	webhookConfig := util.ExecuteTemplate("TestAccWebhook{{ .webhookType }}Type", `
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
				any_federated = {{ .anyFederated }}
				repo_keys  = [artifactory_local_{{ .repoType }}_repository.{{ .repoName }}.key]
				include_patterns = ["foo/**"]
				exclude_patterns = ["bar/**"]
			}
			handler {
				url                    = "https://google.com"
				secret                 = "fake-secret"
				use_secret_for_signing = true
				custom_http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
			}
			handler {
				url = "https://tempurl.com"
			}
		}
	`, params)

	updatedConfig := util.ExecuteTemplate("TestAccWebhook{{ .webhookType }}Type", `
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
				any_federated = {{ .anyFederated }}
				repo_keys  = [artifactory_local_{{ .repoType }}_repository.{{ .repoName }}.key]
			}
			handler {
				url                    = "https://google.com"
				secret                 = "fake-secret"
				custom_http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
			}
			handler {
				url = "https://tempurl.com"
			}
		}
	`, params)

	testChecks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(fqrn, "key", name),
		resource.TestCheckResourceAttr(fqrn, "event_types.#", fmt.Sprintf("%d", len(eventTypes))),
		resource.TestCheckResourceAttr(fqrn, "criteria.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_local", fmt.Sprintf("%t", params["anyLocal"])),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_remote", fmt.Sprintf("%t", params["anyRemote"])),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_federated", fmt.Sprintf("%t", params["anyFederated"])),
		resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.repo_keys.*", repoName),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.0", "foo/**"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.0", "bar/**"),
		resource.TestCheckResourceAttr(fqrn, "handler.#", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.secret", "fake-secret"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.use_secret_for_signing", "true"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-1", "value-1"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-2", "value-2"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.url", "https://tempurl.com"),
		resource.TestCheckNoResourceAttr(fqrn, "handler.1.custom_http_headers"),
	}

	updatedTestChecks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(fqrn, "key", name),
		resource.TestCheckResourceAttr(fqrn, "event_types.#", fmt.Sprintf("%d", len(eventTypes))),
		resource.TestCheckResourceAttr(fqrn, "criteria.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_local", fmt.Sprintf("%t", params["anyLocal"])),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_remote", fmt.Sprintf("%t", params["anyRemote"])),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_federated", fmt.Sprintf("%t", params["anyFederated"])),
		resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.repo_keys.*", repoName),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.#", "0"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.#", "0"),
		resource.TestCheckResourceAttr(fqrn, "handler.#", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.secret", "fake-secret"),
		resource.TestCheckNoResourceAttr(fqrn, "handler.0.use_secret_for_signing"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-1", "value-1"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-2", "value-2"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.url", "https://tempurl.com"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.custom_http_headers.#", "0"),
	}

	for _, eventType := range eventTypes {
		eventTypeCheck := resource.TestCheckTypeSetElemAttr(fqrn, "event_types.*", eventType)
		testChecks = append(testChecks, eventTypeCheck)
	}

	return t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "key", testCheckWebhook),
		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check:  resource.ComposeTestCheckFunc(testChecks...),
			},
			{
				Config: updatedConfig,
				Check:  resource.ComposeTestCheckFunc(updatedTestChecks...),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"handler.0.secret", "handler.1.secret"},
			},
		},
	}
}

func TestAccWebhook_UpgradeFromSDKv2(t *testing.T) {
	// Can only realistically test these 3 types of webhook since creating
	// build, release_bundle, or distribution in test environment is almost impossible
	for _, webhookType := range []string{"artifact", "artifact_property", "docker"} {
		t.Run(webhookType, func(t *testing.T) {
			resource.Test(webhookMigrateFromSDKv2TestCase(webhookType, t))
		})
	}
}

func webhookMigrateFromSDKv2TestCase(webhookType string, t *testing.T) (*testing.T, resource.TestCase) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_%s_webhook.%s", webhookType, name)

	repoType := domainRepoTypeLookup[webhookType]
	repoName := fmt.Sprintf("%s-local-%d", webhookType, id)
	eventTypes := webhook.DomainEventTypesSupported[webhookType]

	params := map[string]interface{}{
		"repoName":     repoName,
		"repoType":     repoType,
		"webhookType":  webhookType,
		"webhookName":  name,
		"eventTypes":   eventTypes,
		"anyLocal":     testutil.RandBool(),
		"anyRemote":    testutil.RandBool(),
		"anyFederated": testutil.RandBool(),
	}
	config := util.ExecuteTemplate("TestAccWebhook{{ .webhookType }}Type", `
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
				any_federated = {{ .anyFederated }}
				repo_keys  = [artifactory_local_{{ .repoType }}_repository.{{ .repoName }}.key]
				include_patterns = ["foo/**"]
				exclude_patterns = ["bar/**"]
			}
			handler {
				url                    = "https://google.com"
				secret                 = "fake-secret"
				use_secret_for_signing = true
				custom_http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
			}
			handler {
				url = "https://tempurl.com"
			}
		}
	`, params)

	testChecks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(fqrn, "key", name),
		resource.TestCheckResourceAttr(fqrn, "event_types.#", fmt.Sprintf("%d", len(eventTypes))),
		resource.TestCheckResourceAttr(fqrn, "criteria.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_local", fmt.Sprintf("%t", params["anyLocal"])),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_remote", fmt.Sprintf("%t", params["anyRemote"])),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_federated", fmt.Sprintf("%t", params["anyFederated"])),
		resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.repo_keys.*", repoName),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.0", "foo/**"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.0", "bar/**"),
		resource.TestCheckResourceAttr(fqrn, "handler.#", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.secret", "fake-secret"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.use_secret_for_signing", "true"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-1", "value-1"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.custom_http_headers.header-2", "value-2"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.url", "https://tempurl.com"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.secret", ""),
		resource.TestCheckNoResourceAttr(fqrn, "handler.1.custom_http_headers"),
	}

	for _, eventType := range eventTypes {
		eventTypeCheck := resource.TestCheckTypeSetElemAttr(fqrn, "event_types.*", eventType)
		testChecks = append(testChecks, eventTypeCheck)
	}

	return t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: acctest.VerifyDeleted(fqrn, "key", testCheckWebhook),
		Steps: []resource.TestStep{
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.1.0",
					},
				},
				Check: resource.ComposeTestCheckFunc(testChecks...),
			},
			{
				Config:                   config,
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	}
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

// Unit tests for state migration func
func TestWebhook_ResourceStateUpgradeV1(t *testing.T) {
	v1Data := map[string]interface{}{
		"url":    "https://google.com",
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
				"url":    "https://google.com",
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
