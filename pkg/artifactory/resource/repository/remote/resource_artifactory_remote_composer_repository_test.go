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

func TestAccRemoteComposerRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase(repository.ComposerPackageType, t, map[string]interface{}{
		"url":                         "https://github.com/",
		"vcs_git_provider":            "GITHUB",
		"composer_registry_url":       "https://packagist1.org",
		"missed_cache_period_seconds": 1800,
	}))
}

func TestAccRemoteComposerRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-Composer-remote", "artifactory_remote_composer_repository")

	const temp = `
		resource "artifactory_remote_composer_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://github.com/"
			vcs_git_provider = "GITHUB"
			composer_registry_url = "https://packagist1.org"
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteComposerRepository_migrate_from_SDKv2", temp, params)

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
					resource.TestCheckResourceAttr(fqrn, "vcs_git_provider", "GITHUB"),
					resource.TestCheckResourceAttr(fqrn, "composer_registry_url", "https://packagist1.org"),
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
