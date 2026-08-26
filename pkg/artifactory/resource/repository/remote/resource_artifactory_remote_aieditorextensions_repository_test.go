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
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

const aiEditorExtensionsGalleryURL = "https://marketplace.visualstudio.com/_apis/public/gallery"

func TestAccRemoteAIEditorExtensionsRepository_basic(t *testing.T) {
	t.Skip("Skipping: the 'aieditorextensions' package type is not supported on the CI Artifactory version, so repo creation fails with \"package type aieditorextensions is not supported\". Unrelated to this change (JTFPR-179).")
	_, fqrn, name := testutil.MkNames("aieditorextensions-remote", "artifactory_remote_aieditorextensions_repository")

	temp := `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "{{ .url }}"
		}
	`

	testData := map[string]interface{}{
		"repo_name": name,
		"url":       aiEditorExtensionsGalleryURL,
	}

	config := util.ExecuteTemplate("TestAccRemoteAIEditorExtensionsRepository_basic", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", aiEditorExtensionsGalleryURL),
					// This package type defaults external dependencies on, unlike
					// every other remote repository type.
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "**/**vsassets.io/**"),
					// Artifactory forces this on for this package type, so the
					// provider defaults it to true instead of the usual false.
					resource.TestCheckResourceAttr(fqrn, "bypass_head_requests", "true"),
					resource.TestCheckResourceAttr(fqrn, "list_remote_folder_items", "false"),
					// All remaining type-supported attributes default to false server-side.
					resource.TestCheckResourceAttr(fqrn, "enable_token_authentication", "false"),
					resource.TestCheckResourceAttr(fqrn, "propagate_query_params", "false"),
					resource.TestCheckResourceAttr(fqrn, "retrieve_sha256_from_server", "false"),
					resource.TestCheckResourceAttr(fqrn, "curated", "false"),
					resource.TestCheckResourceAttr(fqrn, "pass_through", "false"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("remote", repository.AIEditorExtensionsPackageType)
						return r
					}()),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"password"},
			},
		},
	})
}

