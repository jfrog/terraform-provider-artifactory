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
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalTerraformModuleRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("terraform-local", "artifactory_local_terraform_module_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := util.ExecuteTemplate(
		"TestAccLocalTerraformModuleRepository",
		`resource "artifactory_local_terraform_module_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "terraform-module-default"),
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

func TestAccLocalTerraformProviderRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("terraform-local", "artifactory_local_terraform_provider_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := util.ExecuteTemplate(
		"TestAccLocalTerraformProviderRepository",
		`resource "artifactory_local_terraform_provider_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "terraform-provider-default"),
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

func TestAccLocalTerraformProviderRepository_GPGKeyPair(t *testing.T) {
	_, fqrn, name := testutil.MkNames("terraform-local", "artifactory_local_terraform_provider_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")

	config := util.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_keypair" "{{ .kp_name2 }}" {
			pair_name  = "{{ .kp_name2 }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id2 }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_local_terraform_provider_repository" "{{ .repo_name }}" {
			key                   = "{{ .repo_name }}"
			primary_keypair_ref   = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
		}
	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"kp_id2":      kpId2,
		"kp_name2":    kpName2,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_PGP_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_PGP_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
			acctest.VerifyDeleted(t, kpFqrn2, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "terraform-provider-default"),
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

func TestAccLocalTerraformModuleRepository_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("terraform-local", "artifactory_local_terraform_module_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := util.ExecuteTemplate(
		"TestAccLocalTerraformModuleRepository",
		`resource "artifactory_local_terraform_module_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "12.8.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "id", name),
					resource.TestCheckResourceAttr(fqrn, "key", name),
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

func TestAccLocalTerraformProviderRepository_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("terraform-local", "artifactory_local_terraform_provider_repository")
	params := map[string]interface{}{
		"name": name,
	}
	config := util.ExecuteTemplate(
		"TestAccLocalTerraformProviderRepository",
		`resource "artifactory_local_terraform_provider_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "12.8.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "id", name),
					resource.TestCheckResourceAttr(fqrn, "key", name),
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
