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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

// jetBrainsPluginsMarketplaceURL is the JetBrains Marketplace URL the backend
// declares as this package type's remote registry (RepoConfigDefaultValues.java:
// DEFAULT_JETBRAINS_PLUGINS_REGISTRY_URL). Used as the recommended URL in tests.
const jetBrainsPluginsMarketplaceURL = "https://plugins.jetbrains.com"

// jetBrainsPluginsRESTFields maps Terraform attribute names to REST JSON field
// names for the parity check. Only non-secret fields are compared.
// `package_type` is not exported by the resource schema (only by the data
// source), so it is deliberately absent.
var jetBrainsPluginsRESTFields = map[string]string{
	"key":                                   "key",
	"url":                                   "url",
	"repo_layout_ref":                       "repoLayoutRef",
	"description":                           "description",
	"notes":                                 "notes",
	"includes_pattern":                      "includesPattern",
	"excludes_pattern":                      "excludesPattern",
	"hard_fail":                             "hardFail",
	"offline":                               "offline",
	"blacked_out":                           "blackedOut",
	"xray_index":                            "xrayIndex",
	"priority_resolution":                   "priorityResolution",
	"store_artifacts_locally":               "storeArtifactsLocally",
	"socket_timeout_millis":                 "socketTimeoutMillis",
	"retrieval_cache_period_seconds":        "retrievalCachePeriodSecs",
	"missed_cache_period_seconds":           "missedRetrievalCachePeriodSecs",
	"assumed_offline_period_secs":           "assumedOfflinePeriodSecs",
	"bypass_head_requests":                  "bypassHeadRequests",
	"list_remote_folder_items":              "listRemoteFolderItems",
	"enable_cookie_management":              "enableCookieManagement",
	"allow_any_host_auth":                   "allowAnyHostAuth",
	"block_mismatching_mime_types":          "blockMismatchingMimeTypes",
	"synchronize_properties":                "synchronizeProperties",
	"disable_url_normalization":             "disableUrlNormalization",
	"archive_browsing_enabled":              "archiveBrowsingEnabled",
	"metadata_retrieval_timeout_secs":       "metadataRetrievalTimeoutSecs",
	"unused_artifacts_cleanup_period_hours": "unusedArtifactsCleanupPeriodHours",
}

// checkJetBrainsPluginsRESTMatchesState reads the repository back over the
// public REST API and asserts every value matches Terraform state. Secret
// fields are excluded by construction (fieldMap).
func checkJetBrainsPluginsRESTMatchesState(t *testing.T, fqrn string, fieldMap map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[fqrn]
		if !ok {
			return fmt.Errorf("%s not found in state", fqrn)
		}
		key := rs.Primary.Attributes["key"]

		var repo map[string]interface{}
		response, err := acctest.GetTestResty(t).R().
			SetResult(&repo).
			Get("artifactory/api/repositories/" + key)
		if err != nil {
			return err
		}
		if response.IsError() {
			return fmt.Errorf("REST GET for repository %s failed: %s", key, response.Status())
		}

		for tfAttr, jsonKey := range fieldMap {
			got := fmt.Sprintf("%v", repo[jsonKey])
			if want := rs.Primary.Attributes[tfAttr]; got != want {
				return fmt.Errorf("field %q: REST=%q, Terraform=%q", jsonKey, got, want)
			}
		}
		return nil
	}
}

