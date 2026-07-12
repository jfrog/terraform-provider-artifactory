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

package remote_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRemoteGenericRepository_full(t *testing.T) {
	const remoteGenericRepoBasicWithPropagate = `
		resource "artifactory_remote_generic_repository" "%s" {
			key                     		= "%s"
			description 					= "This is a test"
			url                     		= "https://registry.npmjs.org/"
			repo_layout_ref         		= "simple-default"
			propagate_query_params  		= true
			retrieve_sha256_from_server     = true
			retrieval_cache_period_seconds  = 70
		}
	`
	id := testutil.RandomInt()
	name := fmt.Sprintf("remote-test-repo-basic%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_generic_repository.%s", name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(remoteGenericRepoBasicWithPropagate, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "propagate_query_params", "true"),
					resource.TestCheckResourceAttr(fqrn, "retrieve_sha256_from_server", "true"),
				),
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

func TestAccRemoteGenericRepository_migrate_to_schema_v4(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-generic-remote", "artifactory_remote_generic_repository")

	const temp = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
			key                     		= "{{ .name }}"
			description 					= "This is a test"
			url                     		= "https://registry.npmjs.org/"
			repo_layout_ref         		= "simple-default"
			propagate_query_params  		= true
			retrieval_cache_period_seconds  = 70
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteRepository_generic_migrate_to_schema_v4", temp, params)

	resource.Test(t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.0.0",
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckNoResourceAttr(fqrn, "retrieve_sha256_from_server"),
				),
			},
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.8.3",
					},
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccRemoteGenericRepository_with_custom_http_headers(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-generic-remote", "artifactory_remote_generic_repository")

	const tmplWithHeaders = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://registry.npmjs.org/"
			custom_http_headers = [
				{ name = "x-api-key",    value = "test-key-value", sensitive = false },
				{ name = "x-ms-version", value = "2021-12-02" },
			]
		}
	`

	const tmplWithSensitiveHeader = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://registry.npmjs.org/"
			custom_http_headers = [
				{ name = "x-api-key", value = "new-key-value", sensitive = true },
			]
		}
	`

	const tmplUpdatedHeaders = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://registry.npmjs.org/"
			custom_http_headers = [
				{ name = "x-api-key", value = "new-key-value" },
			]
		}
	`

	const tmplNoHeaders = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://registry.npmjs.org/"
		}
	`

	params := map[string]interface{}{"name": name}
	configWithHeaders := util.ExecuteTemplate("TestAccRemoteGenericRepository_with_custom_http_headers_set", tmplWithHeaders, params)
	configWithSensitive := util.ExecuteTemplate("TestAccRemoteGenericRepository_with_custom_http_headers_sensitive", tmplWithSensitiveHeader, params)
	configUpdated := util.ExecuteTemplate("TestAccRemoteGenericRepository_with_custom_http_headers_update", tmplUpdatedHeaders, params)
	configNoHeaders := util.ExecuteTemplate("TestAccRemoteGenericRepository_with_custom_http_headers_clear", tmplNoHeaders, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				// create with 2 headers, sensitive=false (explicit) and sensitive omitted (defaults to false)
				Config: configWithHeaders,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.name", "x-api-key"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.value", "test-key-value"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.sensitive", "false"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.1.name", "x-ms-version"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.1.value", "2021-12-02"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.1.sensitive", "false"),
				),
			},
			{
				// update to single header with sensitive=true (Artifactory encrypts server-side)
				Config: configWithSensitive,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.name", "x-api-key"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.value", "new-key-value"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.sensitive", "true"),
				),
			},
			{
				// update to single header with sensitive=false (default)
				Config: configUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.name", "x-api-key"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.sensitive", "false"),
				),
			},
			{
				// remove all headers
				Config: configNoHeaders,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.#", "0"),
				),
			},
			{
				// re-add headers to verify idempotency
				Config: configWithHeaders,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.#", "2"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"custom_http_headers"},
			},
		},
	})
}

func TestAccRemoteGenericRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-generic-remote", "artifactory_remote_generic_repository")

	const temp = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
			key                     		= "{{ .name }}"
			description 					= "This is a test"
			url                     		= "https://registry.npmjs.org/"
			repo_layout_ref         		= "simple-default"
			retrieve_sha256_from_server     = true
			retrieval_cache_period_seconds  = 70
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteRepository_generic_migrate_from_SDKv2", temp, params)

	resource.Test(t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.8.3",
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "retrieve_sha256_from_server", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccRemoteGenericRepository_migrate_to_schema_v5 verifies that upgrading from provider
// 12.11.3 (last release with V4 Framework schema, no custom_http_headers) to V5 produces an empty plan.
func TestAccRemoteGenericRepository_migrate_to_schema_v5(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-generic-remote", "artifactory_remote_generic_repository")

	const temp = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
			key                            = "{{ .name }}"
			description                    = "This is a test"
			url                            = "https://registry.npmjs.org/"
			repo_layout_ref                = "simple-default"
			retrieve_sha256_from_server    = true
			retrieval_cache_period_seconds = 70
		}
	`

	params := map[string]interface{}{"name": name}
	config := util.ExecuteTemplate("TestAccRemoteRepository_generic_migrate_to_schema_v5", temp, params)

	resource.Test(t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				// Last released version with Framework V4 schema (no custom_http_headers).
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.11.3",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
