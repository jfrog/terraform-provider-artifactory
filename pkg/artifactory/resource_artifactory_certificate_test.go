package artifactory

import (
	"context"
	"fmt"
	"testing"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const certificateFull = `
resource "artifactory_certificate" "foobar" {
    alias   = "foobar"
    content = <<EOF
-----BEGIN CERTIFICATE-----
MIICUjCCAbugAwIBAgIJALRDng3rGeQvMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNV
BAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQg
Q29tcGFueSBMdGQwHhcNMTkwNTE3MTAwMzI2WhcNMjkwNTE0MTAwMzI2WjBCMQsw
CQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNEZWZh
dWx0IENvbXBhbnkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDVBRt7
Ua3j7K2htVRu1tw629ZZZQI35RGm/53ffF/QUUFXk35at+IiwYZGGQbOGuN1pdji
gki9/Qit/WO/3uadSkGelKOUYD0DIemlhcZt6iPMQq8mYlUkMPZz5Qlj0ldKI3g+
Q8Tc/6vEeBv/9jrm9Efg/uwc0DjD8B4Ny6xMHQIDAQABo1AwTjAdBgNVHQ4EFgQU
VrBaHnYLayO2lKIUde8etG0H6owwHwYDVR0jBBgwFoAUVrBaHnYLayO2lKIUde8e
tG0H6owwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQA4VBFCrbuOsKtY
uNlSQCBkTXg907iXihZ+Of/2rerS2gfDCUHdz0xbYdlttNjoGVCA+0alt7ugfYpl
fy5aAfCHLXEgYrlhe6oDtCMSskbkKFTEI/bRqwGMDb+9NO/yh2KLbNueKJz9Vs5V
GV9pUrgW6c7kLrC9vpHP+47iyQEbnw==
-----END CERTIFICATE-----
-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBANUFG3tRrePsraG1
VG7W3Drb1lllAjflEab/nd98X9BRQVeTflq34iLBhkYZBs4a43Wl2OKCSL39CK39
Y7/e5p1KQZ6Uo5RgPQMh6aWFxm3qI8xCryZiVSQw9nPlCWPSV0ojeD5DxNz/q8R4
G//2Oub0R+D+7BzQOMPwHg3LrEwdAgMBAAECgYAxWA6GoWQDcRbDZ6qYRkMbi0L6
0DAUXIabRYj/dOMI8VmOfMb/IqtKW8PLxw5Rfd8EqJc12PIauFtjWlfZ4TtP9erQ
1imw2SpVMAWt4HLUw7oONKgNMnBtVQBCoXLuXcnJbCxeRiV1oJtvrddUJPOtUc+y
t5gGTyx/zUAXzPzT7QJBAOvu4CH0Xc+1GdXFUFLzF8B3SFwnOFRERJxFq43dw4t3
tXcON/UyegYcQz2JqKcofwRhM4+uXGnWE+9oOOnxL8sCQQDnI1QtMv+tZcqIcmk6
1ykyNa530eCfoqAvVTRwPIsAD/DZLC4HJNSQauPXC4Unt1tqmOmUoZmgzYQlVsGO
ISa3AkB2xWpPrZUMWz8GPq6RE4+BdIsY2SWiRjvD787NPDaUn07bAG1rIl4LdW7k
K8ibXeeTbNtoGX6sSPkALJd6LdDBAkEA5FAhdgRKSh2iUeWxzE18g/xCuli2aPlb
AWZIxhUHuKgGYH8jeCsJTR5IsMLQZMrZohIpqId4GT7oqXlo99wHQQJBAOvX+5z6
iCooatRyMnwUV6sJ225ZawuJ4sXFt6CA7aOZQ+G5zvG694ONxG9qeF2YnySQp1HH
V87CqqFaUigTzmI=
-----END PRIVATE KEY-----
EOF
}`

func TestAccCertificate_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckCertificateDestroy("artifactory_certificate.foobar"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: certificateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_certificate.foobar", "alias", "foobar"),
					resource.TestCheckResourceAttr("artifactory_certificate.foobar", "fingerprint", "ED:67:0B:D2:84:C2:93:6D:56:6F:A7:4D:5A:CC:B7:AF:8A:C0:1D:2A:7C:F3:4A:57:31:83:22:30:44:5F:63:9D"),
					resource.TestCheckResourceAttr("artifactory_certificate.foobar", "issued_by", "Unknown"),
					resource.TestCheckResourceAttr("artifactory_certificate.foobar", "issued_on", "2019-05-17T11:03:26.000+01:00"),
					resource.TestCheckResourceAttr("artifactory_certificate.foobar", "issued_to", "Unknown"),
					resource.TestCheckResourceAttr("artifactory_certificate.foobar", "valid_until", "2029-05-14T11:03:26.000+01:00"),
				),
			},
		},
	})
}

func testAccCheckCertificateDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*artifactory.Artifactory)
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		certs, _, err := client.V1.Security.GetCertificates(context.Background())
		if err != nil {
			return err
		}

		for _, cert := range *certs {
			if *cert.CertificateAlias == rs.Primary.ID {
				return fmt.Errorf("error: Certificate %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}
