package local_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func TestAccDataSourceLocalAllPackageTypes(t *testing.T) {
	for _, packageType := range local.PackageTypesLikeGeneric {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(mkTestCase(packageType, t))
		})
	}
}

func mkTestCase(packageType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("local-%s-%d-full", packageType, testutil.RandomInt())
	resourceName := fmt.Sprintf("data.artifactory_local_%s_repository.%s", packageType, name)
	xrayIndex := testutil.RandBool()

	params := map[string]interface{}{
		"packageType": packageType,
		"name":        name,
		"xrayIndex":   xrayIndex,
	}
	config := util.ExecuteTemplate("TestAccLocalRepository", `
		resource "artifactory_local_{{ .packageType }}_repository" "{{ .name }}" {
			key         = "{{ .name }}"
			description = "Test repo for {{ .name }}"
			notes       = "Test repo for {{ .name }}"
			xray_index  = {{ .xrayIndex }}
		}

		data "artifactory_local_{{ .packageType }}_repository" "{{ .name }}" {
			key = artifactory_local_{{ .packageType }}_repository.{{ .name }}.id
		}
	`, params)

	return t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", packageType),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test repo for %s", name)),
				),
			},
		},
	}
}

func TestAccDataSourceLocalAlpineRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("alpine-local-test-repo-basic", "data.artifactory_local_alpine_repository")
	kpId, _, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")
	localRepositoryBasic := util.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "RSA"
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

		resource "artifactory_local_alpine_repository" "{{ .repo_name }}" {
			key 	     = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			depends_on = [artifactory_keypair.{{ .kp_name }}]
		}

		data "artifactory_local_alpine_repository" "{{ .repo_name }}" {
			key = artifactory_local_alpine_repository.{{ .repo_name }}.id
		}
	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "alpine"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "alpine"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceLocalAnsibleRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ansible-local-test-repo-basic", "data.artifactory_local_ansible_repository")
	kpId, _, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")
	localRepositoryBasic := util.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "RSA"
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

		resource "artifactory_local_ansible_repository" "{{ .repo_name }}" {
			key 	     = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		data "artifactory_local_ansible_repository" "{{ .repo_name }}" {
			key = artifactory_local_ansible_repository.{{ .repo_name }}.id
		}
	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "ansible"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "ansible"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceLocalCargoRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("cargo-local", "data.artifactory_local_cargo_repository")
	params := map[string]interface{}{
		"anonymous_access": testutil.RandBool(),
		"name":             name,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalCargoRepository", `
		resource "artifactory_local_cargo_repository" "{{ .name }}" {
		  key                 = "{{ .name }}"
		  anonymous_access = {{ .anonymous_access }}
		}

		data "artifactory_local_cargo_repository" "{{ .name }}" {
		  key = artifactory_local_cargo_repository.{{ .name }}.id
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "anonymous_access", fmt.Sprintf("%t", params["anonymous_access"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "cargo"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
					resource.TestCheckResourceAttr(fqrn, "enable_sparse_index", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceLocalConanRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("conan-local", "data.artifactory_local_conan_repository")
	params := map[string]interface{}{
		"force_conan_authentication": testutil.RandBool(),
		"name":                       name,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalConanRepository", `
		resource "artifactory_local_conan_repository" "{{ .name }}" {
			key                        = "{{ .name }}"
			force_conan_authentication = {{ .force_conan_authentication }}
		}

		data "artifactory_local_conan_repository" "{{ .name }}" {
			key = artifactory_local_conan_repository.{{ .name }}.id
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "force_conan_authentication", fmt.Sprintf("%t", params["force_conan_authentication"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "conan"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceLocalDebianRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("local-debian-repo", "data.artifactory_local_debian_repository")
	kpId, _, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, _, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")
	localRepositoryBasic := util.ExecuteTemplate("keypair", `
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

		resource "artifactory_local_debian_repository" "{{ .repo_name }}" {
			key 	     = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
			index_compression_formats = ["bz2","lzma","xz"]
			trivial_layout = true
		}

		data "artifactory_local_debian_repository" "{{ .repo_name }}" {
			key = artifactory_local_debian_repository.{{ .repo_name }}.id
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
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "debian"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "trivial_layout", "true"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.0", "bz2"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.1", "lzma"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.2", "xz"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "debian"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceLocalDockerV2Repository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("dockerv2-local", "data.artifactory_local_docker_v2_repository")
	params := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalDockerV2Repository", `
	resource "artifactory_local_docker_v2_repository" "{{ .name }}" {
	  key 	     = "{{ .name }}"
	  tag_retention = {{ .retention }}
	  max_unique_tags = {{ .max_tags }}
	  block_pushing_schema1 = {{ .block }}
	}

	data "artifactory_local_docker_v2_repository" "{{ .name }}" {
	  key = artifactory_local_docker_v2_repository.{{ .name }}.id
	}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", params["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "docker"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceLocalDockerV1Repository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("dockerv1-local", "data.artifactory_local_docker_v1_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalDockerv2Repository", `
	resource "artifactory_local_docker_v1_repository" "{{ .name }}" {
		key = "{{ .name }}"
	}
	data "artifactory_local_docker_v1_repository" "{{ .name }}" {
		key = artifactory_local_docker_v1_repository.{{ .name }}.id
	}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", "false"),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", "1"),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", "0"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "docker"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

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
		key                 			= "{{ .name }}"
		checksum_policy_type            = "{{ .checksum_policy_type }}"
		snapshot_version_behavior       = "{{ .snapshot_version_behavior }}"
		max_unique_snapshots            = {{ .max_unique_snapshots }}
		handle_releases                 = {{ .handle_releases }}
		handle_snapshots                = {{ .handle_snapshots }}
		suppress_pom_consistency_checks = {{ .suppress_pom_consistency_checks }}
	}

	data "{{ .resource_name }}" "{{ .name }}" {
		key = {{ .resource_name }}.{{ .name }}.id
	}
`

func TestAccDataSourceLocalMavenRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("maven-local", "data.artifactory_local_maven_repository")
	tempStruct := utilsdk.MergeMaps(commonJavaParams)

	tempStruct["name"] = name
	tempStruct["resource_name"] = "artifactory_local_maven_repository"
	tempStruct["suppress_pom_consistency_checks"] = false

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, localJavaRepositoryBasic, tempStruct),
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
		},
	})
}

func makeDataSourceLocalGradleLikeRepoTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("%s-local", repoType)
	resourceName := fmt.Sprintf("artifactory_local_%s_repository", repoType)
	_, tempFqrn, name := testutil.MkNames(name, resourceName)
	tempStruct := utilsdk.MergeMaps(commonJavaParams)

	tempStruct["name"] = name
	tempStruct["resource_name"] = resourceName
	tempStruct["suppress_pom_consistency_checks"] = true

	fqrn := "data." + tempFqrn

	return t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, localJavaRepositoryBasic, tempStruct),
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
		},
	}
}

func TestAccDataSourceLocalHelmOciRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("helmoci-local", "data.artifactory_local_helmoci_repository")
	params := map[string]interface{}{
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalHelmOciRepository", `
	resource "artifactory_local_helmoci_repository" "{{ .name }}" {
		key 	          = "{{ .name }}"
		tag_retention   = {{ .retention }}
		max_unique_tags = {{ .max_tags }}
	}

	data "artifactory_local_helmoci_repository" "{{ .name }}" {
		key = artifactory_local_helmoci_repository.{{ .name }}.id
	}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "helmoci"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceLocalAllGradleLikePackageTypes(t *testing.T) {
	for _, packageType := range repository.PackageTypesLikeGradle {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(makeDataSourceLocalGradleLikeRepoTestCase(packageType, t))
		})
	}
}

func TestAccDataSourceLocalNugetRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("nuget-local", "data.artifactory_local_nuget_repository")
	params := map[string]interface{}{
		"force_nuget_authentication": testutil.RandBool(),
		"max_unique_snapshots":       testutil.RandSelect(0, 5, 10),
		"name":                       name,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccDataSourceLocalNugetRepository", `
	resource "artifactory_local_nuget_repository" "{{ .name }}" {
		key                 = "{{ .name }}"
		max_unique_snapshots = {{ .max_unique_snapshots }}
		force_nuget_authentication = {{ .force_nuget_authentication }}
	}
	data "artifactory_local_nuget_repository" "{{ .name }}" {
		key = artifactory_local_nuget_repository.{{ .name }}.id
	}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", params["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", fmt.Sprintf("%t", params["force_nuget_authentication"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "nuget"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceLocalOciRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("oci-local", "data.artifactory_local_oci_repository")
	params := map[string]interface{}{
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalOciRepository", `
	resource "artifactory_local_oci_repository" "{{ .name }}" {
		key 	          = "{{ .name }}"
		tag_retention   = {{ .retention }}
		max_unique_tags = {{ .max_tags }}
	}

	data "artifactory_local_oci_repository" "{{ .name }}" {
		key = artifactory_local_oci_repository.{{ .name }}.id
	}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "oci"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceLocalRpmRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("local-rpm-repo", "data.artifactory_local_rpm_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")
	localRepositoryBasic := util.ExecuteTemplate("keypair", `
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

		resource "artifactory_local_rpm_repository" "{{ .repo_name }}" {
			key 	     = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
			yum_root_depth = 1
			enable_file_lists_indexing = true
			calculate_yum_metadata = true
			depends_on = [
				artifactory_keypair.{{ .kp_name }},
				artifactory_keypair.{{ .kp_name2 }},
			]
		}

		data "artifactory_local_rpm_repository" "{{ .repo_name }}" {
			key = artifactory_local_rpm_repository.{{ .repo_name }}.id
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
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "", security.VerifyKeyPair),
			acctest.VerifyDeleted(t, kpFqrn2, "", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "rpm"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "enable_file_lists_indexing", "true"),
					resource.TestCheckResourceAttr(fqrn, "calculate_yum_metadata", "true"),
					resource.TestCheckResourceAttr(fqrn, "yum_root_depth", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "rpm"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceLocalTerraformModuleRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("terraform-local", "data.artifactory_local_terraform_module_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := util.ExecuteTemplate(
		"TestAccLocalTerraformModuleRepository",
		`resource "artifactory_local_terraform_module_repository" "{{ .name }}" {
			key            = "{{ .name }}"
		}
		data "artifactory_local_terraform_module_repository" "{{ .name }}" {
			key = artifactory_local_terraform_module_repository.{{ .name }}.id
		}
		`,
		params,
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "terraform"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "terraform-module-default"),
				),
			},
		},
	})
}

func TestAccDataSourceLocalTerraformProviderRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("terraform-local", "data.artifactory_local_terraform_provider_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := util.ExecuteTemplate(
		"TestAccLocalTerraformProviderRepository",
		`resource "artifactory_local_terraform_provider_repository" "{{ .name }}" {
			key = "{{ .name }}"
		}
		data "artifactory_local_terraform_provider_repository" "{{ .name }}" {
			key = artifactory_local_terraform_provider_repository.{{ .name }}.id
		}
	`,
		params,
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "terraform"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "terraform-provider-default"),
				),
			},
		},
	})
}

func TestAccDataSourceLocalMissingRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("terraform-local", "data.artifactory_local_terraform_provider_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := util.ExecuteTemplate(
		"TestAccLocalTerraformProviderRepository",
		`data "artifactory_local_terraform_provider_repository" "{{ .name }}" {
			key = "non-existent-repo"
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
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
