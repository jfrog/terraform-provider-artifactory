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

func TestAccRemoteGradleRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase(repository.GradlePackageType, t, map[string]interface{}{
		"missed_cache_period_seconds":      1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"metadata_retrieval_timeout_secs":  30,   // https://github.com/jfrog/terraform-provider-artifactory/issues/509
		"fetch_jars_eagerly":               true,
		"fetch_sources_eagerly":            true,
		"remote_repo_checksum_policy_type": "fail",
		"handle_releases":                  true,
		"handle_snapshots":                 true,
		"suppress_pom_consistency_checks":  true,
		"reject_invalid_jars":              true,
		"max_unique_snapshots":             6,
		"curated":                          false,
		"pass_through":                     false,
	}))
}

func TestAccRemoteGradleRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-gradle-remote", "artifactory_remote_gradle_repository")

	const temp = `
		resource "artifactory_remote_gradle_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://github.com/"
			fetch_jars_eagerly =               true
			fetch_sources_eagerly =            true
			remote_repo_checksum_policy_type = "fail"
			handle_releases =                  true
			handle_snapshots =                 true
			suppress_pom_consistency_checks =  true
			reject_invalid_jars =              true
			max_unique_snapshots =             6
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteGradleRepository_migrate_from_SDKv2", temp, params)

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
					resource.TestCheckResourceAttr(fqrn, "fetch_jars_eagerly", "true"),
					resource.TestCheckResourceAttr(fqrn, "fetch_sources_eagerly", "true"),
					resource.TestCheckResourceAttr(fqrn, "remote_repo_checksum_policy_type", "fail"),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", "true"),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", "true"),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", "true"),
					resource.TestCheckResourceAttr(fqrn, "reject_invalid_jars", "true"),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", "6"),
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
