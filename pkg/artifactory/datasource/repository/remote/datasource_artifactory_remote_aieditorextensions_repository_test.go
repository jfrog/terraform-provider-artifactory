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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccDataSourceRemoteAIEditorExtensionsRepository(t *testing.T) {
	t.Skip("Skipping: the 'aieditorextensions' package type is not supported on the CI Artifactory version, so repo creation fails with \"package type aieditorextensions is not supported\". Unrelated to this change (JTFPR-179).")
	_, fqrn, name := testutil.MkNames("aieditorextensions-remote-test-repo-basic", "data.artifactory_remote_aieditorextensions_repository")

	cfg := util.ExecuteTemplate("aieditorextensions-remote-ds", `
		resource "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key                         = "{{ .repo_name }}"
			url                         = "https://marketplace.visualstudio.com/_apis/public/gallery"
			enable_token_authentication = true
			propagate_query_params      = true
			retrieve_sha256_from_server = true
			curated                     = true
			pass_through                = true
		}

		data "artifactory_remote_aieditorextensions_repository" "{{ .repo_name }}" {
			key = artifactory_remote_aieditorextensions_repository.{{ .repo_name }}.key
		}
	`, map[string]interface{}{
		"repo_name": name,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "aieditorextensions"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://marketplace.visualstudio.com/_apis/public/gallery"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "**/**vsassets.io/**"),
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
		},
	})
}
