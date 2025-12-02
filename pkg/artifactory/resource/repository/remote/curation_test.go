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

var curationPackageTypes = []string{
	repository.ConanPackageType,
	repository.DockerPackageType,
	repository.GemsPackageType,
	repository.GoPackageType,
	repository.GradlePackageType,
	repository.HuggingFacePackageType,
	repository.MavenPackageType,
	repository.NPMPackageType,
	repository.NugetPackageType,
	repository.PyPiPackageType,
}

func TestAccRemoteRepository_with_curated(t *testing.T) {
	for _, packageType := range curationPackageTypes {
		t.Run(packageType, func(t *testing.T) {
			rs := fmt.Sprintf("artifactory_remote_%s_repository", packageType)
			_, fqrn, resourceName := testutil.MkNames("test-remote-curated-repo", rs)

			const temp = `
				resource "artifactory_remote_{{ .package_type }}_repository" "{{ .name }}" {
					key                     		= "{{ .name }}"
					description 					= "This is a test"
					url                     		= "https://tempurl.org/"
					repo_layout_ref         		= "simple-default"
					curated                         = true
				}
			`

			testData := map[string]string{
				"name":         resourceName,
				"package_type": packageType,
			}

			config := util.ExecuteTemplate("TestAccRemoteRepository_with_curated", temp, testData)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(fqrn, "key", resourceName),
							resource.TestCheckResourceAttr(fqrn, "curated", "true"),
						),
					},
					{
						ResourceName:                         fqrn,
						ImportState:                          true,
						ImportStateId:                        resourceName,
						ImportStateVerify:                    true,
						ImportStateVerifyIdentifierAttribute: "key",
					},
				},
			})
		})
	}
}
