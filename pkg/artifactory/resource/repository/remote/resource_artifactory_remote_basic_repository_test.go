package remote_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRemoteLikeBasicRepository(t *testing.T) {
	for _, repoType := range remote.PackageTypesLikeBasic {
		t.Run(repoType, func(t *testing.T) {
			resource.Test(mkNewRemoteTestCase(repoType, t, map[string]interface{}{
				"missed_cache_period_seconds": 1800,
			}))
		})
	}
}

func TestAccRemoteLikeBasicRepository_migrate_from_SDKv2(t *testing.T) {
	for _, packageType := range remote.PackageTypesLikeBasic {
		t.Run(packageType, func(t *testing.T) {
			_, fqrn, name := testutil.MkNames(fmt.Sprintf("test-%s-remote", packageType), fmt.Sprintf("artifactory_remote_%s_repository", packageType))

			const temp = `
				resource "artifactory_remote_{{ .package_type }}_repository" "{{ .name }}" {
					key         = "{{ .name }}"
					description = "This is a test"
					url         = "https://tempurl.org/"
				}
			`

			params := map[string]interface{}{
				"name":         name,
				"package_type": packageType,
			}

			config := util.ExecuteTemplate("TestAccRemoteLikeBasicRepository_migrate_from_SDKv2", temp, params)

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
		})
	}
}
