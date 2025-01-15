package federated_test

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func skipFederatedRepo() (bool, string) {
	if len(os.Getenv("ARTIFACTORY_URL_2")) > 0 {
		return false, "Env var `ARTIFACTORY_URL_2` is set. Executing testutil."
	}

	return true, "Env var `ARTIFACTORY_URL_2` is not set. Skipping testutil."
}

func federatedTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	if skip, reason := skipFederatedRepo(); skip {
		t.Skip(reason)
	}

	name := fmt.Sprintf("federated-%s-%d", repoType, rand.Int())
	resourceType := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
	resourceName := fmt.Sprintf("data.%s.%s", resourceType, name)
	xrayIndex := testutil.RandBool()
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	params := map[string]interface{}{
		"resourceType": resourceType,
		"name":         name,
		"xrayIndex":    xrayIndex,
		"memberUrl":    federatedMemberUrl,
	}

	repoTypeAdjusted := local.GetPackageType(repoType)

	federatedRepositoryConfig := util.ExecuteTemplate("TestAccFederatedRepositoryConfig", `
		resource "{{ .resourceType }}" "{{ .name }}" {
			key         = "{{ .name }}"
			description = "Test federated repo for {{ .name }}"
			notes       = "Test federated repo for {{ .name }}"
			xray_index  = {{ .xrayIndex }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
		data "{{ .resourceType }}" "{{ .name }}" {
			key = {{ .resourceType }}.{{ .name }}.id
		}
	`, params)

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, resourceName, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", repoTypeAdjusted),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test federated repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "notes", fmt.Sprintf("Test federated repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "xray_index", fmt.Sprintf("%t", xrayIndex)),

					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member.0.url", federatedMemberUrl),
					resource.TestCheckResourceAttr(resourceName, "member.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", repoType); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	}
}

func TestAccDataSourceFederatedRepoGenericTypes(t *testing.T) {
	for _, packageType := range federated.PackageTypesLikeGeneric {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(federatedTestCase(packageType, t))
		})
	}
}

