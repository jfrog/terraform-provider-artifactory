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

func TestAccRemoteHelmOCIRepository(t *testing.T) {
	_, testCase := mkNewRemoteTestCase(repository.HelmOCIPackageType, t, map[string]interface{}{
		"external_dependencies_enabled":  true,
		"enable_token_authentication":    true,
		"external_dependencies_patterns": []interface{}{"**/hub.docker.io/**", "**/bintray.jfrog.io/**"},
	})
	resource.Test(t, testCase)
}

func TestAccRemoteHelmOCIRepository_migrate_from_SDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-helmoci-remote", "artifactory_remote_helmoci_repository")

	const temp = `
		resource "artifactory_remote_helmoci_repository" "{{ .name }}" {
			key = "{{ .name }}"
			url = "https://github.com/"
			external_dependencies_enabled = true
			enable_token_authentication = true
			external_dependencies_patterns = ["**/hub.docker.io/**", "**/bintray.jfrog.io/**"]
		}
	`

	params := map[string]interface{}{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteHelmOCIRepository_migrate_from_SDKv2", temp, params)

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
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "enable_token_authentication", "true"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/hub.docker.io/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/bintray.jfrog.io/**"),
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