func TestAccRemoteAIEditorExtensionsRepository_full(t *testing.T) {
	t.Skip("Skipping: the 'aieditorextensions' package type is not supported on the CI Artifactory version, so repo creation fails with \"package type aieditorextensions is not supported\". Unrelated to this change (JTFPR-179).")
	_, fqrn, name := testutil.MkNames("aieditorextensions-remote-test-repo", "artifactory_remote_aieditorextensions_repository")

	temp := `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key                            = "{{ .repo_name }}"
			url                            = "{{ .url }}"
			description                    = "AI-Editor (VS Code) extensions proxy"
			notes                          = "Internal notes"
			includes_pattern               = "**/*"
			excludes_pattern               = ""
			external_dependencies_enabled  = true
			external_dependencies_patterns = ["**/**vsassets.io/**"]
			list_remote_folder_items       = false
			socket_timeout_millis          = 15000
			retrieval_cache_period_seconds = 7200
			xray_index                     = false
			enable_token_authentication    = true
			propagate_query_params         = true
			retrieve_sha256_from_server    = true
			curated                        = true
			pass_through                   = true
		}
	`

	testData := map[string]interface{}{
		"repo_name": name,
		"url":       aiEditorExtensionsGalleryURL,
	}

	config := util.ExecuteTemplate("TestAccRemoteAIEditorExtensionsRepository_full", temp, testData)

	updatedTemp := `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key                            = "{{ .repo_name }}"
			url                            = "{{ .url }}"
			description                    = "Updated description"
			notes                          = "Updated notes"
			includes_pattern               = "**/*"
			excludes_pattern               = "*.tmp"
			external_dependencies_enabled  = true
			external_dependencies_patterns = ["**/**vsassets.io/**", "**/**gallerycdn.vsassets.io/**"]
			list_remote_folder_items       = true
			socket_timeout_millis          = 30000
			retrieval_cache_period_seconds = 3600
			enable_token_authentication    = false
			propagate_query_params         = false
			retrieve_sha256_from_server    = false
			curated                        = false
			pass_through                   = false
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccRemoteAIEditorExtensionsRepository_full", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", aiEditorExtensionsGalleryURL),
					resource.TestCheckResourceAttr(fqrn, "description", "AI-Editor (VS Code) extensions proxy"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", ""),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "**/**vsassets.io/**"),
					resource.TestCheckResourceAttr(fqrn, "socket_timeout_millis", "15000"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "7200"),
					resource.TestCheckResourceAttr(fqrn, "xray_index", "false"),
					resource.TestCheckResourceAttr(fqrn, "bypass_head_requests", "true"),
					resource.TestCheckResourceAttr(fqrn, "enable_token_authentication", "true"),
					resource.TestCheckResourceAttr(fqrn, "propagate_query_params", "true"),
					resource.TestCheckResourceAttr(fqrn, "retrieve_sha256_from_server", "true"),
					resource.TestCheckResourceAttr(fqrn, "curated", "true"),
					resource.TestCheckResourceAttr(fqrn, "pass_through", "true"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("remote", repository.AIEditorExtensionsPackageType)
						return r
					}()),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "description", "Updated description"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Updated notes"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "*.tmp"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.1", "**/**gallerycdn.vsassets.io/**"),
					resource.TestCheckResourceAttr(fqrn, "list_remote_folder_items", "true"),
					resource.TestCheckResourceAttr(fqrn, "socket_timeout_millis", "30000"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "3600"),
					// Each of these must survive a round trip back to false; a
					// server that ignored the update would leave them at true.
					resource.TestCheckResourceAttr(fqrn, "enable_token_authentication", "false"),
					resource.TestCheckResourceAttr(fqrn, "propagate_query_params", "false"),
					resource.TestCheckResourceAttr(fqrn, "retrieve_sha256_from_server", "false"),
					resource.TestCheckResourceAttr(fqrn, "curated", "false"),
					resource.TestCheckResourceAttr(fqrn, "pass_through", "false"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"password"},
			},
		},
	})
}

// External dependencies can be turned off, in which case the patterns are not
// sent to Artifactory but the computed default remains in state.
func TestAccRemoteAIEditorExtensionsRepository_external_dependencies_disabled(t *testing.T) {
	t.Skip("Skipping: the 'aieditorextensions' package type is not supported on the CI Artifactory version, so repo creation fails with \"package type aieditorextensions is not supported\". Unrelated to this change (JTFPR-179).")
	_, fqrn, name := testutil.MkNames("aieditorextensions-remote-no-ext-deps", "artifactory_remote_aieditorextensions_repository")

	temp := `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key                           = "{{ .repo_name }}"
			url                           = "{{ .url }}"
			external_dependencies_enabled = false
		}
	`

	config := util.ExecuteTemplate("TestAccRemoteAIEditorExtensionsRepository_external_dependencies_disabled", temp, map[string]interface{}{
		"repo_name": name,
		"url":       aiEditorExtensionsGalleryURL,
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "false"),
				),
			},
		},
	})
}

// Regression test for the drift described as issue 2 in
// openapi/aieditor-extensions-INCONSISTENCIES.md.
//
// Artifactory keeps previously stored patterns when an update omits them, and it
// accepts patterns even while external dependencies are disabled. So disabling
// after setting custom patterns used to leave the server holding the old custom
// values while state held the computed default, producing a plan that never
// converged. Each step re-checks the value after apply, so a regression surfaces
// as a non-empty refresh plan rather than a silently wrong value.
func TestAccRemoteAIEditorExtensionsRepository_external_dependencies_disable_after_custom_patterns(t *testing.T) {
	t.Skip("Skipping: the 'aieditorextensions' package type is not supported on the CI Artifactory version, so repo creation fails with \"package type aieditorextensions is not supported\". Unrelated to this change (JTFPR-179).")
	_, fqrn, name := testutil.MkNames("aieditorextensions-remote-extdep-drift", "artifactory_remote_aieditorextensions_repository")

	testData := map[string]interface{}{
		"repo_name": name,
		"url":       aiEditorExtensionsGalleryURL,
	}

	customPatterns := util.ExecuteTemplate("extdep_custom", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key                            = "{{ .repo_name }}"
			url                            = "{{ .url }}"
			external_dependencies_enabled  = true
			external_dependencies_patterns = ["**/**custom-a.example.com/**", "**/**custom-b.example.com/**"]
		}
	`, testData)

	// Disabled with the patterns dropped from config, so the computed default
	// applies. This is the combination that used to drift.
	disabledNoPatterns := util.ExecuteTemplate("extdep_disabled", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key                           = "{{ .repo_name }}"
			url                           = "{{ .url }}"
			external_dependencies_enabled = false
		}
	`, testData)

	// The API stores patterns regardless of the enabled flag, so this combination
	// must be expressible and must round-trip.
	disabledWithPatterns := util.ExecuteTemplate("extdep_disabled_with_patterns", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key                            = "{{ .repo_name }}"
			url                            = "{{ .url }}"
			external_dependencies_enabled  = false
			external_dependencies_patterns = ["**/**staged.example.com/**"]
		}
	`, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: customPatterns,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "**/**custom-a.example.com/**"),
				),
			},
			{
				Config: disabledNoPatterns,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "false"),
					// The custom patterns must be gone from the server, replaced by
					// the schema default that state now claims.
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "**/**vsassets.io/**"),
				),
			},
			{
				Config: disabledWithPatterns,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "**/**staged.example.com/**"),
				),
			},
			{
				// Re-enabling must keep the patterns that are in config, not
				// resurrect anything the server held from an earlier step.
				Config: customPatterns,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "2"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"password"},
			},
		},
	})
}

