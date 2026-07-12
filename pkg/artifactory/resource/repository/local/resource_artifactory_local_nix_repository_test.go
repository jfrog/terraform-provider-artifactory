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

package local_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalNixRepository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("nix-local-test-repo", "artifactory_local_nix_repository")

	temp := `
		resource "artifactory_local_nix_repository" "{{ .repo_name }}" {
			key             = "{{ .repo_name }}"
			description     = "Test repository for Nix"
			notes           = "Internal notes"
			includes_pattern = "**/*"
			excludes_pattern   = ""
		}
	`

	testData := map[string]interface{}{
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccLocalNixRepository_full", temp, testData)

	updatedTemp := `
		resource "artifactory_local_nix_repository" "{{ .repo_name }}" {
			key             = "{{ .repo_name }}"
			description     = "Updated description"
			notes           = "Updated notes"
			includes_pattern = "**/*"
			excludes_pattern   = "*.tmp"
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccLocalNixRepository_full", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "description", "Test repository for Nix"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", ""),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.NixPackageType)
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

func TestAccLocalNixRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("nix-local", "artifactory_local_nix_repository")

	temp := `
		resource "artifactory_local_nix_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}
	`

	testData := map[string]interface{}{
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccLocalNixRepository_basic", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.NixPackageType)
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
			},
		},
	})
}
