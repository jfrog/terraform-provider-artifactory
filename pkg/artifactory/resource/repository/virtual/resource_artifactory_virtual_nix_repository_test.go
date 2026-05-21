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

package virtual_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccVirtualNixRepository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("nix-virtual-test-repo", "artifactory_virtual_nix_repository")
	localRepoName := testutil.RandSelect("nix-local-repo-1", "nix-local-repo-2", "nix-local-repo-3").(string)

	temp := `
		resource "artifactory_local_nix_repository" "{{ .local_repo_name }}" {
			key = "{{ .local_repo_name }}"
		}

		resource "artifactory_remote_nix_repository" "{{ .repo_name }}-remote" {
			key = "{{ .repo_name }}-remote"
			url = "https://cache.nixos.org"
		}

		resource "artifactory_virtual_nix_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			repositories = [
				artifactory_local_nix_repository.{{ .local_repo_name }}.key,
				artifactory_remote_nix_repository.{{ .repo_name }}-remote.key
			]
			description = "Test repository for Nix"
			notes       = "Internal notes"
			includes_pattern = "**/*"
			excludes_pattern   = ""
			artifactory_requests_can_retrieve_remote_artifacts = true
			depends_on = [
				artifactory_local_nix_repository.{{ .local_repo_name }},
				artifactory_remote_nix_repository.{{ .repo_name }}-remote
			]
		}
	`

	testData := map[string]interface{}{
		"repo_name":       name,
		"local_repo_name": localRepoName,
	}

	config := util.ExecuteTemplate("TestAccVirtualNixRepository_full", temp, testData)

	updatedTemp := `
		resource "artifactory_local_nix_repository" "{{ .local_repo_name }}" {
			key = "{{ .local_repo_name }}"
		}

		resource "artifactory_remote_nix_repository" "{{ .repo_name }}-remote" {
			key = "{{ .repo_name }}-remote"
			url = "https://cache.nixos.org"
		}

		resource "artifactory_virtual_nix_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			repositories = [
				artifactory_local_nix_repository.{{ .local_repo_name }}.key,
				artifactory_remote_nix_repository.{{ .repo_name }}-remote.key
			]
			description = "Updated description"
			notes       = "Updated notes"
			includes_pattern = "**/*"
			excludes_pattern   = "*.tmp"
			artifactory_requests_can_retrieve_remote_artifacts = false
			depends_on = [
				artifactory_local_nix_repository.{{ .local_repo_name }},
				artifactory_remote_nix_repository.{{ .repo_name }}-remote
			]
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccVirtualNixRepository_full", updatedTemp, testData)

	localFqrn := fmt.Sprintf("artifactory_local_nix_repository.%s", localRepoName)
	remoteFqrn := fmt.Sprintf("artifactory_remote_nix_repository.%s-remote", name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, localFqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, remoteFqrn, "key", acctest.CheckRepo),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "description", "Test repository for Nix"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", ""),
					resource.TestCheckResourceAttr(fqrn, "artifactory_requests_can_retrieve_remote_artifacts", "true"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("virtual", repository.NixPackageType)
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
					resource.TestCheckResourceAttr(fqrn, "artifactory_requests_can_retrieve_remote_artifacts", "false"),
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

func TestAccVirtualNixRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("nix-virtual", "artifactory_virtual_nix_repository")

	temp := `
		resource "artifactory_virtual_nix_repository" "{{ .repo_name }}" {
			key            = "{{ .repo_name }}"
			repositories   = []
		}
	`

	testData := map[string]interface{}{
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccVirtualNixRepository_basic", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef(virtual.Rclass, repository.NixPackageType)
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
