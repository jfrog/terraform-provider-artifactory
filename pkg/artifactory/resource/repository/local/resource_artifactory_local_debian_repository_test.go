package local_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalDebianRepository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("local-debian-repo", "artifactory_local_debian_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")

	temp := `
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
			key = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
			index_compression_formats = {{ .compression_formats }}
			trivial_layout = {{ .trivial_layout }}
			ddeb_supported = {{ .ddeb_supported }}
		}
	`

	config := util.ExecuteTemplate("keypair", temp, map[string]interface{}{
		"kp_id":               kpId,
		"kp_name":             kpName,
		"kp_id2":              kpId2,
		"kp_name2":            kpName2,
		"repo_name":           name,
		"private_key":         os.Getenv("JFROG_TEST_PGP_PRIVATE_KEY"),
		"public_key":          os.Getenv("JFROG_TEST_PGP_PUBLIC_KEY"),
		"compression_formats": `["bz2", "lzma", "xz"]`,
		"trivial_layout":      true,
		"ddeb_supported":      true,
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	updatedTemp := `
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
			key = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
			trivial_layout = {{ .trivial_layout }}
			ddeb_supported = {{ .ddeb_supported }}
		}
	`

	updatedConfig := util.ExecuteTemplate("keypair", updatedTemp, map[string]interface{}{
		"kp_id":          kpId,
		"kp_name":        kpName,
		"kp_id2":         kpId2,
		"kp_name2":       kpName2,
		"repo_name":      name,
		"private_key":    os.Getenv("JFROG_TEST_PGP_PRIVATE_KEY"),
		"public_key":     os.Getenv("JFROG_TEST_PGP_PUBLIC_KEY"),
		"trivial_layout": false,
		"ddeb_supported": false,
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
			acctest.VerifyDeleted(t, kpFqrn2, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.#", "3"),
					resource.TestCheckTypeSetElemAttr(fqrn, "index_compression_formats.*", "bz2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "index_compression_formats.*", "lzma"),
					resource.TestCheckTypeSetElemAttr(fqrn, "index_compression_formats.*", "xz"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.DebianPackageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
					resource.TestCheckResourceAttr(fqrn, "trivial_layout", "true"),
					resource.TestCheckResourceAttr(fqrn, "ddeb_supported", "true"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "index_compression_formats.*", "bz2"),
					resource.TestCheckResourceAttr(fqrn, "trivial_layout", "false"),
					resource.TestCheckResourceAttr(fqrn, "ddeb_supported", "false"),
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

func TestAccLocalDebianRepository_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("local-debian-repo", "artifactory_local_debian_repository")
	kpId, _, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, _, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")

	config := util.ExecuteTemplate("keypair", `
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
			key = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
			index_compression_formats = ["bz2","lzma","xz"]
			trivial_layout = true
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
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "10.3.3",
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
