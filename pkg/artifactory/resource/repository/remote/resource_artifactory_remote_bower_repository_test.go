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

func TestAccRemoteBowerRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase(repository.BowerPackageType, t, map[string]interface{}{
		"url":                         "https://github.com/",
		"vcs_git_provider":            "ARTIFACTORY",
		"bower_registry_url":          "https://registry1.bower.io",
		"missed_cache_period_seconds": 1800,
	}))
}

func TestAccRemoteBowerRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-bower-remote", "artifactory_remote_bower_repository")

	const temp = `
		resource "artifactory_remote_bower_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://github.com/"
			bower_registry_url = "https://registry1.bower.io"
			vcs_git_provider = "GITHUB"
			vcs_git_download_url = "https://github.com"
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteBowerRepository_migrate_from_SDKv2", temp, params)

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
					resource.TestCheckResourceAttr(fqrn, "bower_registry_url", "https://registry1.bower.io"),
					resource.TestCheckResourceAttr(fqrn, "vcs_git_provider", "GITHUB"),
					resource.TestCheckResourceAttr(fqrn, "vcs_git_download_url", "https://github.com"),
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
