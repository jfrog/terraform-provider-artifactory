package virtual_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
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
	for _, packageType := range repository.GradleLikePackageTypes {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(mkNewVirtualTestCase(packageType, t, map[string]interface{}{
				"description":                              fmt.Sprintf("%s virtual repository public description testing.", packageType),
				"force_maven_authentication":               true,
				"pom_repository_references_cleanup_policy": "discard_active_reference",
			}))
		})
	}
}

func TestAccDataSourceVirtualAlpineRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(virtual.AlpinePackageType, t, map[string]interface{}{
		"description":                    "alpine virtual repository public description testing.",
		"retrieval_cache_period_seconds": 0,
	}))
}

func TestAccDataSourceVirtualBowerRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(virtual.BowerPackageType, t, map[string]interface{}{
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
	resource.Test(mkNewVirtualTestCase(virtual.DebianPackageType, t, map[string]interface{}{
		"description":                        "bower virtual repository public description testing.",
		"debian_default_architectures":       "i386,amd64",
		"retrieval_cache_period_seconds":     650,
		"optional_index_compression_formats": utilsdk.CastToInterfaceArr([]string{"bz2", "xz"}),
	}))
}

func TestAccDataSourceVirtualDockerRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(virtual.DockerPackageType, t, map[string]interface{}{
		"description":                      "bower virtual repository public description testing.",
		"resolve_docker_tags_by_timestamp": true,
	}))
}

func TestAccDataSourceVirtualGoRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(virtual.GoPackageType, t, map[string]interface{}{
		"description":                    "go virtual repository public description testing.",
		"external_dependencies_enabled":  true,
		"external_dependencies_patterns": utilsdk.CastToInterfaceArr([]string{"**/github.com/**", "**/go.googlesource.com/**"}),
	}))
}

func TestAccDataSourceVirtualHelmRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(virtual.HelmPackageType, t, map[string]interface{}{
		"description":                    "helm virtual repository public description testing.",
		"use_namespaces":                 true,
		"retrieval_cache_period_seconds": 650,
	}))
}

func TestAccDataSourceVirtualNpmRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(virtual.NpmPackageType, t, map[string]interface{}{
		"description":                    "npm virtual repository public description testing.",
		"external_dependencies_enabled":  true,
		"retrieval_cache_period_seconds": 650,
		"external_dependencies_patterns": utilsdk.CastToInterfaceArr([]string{"**/github.com/**", "**/go.googlesource.com/**"}),
	}))
}

func TestAccDataSourceVirtualNugetRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase(virtual.NugetPackageType, t, map[string]interface{}{
		"description":                "nuget virtual repository public description testing.",
		"force_nuget_authentication": true,
		"artifactory_requests_can_retrieve_remote_artifacts": true,
	}))
}