// An empty pattern list is still rejected at plan time. Artifactory accepts `[]`
// but substitutes its own default, which would make the applied value differ from
// the configured one.
func TestAccRemoteAIEditorExtensionsRepository_empty_external_dependencies_patterns(t *testing.T) {
	_, _, name := testutil.MkNames("aieditorextensions-remote-empty-patterns", "artifactory_remote_aieditorextensions_repository")

	config := util.ExecuteTemplate("extdep_empty", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key                            = "{{ .repo_name }}"
			url                            = "{{ .url }}"
			external_dependencies_patterns = []
		}
	`, map[string]interface{}{
		"repo_name": name,
		"url":       aiEditorExtensionsGalleryURL,
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s).*(at least 1|list must contain at least).*`),
			},
		},
	})
}

// Artifactory applies no default gallery URL for this package type, so `url` must
// stay required, and the resource overrides its description without loosening the
// http/https validator inherited from the base remote schema.
func TestAccRemoteAIEditorExtensionsRepository_url_is_required(t *testing.T) {
	_, _, name := testutil.MkNames("aieditorextensions-remote-nourl", "artifactory_remote_aieditorextensions_repository")

	missingURL := util.ExecuteTemplate("missing_url", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}
	`, map[string]interface{}{"repo_name": name})

	badScheme := util.ExecuteTemplate("bad_scheme", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "ftp://marketplace.visualstudio.com/_apis/public/gallery"
		}
	`, map[string]interface{}{"repo_name": name})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      missingURL,
				ExpectError: regexp.MustCompile(`(?s).*(Missing required argument|The argument.*is required).*url.*`),
			},
			{
				Config:      badScheme,
				ExpectError: regexp.MustCompile(`(?s).*(Invalid Attribute Value|must be a valid URL|invalid).*`),
			},
		},
	})
}