// TestAccRemoteJetBrainsPluginsRepository_basic exercises CRUD with a minimal
// config and asserts every default recorded in the spec (verified against the
// live REST API).
func TestAccRemoteJetBrainsPluginsRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("jetbrainsplugins-remote", "artifactory_remote_jetbrainsplugins_repository")

	config := util.ExecuteTemplate("jetbrainsplugins-remote-basic", `
		resource "artifactory_remote_jetbrainsplugins_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "{{ .url }}"
		}
	`, map[string]interface{}{
		"name": name,
		"url":  jetBrainsPluginsMarketplaceURL,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", jetBrainsPluginsMarketplaceURL),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					// REST-verified defaults for a minimal jetbrainsplugins remote.
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", ""),
					resource.TestCheckResourceAttr(fqrn, "hard_fail", "false"),
					resource.TestCheckResourceAttr(fqrn, "offline", "false"),
					resource.TestCheckResourceAttr(fqrn, "blacked_out", "false"),
					resource.TestCheckResourceAttr(fqrn, "xray_index", "false"),
					resource.TestCheckResourceAttr(fqrn, "priority_resolution", "false"),
					resource.TestCheckResourceAttr(fqrn, "store_artifacts_locally", "true"),
					resource.TestCheckResourceAttr(fqrn, "socket_timeout_millis", "15000"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "7200"),
					resource.TestCheckResourceAttr(fqrn, "missed_cache_period_seconds", "1800"),
					resource.TestCheckResourceAttr(fqrn, "metadata_retrieval_timeout_secs", "60"),
					resource.TestCheckResourceAttr(fqrn, "assumed_offline_period_secs", "300"),
					resource.TestCheckResourceAttr(fqrn, "unused_artifacts_cleanup_period_hours", "0"),
					resource.TestCheckResourceAttr(fqrn, "bypass_head_requests", "false"),
					resource.TestCheckResourceAttr(fqrn, "list_remote_folder_items", "false"),
					resource.TestCheckResourceAttr(fqrn, "enable_cookie_management", "false"),
					resource.TestCheckResourceAttr(fqrn, "allow_any_host_auth", "false"),
					resource.TestCheckResourceAttr(fqrn, "archive_browsing_enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "cdn_redirect", "false"),
					checkJetBrainsPluginsRESTMatchesState(t, fqrn, jetBrainsPluginsRESTFields),
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

// TestAccRemoteJetBrainsPluginsRepository_full applies a full config with every
// inherited base attribute set, updates it (flipping toggles and moving
// bounds), and asserts REST/Terraform parity after each apply.
func TestAccRemoteJetBrainsPluginsRepository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("jetbrainsplugins-remote-full", "artifactory_remote_jetbrainsplugins_repository")

	const temp = `
		resource "artifactory_remote_jetbrainsplugins_repository" "{{ .name }}" {
			key                                   = "{{ .name }}"
			url                                   = "{{ .url }}"
			description                           = "{{ .description }}"
			notes                                 = "{{ .notes }}"
			includes_pattern                      = "**/*"
			excludes_pattern                      = "{{ .excludes_pattern }}"
			repo_layout_ref                       = "simple-default"
			hard_fail                             = {{ .hard_fail }}
			offline                               = {{ .offline }}
			blacked_out                           = {{ .blacked_out }}
			xray_index                            = {{ .xray_index }}
			priority_resolution                   = {{ .priority_resolution }}
			store_artifacts_locally               = true
			socket_timeout_millis                 = {{ .socket_timeout_millis }}
			retrieval_cache_period_seconds        = {{ .retrieval_cache_period_seconds }}
			missed_cache_period_seconds           = {{ .missed_cache_period_seconds }}
			metadata_retrieval_timeout_secs       = {{ .metadata_retrieval_timeout_secs }}
			assumed_offline_period_secs           = {{ .assumed_offline_period_secs }}
			unused_artifacts_cleanup_period_hours = {{ .unused_artifacts_cleanup_period_hours }}
			bypass_head_requests                  = {{ .bypass_head_requests }}
			list_remote_folder_items              = {{ .list_remote_folder_items }}
			enable_cookie_management              = {{ .enable_cookie_management }}
			allow_any_host_auth                   = {{ .allow_any_host_auth }}
			block_mismatching_mime_types          = {{ .block_mismatching_mime_types }}
			synchronize_properties                = {{ .synchronize_properties }}
			disable_url_normalization             = {{ .disable_url_normalization }}
			archive_browsing_enabled              = {{ .archive_browsing_enabled }}
		}
	`

	params := map[string]interface{}{
		"name":                                  name,
		"url":                                   jetBrainsPluginsMarketplaceURL,
		"description":                           "JetBrains Plugins proxy",
		"notes":                                 "created by acceptance test",
		"excludes_pattern":                      "excluded/**",
		"hard_fail":                             true,
		"offline":                               true,
		"blacked_out":                           true,
		"xray_index":                            true,
		"priority_resolution":                   true,
		"socket_timeout_millis":                 20000,
		"retrieval_cache_period_seconds":        600,
		"missed_cache_period_seconds":           900,
		"metadata_retrieval_timeout_secs":       30,
		"assumed_offline_period_secs":           600,
		"unused_artifacts_cleanup_period_hours": 5,
		"bypass_head_requests":                  true,
		"list_remote_folder_items":              true,
		"enable_cookie_management":              true,
		"allow_any_host_auth":                   true,
		"block_mismatching_mime_types":          false,
		"synchronize_properties":                true,
		"disable_url_normalization":             true,
		"archive_browsing_enabled":              true,
	}

	updated := map[string]interface{}{}
	for k, v := range params {
		updated[k] = v
	}
	// Flip toggles and move bounds so the update proves the change lands on
	// the server, not just in state.
	updated["description"] = ""
	updated["notes"] = ""
	updated["excludes_pattern"] = ""
	updated["hard_fail"] = false
	updated["offline"] = false
	updated["blacked_out"] = false
	updated["xray_index"] = false
	updated["priority_resolution"] = false
	updated["socket_timeout_millis"] = 30000
	updated["retrieval_cache_period_seconds"] = 1200
	updated["missed_cache_period_seconds"] = 2400
	updated["metadata_retrieval_timeout_secs"] = 90
	updated["assumed_offline_period_secs"] = 900
	updated["unused_artifacts_cleanup_period_hours"] = 0
	updated["bypass_head_requests"] = false
	updated["list_remote_folder_items"] = false
	updated["enable_cookie_management"] = false
	updated["allow_any_host_auth"] = false
	updated["block_mismatching_mime_types"] = true
	updated["synchronize_properties"] = false
	updated["disable_url_normalization"] = false
	updated["archive_browsing_enabled"] = false

	config := util.ExecuteTemplate("jetbrainsplugins-remote-full", temp, params)
	updatedConfig := util.ExecuteTemplate("jetbrainsplugins-remote-full-updated", temp, updated)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "description", "JetBrains Plugins proxy"),
					resource.TestCheckResourceAttr(fqrn, "hard_fail", "true"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "600"),
					resource.TestCheckResourceAttr(fqrn, "bypass_head_requests", "true"),
					checkJetBrainsPluginsRESTMatchesState(t, fqrn, jetBrainsPluginsRESTFields),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "description", ""),
					resource.TestCheckResourceAttr(fqrn, "hard_fail", "false"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "1200"),
					resource.TestCheckResourceAttr(fqrn, "bypass_head_requests", "false"),
					resource.TestCheckResourceAttr(fqrn, "block_mismatching_mime_types", "true"),
					checkJetBrainsPluginsRESTMatchesState(t, fqrn, jetBrainsPluginsRESTFields),
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

// TestAccRemoteJetBrainsPluginsRepository_url_is_required covers the
// spec.edgeCases entry `url_is_required`: the schema must reject a config
// without `url` at plan time. Artifactory itself would return 400 "No URL
// defined for remote repository" if the schema let it through.
func TestAccRemoteJetBrainsPluginsRepository_url_is_required(t *testing.T) {
	_, _, name := testutil.MkNames("jetbrainsplugins-remote-no-url", "artifactory_remote_jetbrainsplugins_repository")

	config := util.ExecuteTemplate("jetbrainsplugins-remote-no-url", `
		resource "artifactory_remote_jetbrainsplugins_repository" "{{ .name }}" {
			key = "{{ .name }}"
		}
	`, map[string]interface{}{"name": name})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s)The argument "url" is required`),
			},
		},
	})
}

// TestAccRemoteJetBrainsPluginsRepository_url_must_be_http_or_https covers the
// spec.edgeCases entry `url_must_be_http_or_https`.
func TestAccRemoteJetBrainsPluginsRepository_url_must_be_http_or_https(t *testing.T) {
	_, _, name := testutil.MkNames("jetbrainsplugins-remote-bad-url", "artifactory_remote_jetbrainsplugins_repository")

	config := util.ExecuteTemplate("jetbrainsplugins-remote-bad-url", `
		resource "artifactory_remote_jetbrainsplugins_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "not-a-url"
		}
	`, map[string]interface{}{"name": name})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value|is not a valid URL`),
			},
		},
	})
}

