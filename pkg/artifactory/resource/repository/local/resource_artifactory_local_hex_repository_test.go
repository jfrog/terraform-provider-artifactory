package local_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalHexRepository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-local-test-repo", "artifactory_local_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	temp := `
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
        }

        resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
            key = "{{ .repo_name }}"
            hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
        }
    `

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("keypair", temp, testData)

	updatedTemp := `
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
        }

        resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
            key = "{{ .repo_name }}"
            project_environments = []
            hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
        }
    `

	updatedConfig := util.ExecuteTemplate("keypair", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.HexPackageType)
						return r
					}()),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
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