func TestAccDataSourceVirtualRpmRepository(t *testing.T) {
	const packageType = "rpm"
	_, fqrn, name := testutil.MkNames("virtual-rpm-repo", "artifactory_virtual_rpm_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1-", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2-", "artifactory_keypair")
	virtualRepositoryBasic := utilsdk.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
-----BEGIN PGP PRIVATE KEY BLOCK-----
lIYEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib/+BwMCFjb4odY28+n0NWj7KZ53BkA0qzzqT9IpIfsW/tLNPTxYEFrDVbcF
1CuiAgAhyUfBEr9HQaMJBLfIIvo/B3nlWvwWHkiQFuWpsnJ2pj8F8LQqQ2hyaXN0
aWFuIEJvbmdpb3JubyA8Y2hyaXN0aWFuYkBqZnJvZy5jb20+iJoEExYKAEIWIQSS
w8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbAwUJA8JnAAULCQgHAgMiAgEGFQoJ
CAsCBBYCAwECHgcCF4AACgkQwL80hJIR2yRQDgD/X1t/hW9+uXdSY59FOClhQw/t
AzTYjDW+KLKadYJ3RAIBALD53rj7EnrXsSqv9Vqj3mJ7O38eXu50P57tD8ErpHMD
nIsEYYU7tRIKKwYBBAGXVQEFAQEHQCfT+jXHVkslGAJqVafoeWO8Nwz/oPPzNDJb
EOASsMRcAwEIB/4HAwK+Wi8OaidLuvQ6yknLUspoRL8KJlQu0JkfLxj6Wl6GrRtf
MdUBxaGUQX5UzMIqyYstgHKz2kBYvrJijWdOkkRuL82FySSh4yi/97FBikOBiHgE
GBYKACAWIQSSw8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbDAAKCRDAvzSEkhHb
JNR/AQCQjGWljmP8pYj6ohP8bOwVB4VE5qxjdfWQvBCUA0LFwgEAxLGVeT88pw3+
x7Cwd7SsuxlIOOCIJssFnUhA9Qsq2wE=
=qCzy
-----END PGP PRIVATE KEY BLOCK-----
EOF
			public_key = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----
mDMEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib+0KkNocmlzdGlhbiBCb25naW9ybm8gPGNocmlzdGlhbmJAamZyb2cuY29t
PoiaBBMWCgBCFiEEksPI7fvaXVQtxrbOwL80hJIR2yQFAmGFO7UCGwMFCQPCZwAF
CwkIBwIDIgIBBhUKCQgLAgQWAgMBAh4HAheAAAoJEMC/NISSEdskUA4A/19bf4Vv
frl3UmOfRTgpYUMP7QM02Iw1viiymnWCd0QCAQCw+d64+xJ617Eqr/Vao95iezt/
Hl7udD+e7Q/BK6RzA7g4BGGFO7USCisGAQQBl1UBBQEBB0An0/o1x1ZLJRgCalWn
6HljvDcM/6Dz8zQyWxDgErDEXAMBCAeIeAQYFgoAIBYhBJLDyO372l1ULca2zsC/
NISSEdskBQJhhTu1AhsMAAoJEMC/NISSEdsk1H8BAJCMZaWOY/yliPqiE/xs7BUH
hUTmrGN19ZC8EJQDQsXCAQDEsZV5PzynDf7HsLB3tKy7GUg44IgmywWdSED1Cyrb
AQ==
=2kMe
-----END PGP PUBLIC KEY BLOCK-----
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
-----BEGIN PGP PRIVATE KEY BLOCK-----
lIYEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib/+BwMCFjb4odY28+n0NWj7KZ53BkA0qzzqT9IpIfsW/tLNPTxYEFrDVbcF
1CuiAgAhyUfBEr9HQaMJBLfIIvo/B3nlWvwWHkiQFuWpsnJ2pj8F8LQqQ2hyaXN0
aWFuIEJvbmdpb3JubyA8Y2hyaXN0aWFuYkBqZnJvZy5jb20+iJoEExYKAEIWIQSS
w8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbAwUJA8JnAAULCQgHAgMiAgEGFQoJ
CAsCBBYCAwECHgcCF4AACgkQwL80hJIR2yRQDgD/X1t/hW9+uXdSY59FOClhQw/t
AzTYjDW+KLKadYJ3RAIBALD53rj7EnrXsSqv9Vqj3mJ7O38eXu50P57tD8ErpHMD
nIsEYYU7tRIKKwYBBAGXVQEFAQEHQCfT+jXHVkslGAJqVafoeWO8Nwz/oPPzNDJb
EOASsMRcAwEIB/4HAwK+Wi8OaidLuvQ6yknLUspoRL8KJlQu0JkfLxj6Wl6GrRtf
MdUBxaGUQX5UzMIqyYstgHKz2kBYvrJijWdOkkRuL82FySSh4yi/97FBikOBiHgE
GBYKACAWIQSSw8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbDAAKCRDAvzSEkhHb
JNR/AQCQjGWljmP8pYj6ohP8bOwVB4VE5qxjdfWQvBCUA0LFwgEAxLGVeT88pw3+
x7Cwd7SsuxlIOOCIJssFnUhA9Qsq2wE=
=qCzy
-----END PGP PRIVATE KEY BLOCK-----
EOF
			public_key = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----
mDMEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib+0KkNocmlzdGlhbiBCb25naW9ybm8gPGNocmlzdGlhbmJAamZyb2cuY29t
PoiaBBMWCgBCFiEEksPI7fvaXVQtxrbOwL80hJIR2yQFAmGFO7UCGwMFCQPCZwAF
CwkIBwIDIgIBBhUKCQgLAgQWAgMBAh4HAheAAAoJEMC/NISSEdskUA4A/19bf4Vv
frl3UmOfRTgpYUMP7QM02Iw1viiymnWCd0QCAQCw+d64+xJ617Eqr/Vao95iezt/
Hl7udD+e7Q/BK6RzA7g4BGGFO7USCisGAQQBl1UBBQEBB0An0/o1x1ZLJRgCalWn
6HljvDcM/6Dz8zQyWxDgErDEXAMBCAeIeAQYFgoAIBYhBJLDyO372l1ULca2zsC/
NISSEdskBQJhhTu1AhsMAAoJEMC/NISSEdsk1H8BAJCMZaWOY/yliPqiE/xs7BUH
hUTmrGN19ZC8EJQDQsXCAQDEsZV5PzynDf7HsLB3tKy7GUg44IgmywWdSED1Cyrb
AQ==
=2kMe
-----END PGP PUBLIC KEY BLOCK-----
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
		"kp_id":     kpId,
		"kp_name":   kpName,
		"kp_id2":    kpId2,
		"kp_name2":  kpName2,
		"repo_name": name,
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
			acctest.VerifyDeleted(kpFqrn, security.VerifyKeyPair),
			acctest.VerifyDeleted(kpFqrn2, security.VerifyKeyPair),
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
						r, _ := repository.GetDefaultRepoLayoutRef(virtual.Rclass, packageType)()
						return r.(string)
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
            repositories = ["%[3]s"]
            depends_on = [artifactory_remote_%[1]s_repository.%[3]s]
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	}
}
