// Copyright (c) JFrog Ltd. (2026)
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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

const nimModelRegistryURL = "https://api.ngc.nvidia.com"

// checkEnableTokenAuthenticationOnServer asserts the raw REST repo config's
// enableTokenAuthentication matches what Terraform state claims. This is the one
// attribute this package type adds, and its default (`true`) was only confirmed
// by probing a live instance rather than from the UI form, so the test asserts
// the server value directly rather than relying solely on Read/Import parity.
func checkEnableTokenAuthenticationOnServer(t *testing.T, fqrn string, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[fqrn]
		if !ok {
			return fmt.Errorf("%s not found in state", fqrn)
		}
		key := rs.Primary.Attributes["key"]

		var repo map[string]interface{}
		resp, err := acctest.GetTestResty(t).R().SetResult(&repo).Get("artifactory/api/repositories/" + key)
		if err != nil {
			return err
		}
		if resp.IsError() {
			return fmt.Errorf("REST GET %s failed: %s", key, resp.Status())
		}

		got := fmt.Sprintf("%v", repo["enableTokenAuthentication"])
		if got != want {
			return fmt.Errorf("enableTokenAuthentication: REST=%q, want=%q", got, want)
		}
		return nil
	}
}

func TestAccRemoteNimModelRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("nimmodel-remote", "artifactory_remote_nimmodel_repository")

	config := util.ExecuteTemplate("TestAccRemoteNimModelRepository_basic", `
		resource "artifactory_remote_nimmodel_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "{{ .url }}"
		}
	`, map[string]interface{}{
		"repo_name": name,
		"url":       nimModelRegistryURL,
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", nimModelRegistryURL),
					resource.TestCheckResourceAttr(fqrn, "list_remote_folder_items", "false"),
					// Artifactory defaults this to true for NimModel repositories,
					// unlike most remote package types which default it to false.
					resource.TestCheckResourceAttr(fqrn, "enable_token_authentication", "true"),
					checkEnableTokenAuthenticationOnServer(t, fqrn, "true"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("remote", repository.NimModelPackageType)
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

func TestAccRemoteNimModelRepository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("nimmodel-remote-test-repo", "artifactory_remote_nimmodel_repository")

	testData := map[string]interface{}{
		"repo_name": name,
		"url":       nimModelRegistryURL,
	}

	config := util.ExecuteTemplate("TestAccRemoteNimModelRepository_full", `
		resource "artifactory_remote_nimmodel_repository" "{{ .repo_name }}" {
			key                            = "{{ .repo_name }}"
			url                            = "{{ .url }}"
			description                    = "NVIDIA NIM models proxy"
			notes                          = "Internal notes"
			includes_pattern               = "**/*"
			excludes_pattern               = ""
			list_remote_folder_items       = false
			socket_timeout_millis          = 15000
			retrieval_cache_period_seconds = 7200
			xray_index                     = false
			enable_token_authentication    = true
		}
	`, testData)

	updatedConfig := util.ExecuteTemplate("TestAccRemoteNimModelRepository_full", `
		resource "artifactory_remote_nimmodel_repository" "{{ .repo_name }}" {
			key                            = "{{ .repo_name }}"
			url                            = "{{ .url }}"
			description                    = "Updated description"
			notes                          = "Updated notes"
			includes_pattern               = "**/*"
			excludes_pattern               = "*.tmp"
			list_remote_folder_items       = true
			socket_timeout_millis          = 30000
			retrieval_cache_period_seconds = 3600
			enable_token_authentication    = false
		}
	`, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", nimModelRegistryURL),
					resource.TestCheckResourceAttr(fqrn, "description", "NVIDIA NIM models proxy"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", ""),
					resource.TestCheckResourceAttr(fqrn, "socket_timeout_millis", "15000"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "7200"),
					resource.TestCheckResourceAttr(fqrn, "xray_index", "false"),
					resource.TestCheckResourceAttr(fqrn, "enable_token_authentication", "true"),
					checkEnableTokenAuthenticationOnServer(t, fqrn, "true"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("remote", repository.NimModelPackageType)
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
					resource.TestCheckResourceAttr(fqrn, "list_remote_folder_items", "true"),
					resource.TestCheckResourceAttr(fqrn, "socket_timeout_millis", "30000"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "3600"),
					// Must survive a round trip back to false; a server that ignored
					// the update would leave this at the true default.
					resource.TestCheckResourceAttr(fqrn, "enable_token_authentication", "false"),
					checkEnableTokenAuthenticationOnServer(t, fqrn, "false"),
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

// Artifactory applies no default registry URL for this package type, so `url`
// must stay required, and the resource's http/https validator inherited from
// the base remote schema still applies.
func TestAccRemoteNimModelRepository_url_is_required(t *testing.T) {
	_, _, name := testutil.MkNames("nimmodel-remote-nourl", "artifactory_remote_nimmodel_repository")

	missingURL := util.ExecuteTemplate("missing_url", `
		resource "artifactory_remote_nimmodel_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}
	`, map[string]interface{}{"repo_name": name})

	badScheme := util.ExecuteTemplate("bad_scheme", `
		resource "artifactory_remote_nimmodel_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "ftp://api.ngc.nvidia.com"
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

// Project assignment is inherited base behaviour, verified here because it had
// no coverage for this package type.
func TestAccRemoteNimModelRepository_with_project(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	// Artifactory requires a project-assigned repository key to be prefixed with
	// the project key and a dash.
	_, fqrn, name := testutil.MkNames(fmt.Sprintf("%s-nimmodel-remote", projectKey), "artifactory_remote_nimmodel_repository")

	testData := map[string]interface{}{
		"repo_name":   name,
		"url":         nimModelRegistryURL,
		"project_key": projectKey,
	}

	config := util.ExecuteTemplate("with_project", `
		resource "artifactory_remote_nimmodel_repository" "{{ .repo_name }}" {
			key                  = "{{ .repo_name }}"
			url                  = "{{ .url }}"
			project_key          = "{{ .project_key }}"
			project_environments = ["DEV", "PROD"]
		}
	`, testData)

	updatedConfig := util.ExecuteTemplate("with_project_updated", `
		resource "artifactory_remote_nimmodel_repository" "{{ .repo_name }}" {
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
func TestAccRemoteNimModelRepository_other_classes_not_supported(t *testing.T) {
	for _, class := range []string{"local", "virtual", "federated"} {
		t.Run(class, func(t *testing.T) {
			resourceType := fmt.Sprintf("artifactory_%s_nimmodel_repository", class)
			_, _, name := testutil.MkNames("nimmodel-"+class, resourceType)

			config := util.ExecuteTemplate("TestAccRemoteNimModelRepository_other_classes_not_supported", `
				resource "{{ .resource_type }}" "{{ .repo_name }}" {
					key = "{{ .repo_name }}"
				}
			`, map[string]interface{}{
				"resource_type": resourceType,
				"repo_name":     name,
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
		})
	}
}
