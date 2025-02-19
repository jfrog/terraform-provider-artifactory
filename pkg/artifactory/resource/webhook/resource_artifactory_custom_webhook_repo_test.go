package webhook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/webhook"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccCustomWebhook_AllRepoTypes(t *testing.T) {
	// Can only realistically test these 3 types of webhook since creating
	// build, release_bundle, or distribution in test environment is almost impossible
	for _, webhookType := range []string{"artifact", "artifact_property", "docker"} {
		t.Run(webhookType, func(t *testing.T) {
			resource.Test(customWebhookTestCase(webhookType, t))
		})
	}
}

func customWebhookTestCase(webhookType string, t *testing.T) (*testing.T, resource.TestCase) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("custom-webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_%s_custom_webhook.%s", webhookType, name)

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

		resource "artifactory_{{ .webhookType }}_custom_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = [{{ range $index, $eventType := .eventTypes}}{{if $index}},{{end}}"{{$eventType}}"{{end}}]
			criteria {
				any_local  = {{ .anyLocal }}
				any_remote = {{ .anyRemote }}
				any_federated = {{ .anyFederated }}
				repo_keys  = ["{{ .repoName }}"]
				include_patterns = ["foo/**"]
				exclude_patterns = ["bar/**"]
			}
			handler {
				url     = "https://google.com"
				method = "POST"
				secrets = {
					secret1 = "value1"
					secret2 = "value2"
				}
				http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
				payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/1\" } }"
			}
			handler {
				url     = "https://yahoo.com"
				method = "PUT"
				secrets = {
					secret3 = "value3"
					secret4 = "value4"
				}
				http_headers = {
					header-3 = "value-3"
					header-4 = "value-4"
				}
				payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/2\" } }"
			}
			handler {
				url = "https://msnbc.com"
				payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/3\" } }"
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
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_federated", fmt.Sprintf("%t", params["anyFederated"])),
		resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.repo_keys.*", repoName),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.0", "foo/**"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.0", "bar/**"),
		resource.TestCheckResourceAttr(fqrn, "handler.#", "3"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.method", "POST"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.secret1", "value1"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.secret2", "value2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.header-1", "value-1"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.header-2", "value-2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.payload", "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/1\" } }"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.url", "https://yahoo.com"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.method", "PUT"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.secrets.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.secrets.secret3", "value3"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.secrets.secret4", "value4"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.http_headers.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.http_headers.header-3", "value-3"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.http_headers.header-4", "value-4"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.payload", "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/2\" } }"),
		resource.TestCheckResourceAttr(fqrn, "handler.2.url", "https://msnbc.com"),
		resource.TestCheckNoResourceAttr(fqrn, "handler.2.secrets"),
		resource.TestCheckNoResourceAttr(fqrn, "handler.2.http_headers"),
		resource.TestCheckResourceAttr(fqrn, "handler.2.payload", "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/3\" } }"),
	}

	for _, eventType := range eventTypes {
		eventTypeCheck := resource.TestCheckTypeSetElemAttr(fqrn, "event_types.*", eventType)
		testChecks = append(testChecks, eventTypeCheck)
	}

	return t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", testCheckWebhook),

		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check:  resource.ComposeTestCheckFunc(testChecks...),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"handler.0.secrets", "handler.1.secrets"},
			},
		},
	}
}

func TestAccCustomWebhook_AllRepoTypes_UpgradeFromSDKv2(t *testing.T) {
	// Can only realistically test these 3 types of webhook since creating
	// build, release_bundle, or distribution in test environment is almost impossible
	for _, webhookType := range []string{ /*"artifact", "artifact_property", */ "docker"} {
		t.Run(webhookType, func(t *testing.T) {
			resource.Test(customWebhookMigrateFromSDKv2TestCase(webhookType, t))
		})
	}
}

func customWebhookMigrateFromSDKv2TestCase(webhookType string, t *testing.T) (*testing.T, resource.TestCase) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("custom-webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_%s_custom_webhook.%s", webhookType, name)

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

		resource "artifactory_{{ .webhookType }}_custom_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = [{{ range $index, $eventType := .eventTypes}}{{if $index}},{{end}}"{{$eventType}}"{{end}}]
			criteria {
				any_local  = {{ .anyLocal }}
				any_remote = {{ .anyRemote }}
				any_federated = {{ .anyFederated }}
				repo_keys  = ["{{ .repoName }}"]
				include_patterns = ["foo/**"]
				exclude_patterns = ["bar/**"]
			}
			handler {
				url     = "https://google.com"
				secrets = {
					secret1 = "value1"
					secret2 = "value2"
				}
				http_headers = {
					header-1 = "value-1"
					header-2 = "value-2"
				}
				payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/1\" } }"
			}
			handler {
				url     = "https://yahoo.com"
				secrets = {
					secret3 = "value3"
					secret4 = "value4"
				}
				http_headers = {
					header-3 = "value-3"
					header-4 = "value-4"
				}
				payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/2\" } }"
			}
			handler {
				url = "https://msnbc.com"
				payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/3\" } }"
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
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_federated", fmt.Sprintf("%t", params["anyFederated"])),
		resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.repo_keys.*", repoName),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.0", "foo/**"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.exclude_patterns.0", "bar/**"),
		resource.TestCheckResourceAttr(fqrn, "handler.#", "3"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.url", "https://google.com"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.secret1", "value1"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.secrets.secret2", "value2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.header-1", "value-1"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.http_headers.header-2", "value-2"),
		resource.TestCheckResourceAttr(fqrn, "handler.0.payload", "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/1\" } }"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.url", "https://yahoo.com"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.secrets.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.secrets.secret3", "value3"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.secrets.secret4", "value4"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.http_headers.%", "2"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.http_headers.header-3", "value-3"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.http_headers.header-4", "value-4"),
		resource.TestCheckResourceAttr(fqrn, "handler.1.payload", "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/2\" } }"),
		resource.TestCheckResourceAttr(fqrn, "handler.2.url", "https://msnbc.com"),
		resource.TestCheckNoResourceAttr(fqrn, "handler.2.secrets"),
		resource.TestCheckNoResourceAttr(fqrn, "handler.2.http_headers"),
		resource.TestCheckResourceAttr(fqrn, "handler.2.payload", "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path/3\" } }"),
	}

	for _, eventType := range eventTypes {
		eventTypeCheck := resource.TestCheckTypeSetElemAttr(fqrn, "event_types.*", eventType)
		testChecks = append(testChecks, eventTypeCheck)
	}

	return t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", testCheckWebhook),
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
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				ConfigPlanChecks:         testutil.ConfigPlanChecks(fqrn),
				// ConfigPlanChecks: resource.ConfigPlanChecks{
				// 	PreApply: []plancheck.PlanCheck{
				// 		plancheck.ExpectEmptyPlan(),
				// 	},
				// },
			},
		},
	}
}
