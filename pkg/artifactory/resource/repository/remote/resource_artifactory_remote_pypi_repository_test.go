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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRemotePyPIRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase(repository.PyPiPackageType, t, map[string]interface{}{
		"pypi_registry_url":           "https://custom.PYPI.registry.url",
		"priority_resolution":         true,
		"missed_cache_period_seconds": 1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"curated":                     false,
		"pass_through":                false,
	}))
}

func TestAccRemotePyPIRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-pypi-remote", "artifactory_remote_pypi_repository")

	const temp = `
		resource "artifactory_remote_pypi_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://github.com/"
			pypi_registry_url = "https://custom.PYPI.registry.url"
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemotePyPIRepository_migrate_from_SDKv2", temp, params)

	resource.Test(t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.8.3",
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com/"),
					resource.TestCheckResourceAttr(fqrn, "pypi_registry_url", "https://custom.PYPI.registry.url"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
