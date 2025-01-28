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

func TestAccRemoteNugetRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase(repository.NugetPackageType, t, map[string]interface{}{
		"url":                         "https://www.nuget.org/",
		"download_context_path":       "api/v2/package",
		"force_nuget_authentication":  true,
		"missed_cache_period_seconds": 1800,
		"symbol_server_url":           "https://symbols.nuget.org/download/symbols",
	}))
}

func TestAccRemoteNugetRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-nuget-remote", "artifactory_remote_nuget_repository")

	const temp = `
		resource "artifactory_remote_nuget_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://www.nuget.org/"
			download_context_path = "api/v2/package"
			force_nuget_authentication = true
			symbol_server_url = "https://symbols.nuget.org/download/symbols"
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteNugetRepository_migrate_from_SDKv2", temp, params)

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
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", "true"),
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