// TestAccRemoteJetBrainsPluginsRepository_invalid_key covers the
// spec.edgeCases entry `invalid_key`: reserved characters in `key`.
func TestAccRemoteJetBrainsPluginsRepository_invalid_key(t *testing.T) {
	config := util.ExecuteTemplate("jetbrainsplugins-remote-bad-key", `
		resource "artifactory_remote_jetbrainsplugins_repository" "bad_key" {
			key = "invalid key with spaces"
			url = "{{ .url }}"
		}
	`, map[string]interface{}{"url": jetBrainsPluginsMarketplaceURL})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value|must be 1 - 64 alphanumeric`),
			},
		},
	})
}

// TestAccRemoteJetBrainsPluginsRepository_curated_not_supported covers the
// spec.edgeCases entry `curated_not_supported`. Curation for JetBrains Plugins
// is not present in the live curation catalog on the target instance today,
// and the CurationRepoConfigInterceptor rejects `curated=true` with HTTP 400.
// The provider does not expose `curated`, so the test proves the raw REST
// still 400s if a user tries to enable curation out-of-band via the API.
func TestAccRemoteJetBrainsPluginsRepository_curated_not_supported(t *testing.T) {
	_, fqrn, name := testutil.MkNames("jetbrainsplugins-remote-curated", "artifactory_remote_jetbrainsplugins_repository")

	config := util.ExecuteTemplate("jetbrainsplugins-remote-curated", `
		resource "artifactory_remote_jetbrainsplugins_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "{{ .url }}"
		}
	`, map[string]interface{}{
		"name": name,
		"url":  jetBrainsPluginsMarketplaceURL,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					// After the resource lands on the server, attempt to enable
					// curation via raw REST — the backend must reject the change.
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[fqrn]
						if !ok {
							return fmt.Errorf("%s not found in state", fqrn)
						}
						key := rs.Primary.Attributes["key"]

						resp, err := acctest.GetTestResty(t).R().
							SetHeader("Content-Type", "application/json").
							SetBody(`{"curated":true}`).
							Post("artifactory/api/repositories/" + key)
						if err != nil {
							return err
						}
						if resp.StatusCode() != 400 {
							return fmt.Errorf("expected HTTP 400 when enabling curation for jetbrainsplugins, got %s", resp.Status())
						}
						return nil
					},
				),
			},
		},
	})
}
