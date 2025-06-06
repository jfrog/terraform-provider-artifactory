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

func TestAccLocalConanRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("conan-local", "artifactory_local_conan_repository")
	params := map[string]interface{}{
		"force_conan_authentication": testutil.RandBool(),
		"name":                       name,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalConanRepository", `
		resource "artifactory_local_conan_repository" "{{ .name }}" {
		  key                        = "{{ .name }}"
		  force_conan_authentication = {{ .force_conan_authentication }}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "force_conan_authentication", fmt.Sprintf("%t", params["force_conan_authentication"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.ConanPackageType)
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

func TestAccLocalConanRepository_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("conan-local", "artifactory_local_conan_repository")
	params := map[string]interface{}{
		"force_conan_authentication": testutil.RandBool(),
		"name":                       name,
	}
	config := util.ExecuteTemplate("TestAccLocalConanRepository", `
		resource "artifactory_local_conan_repository" "{{ .name }}" {
		  key                        = "{{ .name }}"
		  force_conan_authentication = {{ .force_conan_authentication }}
		}
	`, params)

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
