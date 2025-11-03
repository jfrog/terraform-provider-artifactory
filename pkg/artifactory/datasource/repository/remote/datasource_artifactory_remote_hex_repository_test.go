package remote_test

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

func TestAccRemoteHexRepositoryDataSource_framework(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-remote-test-repo", "artifactory_remote_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")
	publicKeyBytes, err := os.ReadFile("../../../../../samples/hex_public_key")
	if err != nil {
		t.Fatalf("failed to read hex public key: %v", err)
	}
	publicKey := string(publicKeyBytes)

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

        resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
            key = "{{ .repo_name }}"
            url = "https://hex.pm"
            public_key_ref = <<EOF
{{ .hex_public_key }}
EOF
            hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
        }

        data "artifactory_remote_hex_repository" "{{ .repo_name }}" {
            key = artifactory_remote_hex_repository.{{ .repo_name }}.key
        }
    `

	testData := map[string]interface{}{
		"kp_id":          kpId,
		"kp_name":        kpName,
		"repo_name":      name,
		"hex_public_key": publicKey,
		"private_key":    os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":     os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("keypair", temp, testData)

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
					resource.TestCheckResourceAttr(fqrn, "url", "https://hex.pm"),
					resource.TestCheckResourceAttrSet(fqrn, "public_key_ref"), // Check that it's set (actual content may vary)
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("remote", repository.HexPackageType)
						return r
					}()),
					// Test the data source
					resource.TestCheckResourceAttr("data.artifactory_remote_hex_repository."+name, "key", name),
					resource.TestCheckResourceAttr("data.artifactory_remote_hex_repository."+name, "url", "https://hex.pm"),
					resource.TestCheckResourceAttrSet("data.artifactory_remote_hex_repository."+name, "public_key_ref"), // Check that it's set (actual content may vary)
					resource.TestCheckResourceAttr("data.artifactory_remote_hex_repository."+name, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr("data.artifactory_remote_hex_repository."+name, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("remote", repository.HexPackageType)
						return r
					}()),
				),
			},
		},
	})
}
