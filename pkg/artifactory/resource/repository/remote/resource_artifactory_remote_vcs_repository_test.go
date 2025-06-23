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

func TestAccRemoteVCSRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase(repository.VCSPackageType, t, map[string]interface{}{
		"url":                  "https://github.com/",
		"vcs_git_provider":     "CUSTOM",
		"vcs_git_download_url": "https://www.customrepo.com",
		"max_unique_snapshots": 5,
	}))
}

func TestAccRemoteVCSRepository_WithFormattedUrl(t *testing.T) {
	resource.Test(mkNewRemoteTestCase(repository.VCSPackageType, t, map[string]interface{}{
		"url":                  "https://github.com/",
		"vcs_git_provider":     "CUSTOM",
		"vcs_git_download_url": "{0}/{1}/+archive/{2}.{3}",
		"max_unique_snapshots": 5,
	}))
}

func TestAccRemoteVCSRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-vcs-remote", "artifactory_remote_vcs_repository")

	const temp = `
		resource "artifactory_remote_vcs_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://github.com/"
			vcs_git_provider = "CUSTOM"
			vcs_git_download_url = "https://www.customrepo.com"
			max_unique_snapshots = 5
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteVCSRepository_migrate_from_SDKv2", temp, params)

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
					resource.TestCheckResourceAttr(fqrn, "vcs_git_provider", "CUSTOM"),
					resource.TestCheckResourceAttr(fqrn, "vcs_git_download_url", "https://www.customrepo.com"),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", "5"),
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
