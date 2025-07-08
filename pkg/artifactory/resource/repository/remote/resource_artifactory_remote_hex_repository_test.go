package remote_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRemoteHexRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-hex-remote", "artifactory_remote_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")
	publicKeyBytes, err := os.ReadFile("../../../../../samples/hex_public_key")
	if err != nil {
		t.Fatalf("failed to read hex public key: %v", err)
	}
	publicKey := string(publicKeyBytes)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate("TestAccRemoteHexRepository", `
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

                    resource "artifactory_remote_hex_repository" "{{ .name }}" {
                        key = "{{ .name }}"
						url = "https://www.hex.pm/"
						public_key_ref = <<EOF
{{ .hex_public_key }}
EOF
                        hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
                    }
                `, map[string]interface{}{
					"name":           name,
					"kp_id":          kpId,
					"kp_name":        kpName,
					"hex_public_key": publicKey,
					"private_key":    os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
					"public_key":     os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", "https://www.hex.pm/"),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
				),
			},
		},
	})
}
