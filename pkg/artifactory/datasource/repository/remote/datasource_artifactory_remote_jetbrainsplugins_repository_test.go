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
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

// jetBrainsPluginsMarketplaceURL is the JetBrains Marketplace URL the backend
// declares as this package type's remote registry.
const jetBrainsPluginsMarketplaceURL = "https://plugins.jetbrains.com"

// TestAccDataSourceRemoteJetBrainsPluginsRepository creates a repository with
// the resource and reads it back through the data source, asserting parity
// across the non-secret fields the spec exposes.
func TestAccDataSourceRemoteJetBrainsPluginsRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("jetbrainsplugins-remote-ds", "artifactory_remote_jetbrainsplugins_repository")
	dsFqrn := fmt.Sprintf("data.artifactory_remote_jetbrainsplugins_repository.%s", name)

	config := util.ExecuteTemplate("jetbrainsplugins-remote-ds", `
		resource "artifactory_remote_jetbrainsplugins_repository" "{{ .name }}" {
			key                            = "{{ .name }}"
			url                            = "{{ .url }}"
			description                    = "JetBrains Plugins data source test"
			notes                          = "notes"
			excludes_pattern               = "excluded/**"
			hard_fail                      = true
			bypass_head_requests           = true
			list_remote_folder_items       = true
			retrieval_cache_period_seconds = 600
			socket_timeout_millis          = 20000
		}

		data "artifactory_remote_jetbrainsplugins_repository" "{{ .name }}" {
			key = artifactory_remote_jetbrainsplugins_repository.{{ .name }}.key
		}
	`, map[string]interface{}{
		"name": name,
		"url":  jetBrainsPluginsMarketplaceURL,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dsFqrn, "key", name),
					resource.TestCheckResourceAttr(dsFqrn, "package_type", repository.JetBrainsPluginsPackageType),
					resource.TestCheckResourceAttrPair(dsFqrn, "url", fqrn, "url"),
					resource.TestCheckResourceAttrPair(dsFqrn, "repo_layout_ref", fqrn, "repo_layout_ref"),
					resource.TestCheckResourceAttrPair(dsFqrn, "description", fqrn, "description"),
					resource.TestCheckResourceAttrPair(dsFqrn, "notes", fqrn, "notes"),
					resource.TestCheckResourceAttrPair(dsFqrn, "includes_pattern", fqrn, "includes_pattern"),
					resource.TestCheckResourceAttrPair(dsFqrn, "excludes_pattern", fqrn, "excludes_pattern"),
					resource.TestCheckResourceAttrPair(dsFqrn, "hard_fail", fqrn, "hard_fail"),
					resource.TestCheckResourceAttrPair(dsFqrn, "offline", fqrn, "offline"),
					resource.TestCheckResourceAttrPair(dsFqrn, "blacked_out", fqrn, "blacked_out"),
					resource.TestCheckResourceAttrPair(dsFqrn, "xray_index", fqrn, "xray_index"),
					resource.TestCheckResourceAttrPair(dsFqrn, "bypass_head_requests", fqrn, "bypass_head_requests"),
					resource.TestCheckResourceAttrPair(dsFqrn, "list_remote_folder_items", fqrn, "list_remote_folder_items"),
					resource.TestCheckResourceAttrPair(dsFqrn, "retrieval_cache_period_seconds", fqrn, "retrieval_cache_period_seconds"),
					resource.TestCheckResourceAttrPair(dsFqrn, "socket_timeout_millis", fqrn, "socket_timeout_millis"),
					resource.TestCheckResourceAttrPair(dsFqrn, "store_artifacts_locally", fqrn, "store_artifacts_locally"),
					// No password is configured on the underlying repo, so
					// the data source must not surface one in state either.
					// `password` remains an Optional attribute in the shared
					// basic-remote data-source schema; this only asserts the
					// no-secret-set path, not that the schema hides it.
					resource.TestCheckNoResourceAttr(dsFqrn, "password"),
					// `password_wo` is a resource-only write-only attribute
					// with no counterpart in the data-source schema.
					resource.TestCheckNoResourceAttr(dsFqrn, "password_wo"),
				),
			},
		},
	})
}

// TestAccDataSourceRemoteJetBrainsPluginsRepository_defaults reads back a
// minimally created repository and asserts the REST-verified defaults.
func TestAccDataSourceRemoteJetBrainsPluginsRepository_defaults(t *testing.T) {
	_, fqrn, name := testutil.MkNames("jetbrainsplugins-remote-ds-min", "artifactory_remote_jetbrainsplugins_repository")
	dsFqrn := fmt.Sprintf("data.artifactory_remote_jetbrainsplugins_repository.%s", name)

	config := util.ExecuteTemplate("jetbrainsplugins-remote-ds-min", `
		resource "artifactory_remote_jetbrainsplugins_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "{{ .url }}"
		}

		data "artifactory_remote_jetbrainsplugins_repository" "{{ .name }}" {
			key = artifactory_remote_jetbrainsplugins_repository.{{ .name }}.key
		}
	`, map[string]interface{}{
		"name": name,
		"url":  jetBrainsPluginsMarketplaceURL,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dsFqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(dsFqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(dsFqrn, "store_artifacts_locally", "true"),
					resource.TestCheckResourceAttr(dsFqrn, "retrieval_cache_period_seconds", "7200"),
					resource.TestCheckResourceAttr(dsFqrn, "missed_cache_period_seconds", "1800"),
					resource.TestCheckResourceAttr(dsFqrn, "assumed_offline_period_secs", "300"),
					resource.TestCheckResourceAttr(dsFqrn, "socket_timeout_millis", "15000"),
					resource.TestCheckResourceAttr(dsFqrn, "bypass_head_requests", "false"),
					resource.TestCheckResourceAttr(dsFqrn, "list_remote_folder_items", "false"),
					resource.TestCheckResourceAttr(dsFqrn, "offline", "false"),
					resource.TestCheckResourceAttr(dsFqrn, "hard_fail", "false"),
				),
			},
		},
	})
}
