package virtual_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func TestAccDataSourceVirtualAllGenericLikePackageTypes(t *testing.T) {
	for _, packageType := range virtual.PackageTypesLikeGeneric {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(mkNewVirtualTestCase(packageType, t, map[string]interface{}{
				"description": fmt.Sprintf("%s virtual repository public description testing.", packageType),
			}))
		})
	}
}

func TestAccDataSourceVirtualAllGenericLikeRetrievalPackageTypes(t *testing.T) {
	for _, packageType := range virtual.PackageTypesLikeGenericWithRetrievalCachePeriodSecs {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(mkNewVirtualTestCase(packageType, t, map[string]interface{}{
				"description":                    fmt.Sprintf("%s virtual repository public description testing.", packageType),
				"retrieval_cache_period_seconds": 650,
			}))
		})
	}
}

func TestAccDataSourceVirtualAllGradleLikePackageTypes(t *testing.T) {
	for _, packageType := range repository.PackageTypesLikeGradle {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(mkNewVirtualTestCase(packageType, t, map[string]interface{}{
				"description": fmt.Sprintf("%s virtual repository public description testing.", packageType),
				"pom_repository_references_cleanup_policy": "discard_active_reference",
			}))
		})
	}
}

func TestAccDataSourceVirtualAlpineRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.AlpinePackageType, t, map[string]interface{}{
		"description":                    "alpine virtual repository public description testing.",
		"retrieval_cache_period_seconds": 0,
	}))
}

func TestAccDataSourceVirtualBowerRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.BowerPackageType, t, map[string]interface{}{
		"description":                    "bower virtual repository public description testing.",
		"external_dependencies_enabled":  true,
		"external_dependencies_patterns": utilsdk.CastToInterfaceArr([]string{"**/github.com/**", "**/go.googlesource.com/**"}),
	}))
}

func TestAccDataSourceVirtualConanRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.ConanPackageType, t, map[string]interface{}{
		"description":                    "conan virtual repository public description testing.",
		"retrieval_cache_period_seconds": 60,
		"force_conan_authentication":     true,
	}))
}

func TestAccDataSourceVirtualDebianRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.DebianPackageType, t, map[string]interface{}{
		"description":                        "debian virtual repository public description testing.",
		"debian_default_architectures":       "i386,amd64",
		"retrieval_cache_period_seconds":     650,
		"optional_index_compression_formats": utilsdk.CastToInterfaceArr([]string{"bz2", "xz"}),
	}))
}

func TestAccDataSourceVirtualDockerRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.DockerPackageType, t, map[string]interface{}{
		"description":                      "docker virtual repository public description testing.",
		"resolve_docker_tags_by_timestamp": true,
	}))
}

func TestAccDataSourceVirtualGoRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.GoPackageType, t, map[string]interface{}{
		"description":                    "go virtual repository public description testing.",
		"external_dependencies_enabled":  true,
		"external_dependencies_patterns": utilsdk.CastToInterfaceArr([]string{"**/github.com/**", "**/go.googlesource.com/**"}),
	}))
}

func TestAccDataSourceVirtualHelmRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.HelmPackageType, t, map[string]interface{}{
		"description":                    "helm virtual repository public description testing.",
		"use_namespaces":                 true,
		"retrieval_cache_period_seconds": 650,
	}))
}

func TestAccDataSourceVirtualHelmOciRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.HelmOCIPackageType, t, map[string]interface{}{
		"description":                   "Helm OCI virtual repository public description testing.",
		"resolve_oci_tags_by_timestamp": true,
	}))
}

func TestAccDataSourceVirtualMavenRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.MavenPackageType, t, map[string]interface{}{
		"description":                "maven virtual repository public description testing.",
		"force_maven_authentication": true,
	}))
}

func TestAccDataSourceVirtualNpmRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.NPMPackageType, t, map[string]interface{}{
		"description":                    "npm virtual repository public description testing.",
		"external_dependencies_enabled":  true,
		"retrieval_cache_period_seconds": 650,
		"external_dependencies_patterns": utilsdk.CastToInterfaceArr([]string{"**/github.com/**", "**/go.googlesource.com/**"}),
	}))
}

func TestAccDataSourceVirtualNugetRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.NugetPackageType, t, map[string]interface{}{
		"description":                "nuget virtual repository public description testing.",
		"force_nuget_authentication": true,
		"artifactory_requests_can_retrieve_remote_artifacts": true,
	}))
}

func TestAccDataSourceVirtualOciRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(repository.OCIPackageType, t, map[string]interface{}{
		"description":                   "OCI virtual repository public description testing.",
		"resolve_oci_tags_by_timestamp": true,
	}))
}

func TestAccDataSourceVirtualRpmRepository(t *testing.T) {
	const packageType = "rpm"
	_, fqrn, name := testutil.MkNames("virtual-rpm-repo", "artifactory_virtual_rpm_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1-", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2-", "artifactory_keypair")
	virtualRepositoryBasic := util.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_keypair" "{{ .kp_name2 }}" {
			pair_name  = "{{ .kp_name2 }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id2 }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}
	
		resource "artifactory_virtual_rpm_repository" "{{ .repo_name }}" {
			key 	              = "{{ .repo_name }}"
			primary_keypair_ref   = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
	
			depends_on = [
				artifactory_keypair.{{ .kp_name }},
				artifactory_keypair.{{ .kp_name2 }},
			]
		}

		data "artifactory_virtual_rpm_repository" "{{ .repo_name }}" {
		    key = artifactory_virtual_rpm_repository.{{ .repo_name }}.id
		}

	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"kp_id2":      kpId2,
		"kp_name2":    kpName2,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_PGP_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_PGP_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "", security.VerifyKeyPair),
			acctest.VerifyDeleted(t, kpFqrn2, "", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", packageType),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef(virtual.Rclass, packageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func mkNewVirtualTestCase(packageType string, t *testing.T, extraFields map[string]interface{}) (*testing.T, resource.TestCase) {
	_, fqrn, name := testutil.MkNames(fmt.Sprintf("virtual-%s-repo-full-", packageType),
		fmt.Sprintf("artifactory_virtual_%s_repository", packageType))
	remoteRepoName := fmt.Sprintf("%s-remote", name)
	defaultFields := map[string]interface{}{
		"key":         name,
		"description": "A test virtual repo",
		"notes":       "Internal description",
	}
	allFields := utilsdk.MergeMaps(defaultFields, extraFields)
	allFieldsHcl := utilsdk.FmtMapToHcl(allFields)
	const virtualRepoFull = `
        resource "artifactory_remote_%[1]s_repository" "%[3]s" {
			key = "%[3]s"
            url = "https://tempurl.org"
			bypass_head_requests = true
		}

		resource "artifactory_virtual_%[1]s_repository" "%[2]s" {
%[4]s
            repositories = [artifactory_remote_%[1]s_repository.%[3]s.key]
		}

		data "artifactory_virtual_%[1]s_repository" "%[2]s" {
		    key = artifactory_virtual_%[1]s_repository.%[2]s.id
		}
	`
	extraChecks := testutil.MapToTestChecks(fqrn, extraFields)
	defaultChecks := testutil.MapToTestChecks(fqrn, allFields)

	checks := append(defaultChecks, extraChecks...)
	config := fmt.Sprintf(virtualRepoFull, packageType, name, remoteRepoName, allFieldsHcl)

	return t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:           config,
				Check:            resource.ComposeTestCheckFunc(checks...),
				ConfigPlanChecks: testutil.ConfigPlanChecks(""),
				// ConfigPlanChecks: resource.ConfigPlanChecks{
				// 	PostApplyPreRefresh: []plancheck.PlanCheck{
				// 		plancheck.ExpectEmptyPlan(),
				// 	},
				// },
			},
		},
	}
}

func TestAccDataSourceVirtualMissingRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("rpm-remote", "data.artifactory_virtual_rpm_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := util.ExecuteTemplate(
		"TestAccVirtualRpmRepository",
		`data "artifactory_virtual_rpm_repository" "{{ .name }}" {
			key = "non-existent-repo"
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(fqrn, "key"),
				),
			},
		},
	})
}
