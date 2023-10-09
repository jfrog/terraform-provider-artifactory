package security_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const resource_name = "artifactory_distribution_public_key"

const template = `resource "%s" "%s" {
  alias = "%s"
  public_key = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: Keybase OpenPGP v1.0.0
Comment: https://keybase.io/crypto

xo0EYGrVNAEEAPD3YDt0qP8kSV8bnmqVP5XDPoN40gEpUGtDLjAn6d+cRMeNGaru
6H0bdgwQpND8Gz9Qx2pCNSxlWDZpY1fCvRQ174iGjvO/3527f148cgKNZtwLsKrZ
laW8z3tB2LuCM2e97ijX+lzRf7YJUXU3pOfoCFWpOPoRg1CHV0NyHl0VABEBAAHN
FmFsYW4gPGFsYW5uQGpmcm9nLmNvbT7CrQQTAQoAFwUCYGrVNAIbLwMLCQcDFQoI
Ah4BAheAAAoJENzR2QJlA6glZmsD/iqhnNFy1Elj3hGL0HaEzeb+KDpcSL/L5a/8
WIGCQFeLcEn9lC+68b/eERKGIoXJ7z8HfPDFNRTKvomKIdAqFiAeDAUUD0B82rsx
xDf8USnTwJlnd0bPe9nxgXYcrwioEYbPVYGl3jima/KQrbW8XlKyiypy4Nd66Wcn
TuM6PwRFzo0EYGrVNAEEANVNINyfCQ+y1haaaAJ0uCgx3dW52LwcZfvOP6i798WZ
dyGA+WSUCEcrklUwZ595E2dNkNKptksftwSeQ0+EH5S1ZlEaq2YUv8fCx32F1ckh
D3eHaCKRxTPx/zbb96q4ruEGKhOBXceid3o341HbtGVKi8VjBx3XNukskQ+EOvgt
ABEBAAHCwIMEGAEKAA8FAmBq1TQFCQ8JnAACGy4AqAkQ3NHZAmUDqCWdIAQZAQoA
BgUCYGrVNAAKCRBddQ63FhKl6NmDBACqxC4lAnsCQERjs02LYAEAwVDhDf0rXxD0
H+hKDyxQZc80M7WIpXaBHmbs8ekJRnY7JHcer7sizDMdfkR3xB62jNGhc0XiW6nc
mlwvtWt3+E6AkObmWnocRy5ztTQI0gye0B3cPs2txE2fCs+WD7yLRnM3HqIAh83W
Cccvh0+uG96dBADlPbZ8g8q6bkeeT72gOi3OCN0A+Y8lUPifhrpSiI9xMpP3aomM
beZJB6fWjEzNoblQ9jUr/E54bF9jMr6L3uE4OJH9SYJ/HvqcKJC+1TFeQ9lXR7g7
MTdfxEvhMDhcsd/pYIgrzvDry+B2+jANW10R1yejT/C8QdlWIndDsEsaKs6NBGBq
1TQBBADJD6gVGTMGb/WsfnSaL5yA3AEMczhPqxD4FDC0vzqGG3XzKgxtmW8cJXls
NCK80F61daxJ72/UAmfxHbNP1qmHSiosEe+h1YZ/Zo3pN/LODzg9JrOs9A2xwjqE
nU9mS0jDz5oEQtr9K4+YKOdJvmuaN85ueBizfQ1TYRNuDtmbnQARAQABwsCDBBgB
CgAPBQJgatU0BQkPCZwAAhsuAKgJENzR2QJlA6glnSAEGQEKAAYFAmBq1TQACgkQ
ehctgYvYtmh38gP6A9lnQaLuVnTElJLy2XSDTqwWOcy/5J842S/xdQEsWUMXh4I5
mlotkZwkrdvXp8E/F3P8X7GbxhNAVZX+Xcm95V3g/kmP+Pq7PeUmoZR5LD8ppBfO
7v6XgaUhraUPAZl6lx4L5pYNCX9JBNUtQAG9xIoap4slvksdz5SN/BwSgV6qqwQA
tr4YTDXvLyoWwMFB2FjWcw4zwV+7yHwGzogKfGCQy5qVlDoQyWdkwwF1awyk5RIe
ZxwPZ2SDaiznOmZ+4LjR2NPmjnT96d9RKRtgEjkfW+a19BofrvEalS9wh/jkboea
d8rDu8wMbLAl77dq1c6dpJDgzoQkekoL4H4GU8QB6GY=
=fot9
-----END PGP PUBLIC KEY BLOCK-----
EOF
}`

func TestAccDistributionPublicKey_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("mykey", resource_name)

	keyBasic := fmt.Sprintf(template, resource_name, name, name)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "9.2.1",
						Source:            "registry.terraform.io/jfrog/artifactory",
					},
				},
				Config: keyBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "alias", name),
					resource.TestMatchResourceAttr(fqrn, "public_key", regexp.MustCompile(".*-----BEGIN PGP PUBLIC KEY BLOCK-----.*")),
					resource.TestCheckResourceAttr(fqrn, "fingerprint", "10:16:2c:c5:1c:db:d0:59:ad:86:d3:66:dc:d1:d9:02:65:03:a8:25"),
					resource.TestCheckResourceAttr(fqrn, "issued_by", "alan <alann@jfrog.com>"),
				),
				ConfigPlanChecks: acctest.ConfigPlanChecks,
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   keyBasic,
				PlanOnly:                 true,
				ConfigPlanChecks:         acctest.ConfigPlanChecks,
			},
		},
	})
}

func TestAccDistributionPublicKey_FormatCheck(t *testing.T) {
	id, _, name := testutil.MkNames("mykey", resource_name)
	keyBasic := fmt.Sprintf(`
		resource "%s" "%s" {
			alias = "foo-alias%d"
			public_key = "not a public key"
		}
	`, resource_name, name, id)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      keyBasic,
				ExpectError: regexp.MustCompile(".*RSA public key not in pem format.*"),
			},
		},
	})
}

func TestAccDistributionPublicKey_Create(t *testing.T) {
	_, fqrn, name := testutil.MkNames("mykey", resource_name)
	keyBasic := fmt.Sprintf(template, resource_name, name, name)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionPublicKeyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: keyBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "key_id"),
					resource.TestCheckResourceAttr(fqrn, "alias", name),
					resource.TestMatchResourceAttr(fqrn, "public_key", regexp.MustCompile(".*-----BEGIN PGP PUBLIC KEY BLOCK-----.*")),
					resource.TestCheckResourceAttr(fqrn, "fingerprint", "10:16:2c:c5:1c:db:d0:59:ad:86:d3:66:dc:d1:d9:02:65:03:a8:25"),
					resource.TestCheckResourceAttr(fqrn, "issued_by", "alan <alann@jfrog.com>"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "alias",
			},
		},
	})
}

func testAccCheckDistributionPublicKeyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(utilsdk.ProvderMetadata).Client

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		data := security.DistributionPublicKeysList{}
		resp, err := client.R().SetResult(&data).Get(security.DistributionPublicKeysAPIEndPoint)
		if err != nil {
			return err
		}
		if resp.IsError() {
			return fmt.Errorf("unable to read keys: http request failed: %s", resp.Status())
		}

		for _, key := range data.Keys {
			if key.KeyId == rs.Primary.ID {
				return fmt.Errorf("error: Distribution Public Key %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}