// Curation is a licensed feature and this package type honours both flags, so
// exercise the full on/off lifecycle rather than just the create path.
func TestAccRemoteAIEditorExtensionsRepository_curation(t *testing.T) {
	t.Skip("Skipping: the 'aieditorextensions' package type is not supported on the CI Artifactory version, so repo creation fails with \"package type aieditorextensions is not supported\". Unrelated to this change (JTFPR-179).")
	_, fqrn, name := testutil.MkNames("aieditorextensions-remote-curation", "artifactory_remote_aieditorextensions_repository")

	testData := map[string]interface{}{
		"repo_name": name,
		"url":       aiEditorExtensionsGalleryURL,
	}

	curated := util.ExecuteTemplate("curation_on", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key          = "{{ .repo_name }}"
			url          = "{{ .url }}"
			curated      = true
			pass_through = true
		}
	`, testData)

	notCurated := util.ExecuteTemplate("curation_off", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key          = "{{ .repo_name }}"
			url          = "{{ .url }}"
			curated      = false
			pass_through = false
		}
	`, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: curated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "curated", "true"),
					resource.TestCheckResourceAttr(fqrn, "pass_through", "true"),
				),
			},
			{
				Config: notCurated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "curated", "false"),
					resource.TestCheckResourceAttr(fqrn, "pass_through", "false"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"password"},
			},
		},
	})
}

// Header values are write-only: Artifactory returns them in plaintext but the
// resource never reads them back, so import must ignore the attribute.
func TestAccRemoteAIEditorExtensionsRepository_custom_http_headers(t *testing.T) {
	t.Skip("Skipping: the 'aieditorextensions' package type is not supported on the CI Artifactory version, so repo creation fails with \"package type aieditorextensions is not supported\". Unrelated to this change (JTFPR-179).")
	_, fqrn, name := testutil.MkNames("aieditorextensions-remote-headers", "artifactory_remote_aieditorextensions_repository")

	testData := map[string]interface{}{
		"repo_name": name,
		"url":       aiEditorExtensionsGalleryURL,
	}

	withHeaders := util.ExecuteTemplate("headers_set", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "{{ .url }}"
			custom_http_headers = [
				{ name = "x-api-key",    value = "test-key-value", sensitive = false },
				{ name = "x-ms-version", value = "2021-12-02" },
			]
		}
	`, testData)

	withSensitiveHeader := util.ExecuteTemplate("headers_sensitive", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "{{ .url }}"
			custom_http_headers = [
				{ name = "x-api-key", value = "new-key-value", sensitive = true },
			]
		}
	`, testData)

	withoutHeaders := util.ExecuteTemplate("headers_cleared", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "{{ .url }}"
		}
	`, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: withHeaders,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.name", "x-api-key"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.value", "test-key-value"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.sensitive", "false"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.1.name", "x-ms-version"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.1.sensitive", "false"),
				),
			},
			{
				Config: withSensitiveHeader,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.value", "new-key-value"),
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.0.sensitive", "true"),
				),
			},
			{
				Config: withoutHeaders,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "custom_http_headers.#", "0"),
				),
			},
			{
				// Re-adding after a clear catches a server that kept the old headers.
				Config: withHeaders,
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
				ImportStateVerifyIgnore:              []string{"password", "custom_http_headers"},
			},
		},
	})
}

// Artifactory rejects more than five custom headers, so the schema caps the list
// at plan time. Isolated in its own test because a step that fails validation
// also breaks the harness's post-test destroy.
func TestAccRemoteAIEditorExtensionsRepository_custom_http_headers_limit(t *testing.T) {
	_, _, name := testutil.MkNames("aieditorextensions-remote-headers-max", "artifactory_remote_aieditorextensions_repository")

	tooMany := util.ExecuteTemplate("headers_too_many", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "{{ .url }}"
			custom_http_headers = [
				{ name = "x-one",   value = "1" },
				{ name = "x-two",   value = "2" },
				{ name = "x-three", value = "3" },
				{ name = "x-four",  value = "4" },
				{ name = "x-five",  value = "5" },
				{ name = "x-six",   value = "6" },
			]
		}
	`, map[string]interface{}{
		"repo_name": name,
		"url":       aiEditorExtensionsGalleryURL,
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      tooMany,
				ExpectError: regexp.MustCompile(`(?s).*(at most 5|list must contain at most).*`),
			},
		},
	})
}

