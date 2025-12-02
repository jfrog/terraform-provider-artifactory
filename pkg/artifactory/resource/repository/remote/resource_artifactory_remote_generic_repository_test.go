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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRemoteGenericRepository_full(t *testing.T) {
	const remoteGenericRepoBasicWithPropagate = `
		resource "artifactory_remote_generic_repository" "%s" {
			key                     		= "%s"
			description 					= "This is a test"
			url                     		= "https://registry.npmjs.org/"
			repo_layout_ref         		= "simple-default"
			propagate_query_params  		= true
			retrieve_sha256_from_server     = true
			retrieval_cache_period_seconds  = 70
		}
	`
	id := testutil.RandomInt()
	name := fmt.Sprintf("remote-test-repo-basic%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_generic_repository.%s", name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(remoteGenericRepoBasicWithPropagate, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "propagate_query_params", "true"),
					resource.TestCheckResourceAttr(fqrn, "retrieve_sha256_from_server", "true"),
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

func TestAccRemoteGenericRepository_migrate_to_schema_v4(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-generic-remote", "artifactory_remote_generic_repository")

	const temp = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
			key                     		= "{{ .name }}"
			description 					= "This is a test"
			url                     		= "https://registry.npmjs.org/"
			repo_layout_ref         		= "simple-default"
			propagate_query_params  		= true
			retrieval_cache_period_seconds  = 70
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteRepository_generic_migrate_to_schema_v4", temp, params)

	resource.Test(t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.0.0",
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckNoResourceAttr(fqrn, "retrieve_sha256_from_server"),
				),
			},
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.8.3",
					},
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccRemoteGenericRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-generic-remote", "artifactory_remote_generic_repository")

	const temp = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
			key                     		= "{{ .name }}"
			description 					= "This is a test"
			url                     		= "https://registry.npmjs.org/"
			repo_layout_ref         		= "simple-default"
			retrieve_sha256_from_server     = true
			retrieval_cache_period_seconds  = 70
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteRepository_generic_migrate_from_SDKv2", temp, params)

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
					resource.TestCheckResourceAttr(fqrn, "retrieve_sha256_from_server", "true"),
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
