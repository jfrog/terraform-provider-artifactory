package local_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalCargoRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("cargo-local", "artifactory_local_cargo_repository")
	params := map[string]interface{}{
		"anonymous_access":    testutil.RandBool(),
		"enable_sparse_index": testutil.RandBool(),
		"name":                name,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalCargoRepository", `
		resource "artifactory_local_cargo_repository" "{{ .name }}" {
		  key              = "{{ .name }}"
		  anonymous_access = {{ .anonymous_access }}
		  enable_sparse_index = {{ .enable_sparse_index }}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "anonymous_access", fmt.Sprintf("%t", params["anonymous_access"])),
					resource.TestCheckResourceAttr(fqrn, "enable_sparse_index", fmt.Sprintf("%t", params["enable_sparse_index"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.CargoPackageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
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

func TestAccLocalCargoRepository_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("cargo-local", "artifactory_local_cargo_repository")
	params := map[string]interface{}{
		"anonymous_access":    testutil.RandBool(),
		"enable_sparse_index": testutil.RandBool(),
		"name":                name,
	}
	config := util.ExecuteTemplate("TestAccLocalCargoRepository", `
		resource "artifactory_local_cargo_repository" "{{ .name }}" {
		  key              = "{{ .name }}"
		  anonymous_access = {{ .anonymous_access }}
		  enable_sparse_index = {{ .enable_sparse_index }}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
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