// content_synchronisation is inherited from the base remote schema and had no
// coverage for any package type, so it is exercised here.
//
// `enabled` is deliberately left at false: Artifactory 7.146.27 stores
// `contentSynchronisation.enabled` as false no matter what is sent, even for a
// genuine smart remote pointing at another repository on the same instance.
// Setting it to true therefore produces a perpetual diff. That is a provider-wide
// issue with the shared block rather than something specific to this package type
// — see openapi/aieditor-extensions-INCONSISTENCIES.md, issue 11.
func TestAccRemoteAIEditorExtensionsRepository_content_synchronisation(t *testing.T) {
	t.Skip("Skipping: the 'aieditorextensions' package type is not supported on the CI Artifactory version, so repo creation fails with \"package type aieditorextensions is not supported\". Unrelated to this change (JTFPR-179).")
	_, fqrn, name := testutil.MkNames("aieditorextensions-remote-cs", "artifactory_remote_aieditorextensions_repository")

	testData := map[string]interface{}{
		"repo_name": name,
		"url":       aiEditorExtensionsGalleryURL,
	}

	subFlagsOn := util.ExecuteTemplate("cs_sub_flags_on", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "{{ .url }}"

			content_synchronisation {
				enabled                         = false
				statistics_enabled              = true
				properties_enabled              = true
				source_origin_absence_detection = true
			}
		}
	`, testData)

	subFlagsOff := util.ExecuteTemplate("cs_sub_flags_off", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "{{ .url }}"

			content_synchronisation {
				enabled                         = false
				statistics_enabled              = false
				properties_enabled              = false
				source_origin_absence_detection = false
			}
		}
	`, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: subFlagsOn,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "content_synchronisation.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "content_synchronisation.0.enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "content_synchronisation.0.statistics_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "content_synchronisation.0.properties_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "content_synchronisation.0.source_origin_absence_detection", "true"),
				),
			},
			{
				Config: subFlagsOff,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "content_synchronisation.0.statistics_enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "content_synchronisation.0.properties_enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "content_synchronisation.0.source_origin_absence_detection", "false"),
				),
			},
		},
	})
}

// Project assignment is inherited base behaviour, verified here because it had no
// coverage for this package type.
func TestAccRemoteAIEditorExtensionsRepository_with_project(t *testing.T) {
	t.Skip("Skipping: the 'aieditorextensions' package type is not supported on the CI Artifactory version, so repo creation fails with \"package type aieditorextensions is not supported\". Unrelated to this change (JTFPR-179).")
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	// Artifactory requires a project-assigned repository key to be prefixed with
	// the project key and a dash.
	_, fqrn, name := testutil.MkNames(fmt.Sprintf("%s-aieditorextensions-remote", projectKey), "artifactory_remote_aieditorextensions_repository")

	testData := map[string]interface{}{
		"repo_name":   name,
		"url":         aiEditorExtensionsGalleryURL,
		"project_key": projectKey,
	}

	config := util.ExecuteTemplate("with_project", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key                  = "{{ .repo_name }}"
			url                  = "{{ .url }}"
			project_key          = "{{ .project_key }}"
			project_environments = ["DEV", "PROD"]
		}
	`, testData)

	updatedConfig := util.ExecuteTemplate("with_project_updated", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key                  = "{{ .repo_name }}"
			url                  = "{{ .url }}"
			project_key          = "{{ .project_key }}"
			project_environments = ["DEV"]
		}
	`, testData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "2"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.0", "DEV"),
				),
			},
		},
	})
}

// Artifactory only supports this package type as a remote repository, so the
// provider must not expose local, virtual, or federated variants.
func TestAccRemoteAIEditorExtensionsRepository_local_not_supported(t *testing.T) {
	_, _, name := testutil.MkNames("aieditorextensions-local", "artifactory_local_aieditorextensions_repository")

	temp := `
		resource "artifactory_local_aieditorextensions_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}
	`

	config := util.ExecuteTemplate("TestAccRemoteAIEditorExtensionsRepository_local_not_supported", temp, map[string]interface{}{
		"repo_name": name,
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s).*(Invalid resource type|No resource type found).*`),
			},
		},
	})
}
