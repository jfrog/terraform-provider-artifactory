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

package repository_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRepositoryExcludes_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-repo-excludes-", "artifactory_repository_excludes")
	_, _, repoName := testutil.MkNames("test-docker-local-", "artifactory_local_docker_v2_repository")

	temp := `
	resource "artifactory_local_docker_v2_repository" "{{ .repoName }}" {
		key              = "{{ .repoName }}"
		excludes_pattern = "unmanaged/**"

		lifecycle {
			ignore_changes = [excludes_pattern]
		}
	}

	resource "artifactory_repository_excludes" "{{ .name }}" {
		repo_key = artifactory_local_docker_v2_repository.{{ .repoName }}.key
		patterns = [{{ range $i, $p := .patterns }}{{ if $i }}, {{ end }}"{{ $p }}"{{ end }}]
	}`

	config := util.ExecuteTemplate(name, temp, map[string]any{
		"name":     name,
		"repoName": repoName,
		"patterns": []string{"nginx/1.0.0/**", "alpine/**"},
	})
	updatedConfig := util.ExecuteTemplate(name, temp, map[string]any{
		"name":     name,
		"repoName": repoName,
		"patterns": []string{"nginx/2.0.0/**"},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", repoName),
					resource.TestCheckResourceAttr(fqrn, "patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "patterns.*", "nginx/1.0.0/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "patterns.*", "alpine/**"),
					testCheckRepoExcludesPattern(t, repoName, "unmanaged/**,nginx/1.0.0/**,alpine/**"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "patterns.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "patterns.*", "nginx/2.0.0/**"),
					testCheckRepoExcludesPattern(t, repoName, "unmanaged/**,nginx/2.0.0/**"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        repoName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "repo_key",
				ImportStateVerifyIgnore:              []string{"patterns"},
			},
		},
	})
}

func testCheckRepoExcludesPattern(t *testing.T, repoKey, expected string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		var repo struct {
			ExcludesPattern string `json:"excludesPattern"`
		}
		response, err := acctest.GetTestResty(t).R().
			SetPathParam("key", repoKey).
			SetResult(&repo).
			Get("artifactory/api/repositories/{key}")
		if err != nil {
			return err
		}
		if response.IsError() {
			return fmt.Errorf("%s", response.String())
		}
		if repo.ExcludesPattern != expected {
			return fmt.Errorf("expected excludesPattern %q, got %q", expected, repo.ExcludesPattern)
		}
		return nil
	}
}
