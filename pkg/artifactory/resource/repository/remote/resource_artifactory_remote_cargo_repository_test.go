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

func TestAccRemoteCargoRepository(t *testing.T) {
	_, testCase := mkNewRemoteTestCase(repository.CargoPackageType, t, map[string]interface{}{
		"git_registry_url":            "https://github.com/rust-lang/foo.index",
		"anonymous_access":            true,
		"enable_sparse_index":         true,
		"priority_resolution":         false,
		"missed_cache_period_seconds": 1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
	})
	resource.Test(t, testCase)
}

func TestAccRemoteCargoRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-cargo-remote", "artifactory_remote_cargo_repository")

	const temp = `
		resource "artifactory_remote_cargo_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://github.com/"
			git_registry_url = "https://github.com/rust-lang/foo.index"
			anonymous_access = true
			enable_sparse_index = true
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteCargoRepository_migrate_from_SDKv2", temp, params)

	resource.Test(t, resource.TestCase{
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.8.1",
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com/"),
					resource.TestCheckResourceAttr(fqrn, "git_registry_url", "https://github.com/rust-lang/foo.index"),
					resource.TestCheckResourceAttr(fqrn, "anonymous_access", "true"),
					resource.TestCheckResourceAttr(fqrn, "enable_sparse_index", "true"),
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
