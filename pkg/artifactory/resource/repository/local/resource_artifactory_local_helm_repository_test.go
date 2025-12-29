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
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalHelmRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("helm-local", "artifactory_local_helm_repository")
	temp := `
		resource "artifactory_local_helm_repository" "{{ .name }}" {
			key = "{{ .name }}"
			force_non_duplicate_chart = {{ .force_non_duplicate_chart }}
			force_metadata_name_version = {{ .force_metadata_name_version }}
		}
	`

	params := map[string]interface{}{
		"force_non_duplicate_chart":   true,
		"force_metadata_name_version": true,
		"name":                        name,
	}
	config := util.ExecuteTemplate("TestAccLocalHelmRepository", temp, params)

	updatedParams := map[string]interface{}{
		"force_non_duplicate_chart":   false,
		"force_metadata_name_version": true,
		"name":                        name,
	}
	updatedConfig := util.ExecuteTemplate("TestAccLocalHelmRepository", temp, updatedParams)

	updatedParams2 := map[string]interface{}{
		"force_non_duplicate_chart":   true,
		"force_metadata_name_version": false,
		"name":                        name,
	}
	updatedConfig2 := util.ExecuteTemplate("TestAccLocalHelmRepository", temp, updatedParams2)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "force_non_duplicate_chart", fmt.Sprintf("%t", params["force_non_duplicate_chart"])),
					resource.TestCheckResourceAttr(fqrn, "force_metadata_name_version", fmt.Sprintf("%t", params["force_metadata_name_version"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.HelmPackageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config:      updatedConfig,
				ExpectError: regexp.MustCompile(`.*force_non_duplicate_chart cannot be updated after it is set.*`),
			},
			{
				Config:      updatedConfig2,
				ExpectError: regexp.MustCompile(`.*force_metadata_name_version cannot be updated after it is set.*`),
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
