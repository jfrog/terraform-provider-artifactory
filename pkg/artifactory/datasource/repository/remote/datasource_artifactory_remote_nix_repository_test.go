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

func TestAccDataSourceRemoteNixRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("nix-remote-test-repo-basic", "data.artifactory_remote_nix_repository")

	cfg := util.ExecuteTemplate("nix-remote-ds", `
		resource "artifactory_remote_nix_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://cache.nixos.org"
		}

		data "artifactory_remote_nix_repository" "{{ .repo_name }}" {
			key = artifactory_remote_nix_repository.{{ .repo_name }}.key
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
					resource.TestCheckResourceAttr(fqrn, "package_type", "nix"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://cache.nixos.org"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("remote", repository.NixPackageType)
						return r
					}()),
				),
			},
		},
	})
}
