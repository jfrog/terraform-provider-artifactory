// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/webhook"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

var domainRepoTypeLookup = map[string]string{
	"artifact":          "generic",
	"artifact_property": "generic",
	"docker":            "docker_v2",
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

func TestAccWebhook_AllRepoTypes(t *testing.T) {
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
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", testCheckWebhook),
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

func TestAccWebhook_AllRepoTypes_UpgradeFromSDKv2(t *testing.T) {
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

			{{if eq .webhookType "docker"}}
			lifecycle {
				ignore_changes = [
					block_pushing_schema1,
					tag_retention,
				]
			}
			{{end}}
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
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", testCheckWebhook),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.1.0",
					},
				},
				Config: config,
				Check:  resource.ComposeTestCheckFunc(testChecks...),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				// ConfigPlanChecks:         testutil.ConfigPlanChecks(""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	}
}
