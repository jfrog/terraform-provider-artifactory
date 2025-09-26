package remote_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRemoteTerraformRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase(repository.TerraformPackageType, t, map[string]interface{}{
		"url":                     "https://github.com/",
		"terraform_registry_url":  "https://registry.terraform.io",
		"terraform_providers_url": "https://releases.hashicorp.com",
		"repo_layout_ref":         "simple-default",
	}))
}

func TestAccRemoteTerraformRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-terraform-remote", "artifactory_remote_terraform_repository")

	const temp = `
		resource "artifactory_remote_terraform_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://github.com/"
			bypass_head_requests = true
			terraform_registry_url = "https://registry.terraform.io"
			terraform_providers_url = "https://releases.hashicorp.com"
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteTerraformRepository_migrate_from_SDKv2", temp, params)

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
					resource.TestCheckResourceAttr(fqrn, "terraform_registry_url", "https://registry.terraform.io"),
					resource.TestCheckResourceAttr(fqrn, "terraform_providers_url", "https://releases.hashicorp.com"),
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

func TestAccRemoteTerraformRepository_bypassHeadRequestsValidation(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-terraform-remote-validation", "artifactory_remote_terraform_repository")

	// Test case 1: bypass_head_requests = false with registry.terraform.io should fail validation
	const invalidConfig = `
		resource "artifactory_remote_terraform_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://github.com/"
			bypass_head_requests = false
			terraform_registry_url = "https://registry.terraform.io"
			terraform_providers_url = "https://releases.hashicorp.com"
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteTerraformRepository_bypassHeadRequestsValidation", invalidConfig, params)

	resource.Test(t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("For terraform registries.*bypass_head_requests must be set to true"),
			},
		},
	})
}

func TestAccRemoteTerraformRepository_bypassHeadRequestsValidationOpenTofu(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-terraform-remote-validation-opentofu", "artifactory_remote_terraform_repository")

	// Test case 2: bypass_head_requests = false with registry.opentofu.org should fail validation
	const invalidConfig = `
		resource "artifactory_remote_terraform_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://github.com/"
			bypass_head_requests = false
			terraform_registry_url = "https://registry.opentofu.org"
			terraform_providers_url = "https://releases.hashicorp.com"
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteTerraformRepository_bypassHeadRequestsValidationOpenTofu", invalidConfig, params)

	resource.Test(t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("For terraform registries.*bypass_head_requests must be set to true"),
			},
		},
	})
}
