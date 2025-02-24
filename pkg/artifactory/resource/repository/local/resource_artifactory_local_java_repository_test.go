package local_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

var commonJavaParams = map[string]interface{}{
	"name":                            "",
	"checksum_policy_type":            testutil.RandSelect("client-checksums", "server-generated-checksums"),
	"snapshot_version_behavior":       testutil.RandSelect("unique", "non-unique", "deployer"),
	"max_unique_snapshots":            testutil.RandSelect(0, 5, 10),
	"handle_releases":                 true,
	"handle_snapshots":                true,
	"suppress_pom_consistency_checks": false,
}

const localJavaRepositoryBasic = `
		resource "{{ .resource_name }}" "{{ .name }}" {
		  key                 			  = "{{ .name }}"
		  checksum_policy_type            = "{{ .checksum_policy_type }}"
		  snapshot_version_behavior       = "{{ .snapshot_version_behavior }}"
		  max_unique_snapshots            = {{ .max_unique_snapshots }}
		  handle_releases                 = {{ .handle_releases }}
		  handle_snapshots                = {{ .handle_snapshots }}
		  suppress_pom_consistency_checks = {{ .suppress_pom_consistency_checks }}
		}
	`

func TestAccLocalMavenRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("maven-local", "artifactory_local_maven_repository")
	tempStruct := utilsdk.MergeMaps(commonJavaParams)

	tempStruct["name"] = name
	tempStruct["resource_name"] = strings.Split(fqrn, ".")[0]
	tempStruct["suppress_pom_consistency_checks"] = false

	config := util.ExecuteTemplate(fqrn, localJavaRepositoryBasic, tempStruct)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "checksum_policy_type", fmt.Sprintf("%s", tempStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(fqrn, "snapshot_version_behavior", fmt.Sprintf("%s", tempStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", tempStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%v", tempStruct["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%v", tempStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", tempStruct["suppress_pom_consistency_checks"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "maven"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
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

func TestAccLocalMavenRepository_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("maven-local", "artifactory_local_maven_repository")
	tempStruct := utilsdk.MergeMaps(commonJavaParams)

	tempStruct["name"] = name
	tempStruct["resource_name"] = strings.Split(fqrn, ".")[0]
	tempStruct["suppress_pom_consistency_checks"] = false

	config := util.ExecuteTemplate(fqrn, localJavaRepositoryBasic, tempStruct)

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

func makeLocalGradleLikeRepoTestCase(packageType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("%s-local", packageType)
	resourceName := fmt.Sprintf("artifactory_local_%s_repository", packageType)
	_, fqrn, name := testutil.MkNames(name, resourceName)
	tempStruct := utilsdk.MergeMaps(commonJavaParams)

	tempStruct["name"] = name
	tempStruct["resource_name"] = strings.Split(fqrn, ".")[0]
	tempStruct["suppress_pom_consistency_checks"] = true

	config := util.ExecuteTemplate(fqrn, localJavaRepositoryBasic, tempStruct)

	return t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "checksum_policy_type", fmt.Sprintf("%s", tempStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(fqrn, "snapshot_version_behavior", fmt.Sprintf("%s", tempStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", tempStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%v", tempStruct["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%v", tempStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", tempStruct["suppress_pom_consistency_checks"])),
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
	}
}

func TestAccLocalGradleLikeRepository(t *testing.T) {
	for _, packageType := range repository.PackageTypesLikeGradle {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(makeLocalGradleLikeRepoTestCase(packageType, t))
		})
	}
}

func TestAccLocalGradleLikeRepository_UpgradeFromSDKv2(t *testing.T) {
	for _, packageType := range repository.PackageTypesLikeGradle {
		t.Run(packageType, func(t *testing.T) {
			name := fmt.Sprintf("%s-local", packageType)
			resourceName := fmt.Sprintf("artifactory_local_%s_repository", packageType)
			_, fqrn, name := testutil.MkNames(name, resourceName)
			tempStruct := utilsdk.MergeMaps(commonJavaParams)

			tempStruct["name"] = name
			tempStruct["resource_name"] = strings.Split(fqrn, ".")[0]
			tempStruct["suppress_pom_consistency_checks"] = true

			config := util.ExecuteTemplate(fqrn, localJavaRepositoryBasic, tempStruct)

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
		})
	}
}