func TestAccDataSourceFederatedAlpineRepository(t *testing.T) {
	_, tempFqrn, name := testutil.MkNames("alpine-federated", "artifactory_federated_alpine_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	federatedRepositoryBasic := util.ExecuteTemplate("keypair", `
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
		resource "artifactory_federated_alpine_repository" "{{ .repo_name }}" {
			key 	            = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}

			depends_on = [artifactory_keypair.{{ .kp_name }}]
		}
		data "artifactory_federated_alpine_repository" "{{ .repo_name }}" {
			key = artifactory_federated_alpine_repository.{{ .repo_name }}.id
		}
	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"memberUrl":   federatedMemberUrl,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	fqrn := "data." + tempFqrn

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "alpine"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "alpine"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceFederatedAnsibleRepository(t *testing.T) {
	_, tempFqrn, name := testutil.MkNames("ansible-federated", "artifactory_federated_ansible_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	federatedRepositoryBasic := util.ExecuteTemplate("keypair", `
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
		resource "artifactory_federated_ansible_repository" "{{ .repo_name }}" {
			key 	            = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}

		data "artifactory_federated_ansible_repository" "{{ .repo_name }}" {
			key = artifactory_federated_ansible_repository.{{ .repo_name }}.id
		}
	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"memberUrl":   federatedMemberUrl,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	fqrn := "data." + tempFqrn

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "ansible"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "ansible"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceFederatedCargoRepository(t *testing.T) {
	_, tempFqrn, name := testutil.MkNames("cargo-federated", "artifactory_federated_cargo_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)
	anonAccess := testutil.RandBool()
	enabledSparseIndex := testutil.RandBool()

	params := map[string]interface{}{
		"anonymous_access":    anonAccess,
		"enable_sparse_index": enabledSparseIndex,
		"name":                name,
		"memberUrl":           federatedMemberUrl,
	}

	template := `
		resource "artifactory_federated_cargo_repository" "{{ .name }}" {
			key                 = "{{ .name }}"
			anonymous_access    = {{ .anonymous_access }}
			enable_sparse_index = {{ .enable_sparse_index }}
			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
		data "artifactory_federated_cargo_repository" "{{ .name }}" {
			key = artifactory_federated_cargo_repository.{{ .name }}.id
		}
	`
	fqrn := "data." + tempFqrn

	federatedRepositoryBasic := util.ExecuteTemplate("TestAccFederatedCargoRepository", template, params)
	federatedRepositoryUpdated := util.ExecuteTemplate(
		"TestAccFederatedCargoRepository",
		template,
		map[string]interface{}{
			"anonymous_access":    !anonAccess,
			"enable_sparse_index": !enabledSparseIndex,
			"name":                name,
			"memberUrl":           federatedMemberUrl,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "anonymous_access", fmt.Sprintf("%t", anonAccess)),
					resource.TestCheckResourceAttr(fqrn, "enable_sparse_index", fmt.Sprintf("%t", enabledSparseIndex)),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "cargo"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "anonymous_access", fmt.Sprintf("%t", !anonAccess)),
					resource.TestCheckResourceAttr(fqrn, "enable_sparse_index", fmt.Sprintf("%t", !enabledSparseIndex)),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "cargo"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceFederatedConanRepository(t *testing.T) {
	_, tempFqrn, name := testutil.MkNames("conan-federated", "artifactory_federated_conan_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)
	forceConanAuthentication := testutil.RandBool()

	params := map[string]interface{}{
		"force_conan_authentication": forceConanAuthentication,
		"name":                       name,
		"memberUrl":                  federatedMemberUrl,
	}

	template := `
		resource "artifactory_federated_conan_repository" "{{ .name }}" {
			key                        = "{{ .name }}"
			force_conan_authentication = {{ .force_conan_authentication }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}

		data "artifactory_federated_conan_repository" "{{ .name }}" {
			key = artifactory_federated_conan_repository.{{ .name }}.id
		}
	`
	fqrn := "data." + tempFqrn

	federatedRepositoryBasic := util.ExecuteTemplate("TestAccFederatedConanRepository", template, params)

	federatedRepositoryUpdated := util.ExecuteTemplate(
		"TestAccFederatedConanRepository",
		template,
		map[string]interface{}{
			"force_conan_authentication": !forceConanAuthentication,
			"name":                       name,
			"memberUrl":                  federatedMemberUrl,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "force_conan_authentication", fmt.Sprintf("%t", forceConanAuthentication)),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "conan"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "force_conan_authentication", fmt.Sprintf("%t", !forceConanAuthentication)),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "conan"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceFederatedDebianRepository(t *testing.T) {
	_, tempFqrn, name := testutil.MkNames("debian-federated", "artifactory_federated_debian_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")

	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
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
		resource "artifactory_federated_debian_repository" "{{ .repo_name }}" {
			key 	                  = "{{ .repo_name }}"
			primary_keypair_ref       = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref     = artifactory_keypair.{{ .kp_name2 }}.pair_name
			index_compression_formats = ["bz2","lzma","xz"]
			trivial_layout            = {{ .trivialLayout }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}

			depends_on = [
				artifactory_keypair.{{ .kp_name }},
				artifactory_keypair.{{ .kp_name2 }},
			]
		}
		data "artifactory_federated_debian_repository" "{{ .repo_name }}" {
			key = artifactory_federated_debian_repository.{{ .repo_name }}.id
		}
	`

	federatedRepositoryBasic := util.ExecuteTemplate("keypair", template, map[string]interface{}{
		"kp_id":         kpId,
		"kp_name":       kpName,
		"kp_id2":        kpId2,
		"kp_name2":      kpName2,
		"repo_name":     name,
		"trivialLayout": true,
		"memberUrl":     federatedMemberUrl,
		"private_key":   os.Getenv("JFROG_TEST_PGP_PRIVATE_KEY"),
		"public_key":    os.Getenv("JFROG_TEST_PGP_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	federatedRepositoryUpdated := util.ExecuteTemplate("keypair", template, map[string]interface{}{
		"kp_id":         kpId,
		"kp_name":       kpName,
		"kp_id2":        kpId2,
		"kp_name2":      kpName2,
		"repo_name":     name,
		"trivialLayout": false,
		"memberUrl":     federatedMemberUrl,
		"private_key":   os.Getenv("JFROG_TEST_PGP_PRIVATE_KEY"),
		"public_key":    os.Getenv("JFROG_TEST_PGP_PUBLIC_KEY"),
	})

	fqrn := "data." + tempFqrn

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "", security.VerifyKeyPair),
			acctest.VerifyDeleted(t, kpFqrn2, "", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "debian"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "trivial_layout", "true"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.0", "bz2"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.1", "lzma"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.2", "xz"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "debian"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "debian"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "trivial_layout", "false"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.0", "bz2"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.1", "lzma"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.2", "xz"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "debian"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceFederatedDockerV2Repository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("docker-federated", "data.artifactory_federated_docker_v2_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_federated_docker_v2_repository" "{{ .name }}" {
			key 	              = "{{ .name }}"
			tag_retention         = {{ .retention }}
			max_unique_tags       = {{ .max_tags }}
			block_pushing_schema1 = {{ .block }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
		data "artifactory_federated_docker_v2_repository" "{{ .name }}" {
			key = artifactory_federated_docker_v2_repository.{{ .name }}.id
		}
	`

	params := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryBasic := util.ExecuteTemplate("TestAccFederatedDockerRepository", template, params)

	updated := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryUpdated := util.ExecuteTemplate("TestAccFederatedDockerRepository", template, updated)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", params["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "docker"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", updated["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", updated["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", updated["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "docker"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

// TestAccFederatedDockerRepository tests for backward compatibility
func TestAccDataSourceFederatedDockerRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("docker-federated", "data.artifactory_federated_docker_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_federated_docker_repository" "{{ .name }}" {
			key 	              = "{{ .name }}"
			tag_retention         = {{ .retention }}
			max_unique_tags       = {{ .max_tags }}
			block_pushing_schema1 = {{ .block }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
		data "artifactory_federated_docker_repository" "{{ .name }}" {
			key = artifactory_federated_docker_repository.{{ .name }}.id
		}
	`

	params := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryBasic := util.ExecuteTemplate("TestAccFederatedDockerRepository", template, params)

	updated := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryUpdated := util.ExecuteTemplate("TestAccFederatedDockerRepository", template, updated)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", params["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "docker"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", updated["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", updated["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", updated["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "docker"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceFederatedDockerV1Repository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("docker-federated", "data.artifactory_federated_docker_v1_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_federated_docker_v1_repository" "{{ .name }}" {
			key = "{{ .name }}"

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
		data "artifactory_federated_docker_v1_repository" "{{ .name }}" {
			key = artifactory_federated_docker_v1_repository.{{ .name }}.id
		}
	`

	params := map[string]interface{}{
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryBasic := util.ExecuteTemplate("TestAccFederatedDockerRepository", template, params)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", "false"),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", "1"),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", "0"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "docker"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
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

const federatedJavaRepositoryBasic = `
	resource "{{ .resource_name }}" "{{ .name }}" {
		key                             = "{{ .name }}"
		checksum_policy_type            = "{{ .checksum_policy_type }}"
		snapshot_version_behavior       = "{{ .snapshot_version_behavior }}"
		max_unique_snapshots            = {{ .max_unique_snapshots }}
		handle_releases                 = {{ .handle_releases }}
		handle_snapshots                = {{ .handle_snapshots }}
		suppress_pom_consistency_checks = {{ .suppress_pom_consistency_checks }}
		member {
			url     = "{{ .memberUrl }}"
			enabled = true
		}
	}
	data "{{ .resource_name }}" "{{ .name }}" {
		key = {{ .resource_name }}.{{ .name }}.id
	}
`

func TestAccDataSourceFederatedMavenRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("maven-federated", "artifactory_federated_maven_repository")

	repoLayoutRef := func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "maven"); return r }()
	tempStruct := utilsdk.MergeMaps(commonJavaParams)
	tempStruct["name"] = name
	tempStruct["resource_name"] = strings.Split(fqrn, ".")[0]
	tempStruct["suppress_pom_consistency_checks"] = false
	tempStruct["memberUrl"] = fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	updatedStruct := tempStruct
	updatedStruct["snapshot_version_behavior"] = "non-unique"
	updatedStruct["handle_releases"] = false
	updatedStruct["handle_snapshots"] = false
	updatedStruct["suppress_pom_consistency_checks"] = true

	dataFqrn := "data." + fqrn
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, federatedJavaRepositoryBasic, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataFqrn, "key", name),
					resource.TestCheckResourceAttr(dataFqrn, "checksum_policy_type", fmt.Sprintf("%s", tempStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(dataFqrn, "snapshot_version_behavior", fmt.Sprintf("%s", tempStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(dataFqrn, "max_unique_snapshots", fmt.Sprintf("%d", tempStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(dataFqrn, "handle_releases", fmt.Sprintf("%v", tempStruct["handle_releases"])),
					resource.TestCheckResourceAttr(dataFqrn, "handle_snapshots", fmt.Sprintf("%v", tempStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(dataFqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", tempStruct["suppress_pom_consistency_checks"])),
					resource.TestCheckResourceAttr(dataFqrn, "repo_layout_ref", repoLayoutRef), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: util.ExecuteTemplate(fqrn, federatedJavaRepositoryBasic, updatedStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataFqrn, "key", name),
					resource.TestCheckResourceAttr(dataFqrn, "checksum_policy_type", fmt.Sprintf("%s", updatedStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(dataFqrn, "snapshot_version_behavior", fmt.Sprintf("%s", updatedStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(dataFqrn, "max_unique_snapshots", fmt.Sprintf("%d", updatedStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(dataFqrn, "handle_releases", fmt.Sprintf("%v", updatedStruct["handle_releases"])),
					resource.TestCheckResourceAttr(dataFqrn, "handle_snapshots", fmt.Sprintf("%v", updatedStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(dataFqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", updatedStruct["suppress_pom_consistency_checks"])),
					resource.TestCheckResourceAttr(dataFqrn, "repo_layout_ref", repoLayoutRef), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func makeFederatedGradleLikeRepoTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("%s-federated", repoType)
	resourceName := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
	_, resourceFqrn, name := testutil.MkNames(name, resourceName)
	tempStruct := utilsdk.MergeMaps(commonJavaParams)

	tempStruct["name"] = name
	tempStruct["resource_name"] = strings.Split(resourceFqrn, ".")[0]
	tempStruct["suppress_pom_consistency_checks"] = true
	tempStruct["memberUrl"] = fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	updatedStruct := tempStruct
	updatedStruct["snapshot_version_behavior"] = "non-unique"
	updatedStruct["handle_releases"] = false
	updatedStruct["handle_snapshots"] = false
	updatedStruct["suppress_pom_consistency_checks"] = true

	fqrn := "data." + resourceFqrn

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, resourceFqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(resourceFqrn, federatedJavaRepositoryBasic, tempStruct),
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
				Config: util.ExecuteTemplate(resourceFqrn, federatedJavaRepositoryBasic, updatedStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "checksum_policy_type", fmt.Sprintf("%s", updatedStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(fqrn, "snapshot_version_behavior", fmt.Sprintf("%s", updatedStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", updatedStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%v", updatedStruct["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%v", updatedStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", updatedStruct["suppress_pom_consistency_checks"])),
				),
			},
		},
	}
}

func TestAccDataSourceFederatedAllGradleLikePackageTypes(t *testing.T) {
	for _, packageType := range repository.PackageTypesLikeGradle {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(makeFederatedGradleLikeRepoTestCase(packageType, t))
		})
	}
}

func TestAccDataSourceFederatedHelmOciRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("helmoci-federated", "data.artifactory_federated_helmoci_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_federated_helmoci_repository" "{{ .name }}" {
			key 	        = "{{ .name }}"
			tag_retention   = {{ .retention }}
			max_unique_tags = {{ .max_tags }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
		data "artifactory_federated_helmoci_repository" "{{ .name }}" {
			key = artifactory_federated_helmoci_repository.{{ .name }}.id
		}
	`

	params := map[string]interface{}{
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryBasic := util.ExecuteTemplate("TestAccFederatedHelmOciRepository", template, params)

	updated := map[string]interface{}{
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryUpdated := util.ExecuteTemplate("TestAccFederatedHelmOciRepository", template, updated)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "helmoci"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", updated["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", updated["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "helmoci"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceFederatedNugetRepository(t *testing.T) {
	_, tempFqrn, name := testutil.MkNames("nuget-federated", "artifactory_federated_nuget_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_federated_nuget_repository" "{{ .name }}" {
			key                        = "{{ .name }}"
			max_unique_snapshots       = {{ .max_unique_snapshots }}
			force_nuget_authentication = {{ .force_nuget_authentication }}
			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
		data "artifactory_federated_nuget_repository" "{{ .name }}" {
			key = artifactory_federated_nuget_repository.{{ .name }}.id
		}
	`

	params := map[string]interface{}{
		"force_nuget_authentication": testutil.RandBool(),
		"max_unique_snapshots":       testutil.RandSelect(0, 5, 10),
		"name":                       name,
		"memberUrl":                  federatedMemberUrl,
	}
	federatedRepositoryBasic := util.ExecuteTemplate("TestAccLocalNugetRepository", template, params)

	updates := map[string]interface{}{
		"force_nuget_authentication": testutil.RandBool(),
		"max_unique_snapshots":       testutil.RandSelect(0, 5, 10),
		"name":                       name,
		"memberUrl":                  federatedMemberUrl,
	}
	federatedRepositoryUpdated := util.ExecuteTemplate("TestAccLocalNugetRepository", template, updates)

	fqrn := "data." + tempFqrn

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, tempFqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", params["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", fmt.Sprintf("%t", params["force_nuget_authentication"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "nuget"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", updates["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", fmt.Sprintf("%t", updates["force_nuget_authentication"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "nuget"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceFederatedOciRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("oci-federated", "data.artifactory_federated_oci_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_federated_oci_repository" "{{ .name }}" {
			key 	        = "{{ .name }}"
			tag_retention   = {{ .retention }}
			max_unique_tags = {{ .max_tags }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
		data "artifactory_federated_oci_repository" "{{ .name }}" {
			key = artifactory_federated_oci_repository.{{ .name }}.id
		}
	`

	params := map[string]interface{}{
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryBasic := util.ExecuteTemplate("TestAccFederatedOciRepository", template, params)

	updated := map[string]interface{}{
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryUpdated := util.ExecuteTemplate("TestAccFederatedOciRepository", template, updated)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "oci"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", updated["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", updated["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "oci"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func TestAccDataSourceFederatedRpmRepository(t *testing.T) {
	_, tempFqrn, name := testutil.MkNames("rpm-federated", "artifactory_federated_rpm_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")

	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
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

		resource "artifactory_federated_rpm_repository" "{{ .repo_name }}" {
			key 	                   = "{{ .repo_name }}"
			primary_keypair_ref        = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref      = artifactory_keypair.{{ .kp_name2 }}.pair_name
			yum_root_depth             = {{ .yum_root_depth }}
			enable_file_lists_indexing = {{ .enable_file_lists_indexing }}
			calculate_yum_metadata     = true

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}

			depends_on = [
				artifactory_keypair.{{ .kp_name }},
				artifactory_keypair.{{ .kp_name2 }},
			]
		}
		data "artifactory_federated_rpm_repository" "{{ .repo_name }}" {
			key = artifactory_federated_rpm_repository.{{ .repo_name }}.id
		}
	`

	federatedRepositoryBasic := util.ExecuteTemplate("keypair", template, map[string]interface{}{
		"kp_id":                      kpId,
		"kp_name":                    kpName,
		"kp_id2":                     kpId2,
		"kp_name2":                   kpName2,
		"repo_name":                  name,
		"yum_root_depth":             1,
		"enable_file_lists_indexing": true,
		"memberUrl":                  federatedMemberUrl,
		"private_key":                os.Getenv("JFROG_TEST_PGP_PRIVATE_KEY"),
		"public_key":                 os.Getenv("JFROG_TEST_PGP_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	federatedRepositoryUpdated := util.ExecuteTemplate("keypair", template, map[string]interface{}{
		"kp_id":                      kpId,
		"kp_name":                    kpName,
		"kp_id2":                     kpId2,
		"kp_name2":                   kpName2,
		"repo_name":                  name,
		"yum_root_depth":             2,
		"enable_file_lists_indexing": false,
		"memberUrl":                  federatedMemberUrl,
		"private_key":                os.Getenv("JFROG_TEST_PGP_PRIVATE_KEY"),
		"public_key":                 os.Getenv("JFROG_TEST_PGP_PUBLIC_KEY"),
	})

	fqrn := "data." + tempFqrn

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, tempFqrn, "", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "", security.VerifyKeyPair),
			acctest.VerifyDeleted(t, kpFqrn2, "", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "rpm"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "enable_file_lists_indexing", "true"),
					resource.TestCheckResourceAttr(fqrn, "calculate_yum_metadata", "true"),
					resource.TestCheckResourceAttr(fqrn, "yum_root_depth", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "rpm"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "rpm"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "enable_file_lists_indexing", "false"),
					resource.TestCheckResourceAttr(fqrn, "calculate_yum_metadata", "true"),
					resource.TestCheckResourceAttr(fqrn, "yum_root_depth", "2"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "rpm"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	})
}

func makeFederatedTerraformRepoTestCase(registryType string, t *testing.T) (*testing.T, resource.TestCase) {
	resourceName := fmt.Sprintf("terraform-module-%s", registryType)
	resourceType := fmt.Sprintf("artifactory_federated_terraform_%s_repository", registryType)
	_, tempFqrn, name := testutil.MkNames(resourceName, resourceType)
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	params := map[string]interface{}{
		"registryType": registryType,
		"name":         name,
		"memberUrl":    federatedMemberUrl,
	}

	template := `
		resource "artifactory_federated_terraform_{{ .registryType }}_repository" "{{ .name }}" {
			key = "{{ .name }}"

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
		data "artifactory_federated_terraform_{{ .registryType }}_repository" "{{ .name }}" {
			key = artifactory_federated_terraform_{{ .registryType }}_repository.{{ .name }}.id
		}
	`
	federatedRepositoryBasic := util.ExecuteTemplate("TestAccFederatedTerraformRepository", template, params)

	fqrn := "data." + tempFqrn

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(t, tempFqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "terraform"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("federated", "terraform_"+registryType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
		},
	}
}

func TestAccDataSourceFederatedTerraformRepositories(t *testing.T) {
	for _, registryType := range []string{"module", "provider"} {
		t.Run(registryType, func(t *testing.T) {
			resource.Test(makeFederatedTerraformRepoTestCase(registryType, t))
		})
	}
}

func TestAccDataSourceFederatedMissingRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("terraform-local", "data.artifactory_federated_terraform_provider_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := util.ExecuteTemplate(
		"TestAccFederatedTerraformProviderRepository",
		`data "artifactory_federated_terraform_provider_repository" "{{ .name }}" {
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
