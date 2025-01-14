package local_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
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
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "terraform-module-default"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
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
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "terraform-provider-default"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
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
