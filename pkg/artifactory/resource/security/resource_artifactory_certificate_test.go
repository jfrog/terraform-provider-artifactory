package security_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccCertHasFileAndContentFails(t *testing.T) {
	const conflictsResource = `
		resource "artifactory_certificate" "fail" {
			alias   = "fail"
			file = "/this/doesnt/exist.pem"
			content = "PEM DATA"
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      conflictsResource,
				ExpectError: regexp.MustCompile(".*only one of `content,file` can be specified, but .* were.*"),
			},
		},
	})
}
func TestAccCertWithFileMissing(t *testing.T) {
	const certWithMissingFile = `
		resource "artifactory_certificate" "fail" {
			alias   = "fail"
			file = "/this/doesnt/exist.pem"
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy("artifactory_certificate.fail"),
		Steps: []resource.TestStep{
			{
				Config:      certWithMissingFile,
				ExpectError: regexp.MustCompile(`.*stat /this/doesnt/exist.pem: no such file or directory.*`),
			},
		},
	})
}

func TestAccCertWithFile(t *testing.T) {
	const certWithFile = `
		resource "artifactory_certificate" "%s" {
			alias   = "%s"
			file = "../../../../samples/cert.pem"
		}
	`
	id := test.RandomInt()
	name := fmt.Sprintf("foobar%d", id)
	fqrn := fmt.Sprintf("artifactory_certificate.%s", name)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(certWithFile, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "alias", name),
					resource.TestCheckResourceAttr(fqrn, "fingerprint", "ED:67:0B:D2:84:C2:93:6D:56:6F:A7:4D:5A:CC:B7:AF:8A:C0:1D:2A:7C:F3:4A:57:31:83:22:30:44:5F:63:9D"),
					resource.TestCheckResourceAttr(fqrn, "issued_by", "Unknown"),
					resource.TestCheckResourceAttr(fqrn, "issued_on", "2019-05-17T10:03:26.000Z"),
					resource.TestCheckResourceAttr(fqrn, "issued_to", "Unknown"),
					resource.TestCheckResourceAttr(fqrn, "valid_until", "2029-05-14T10:03:26.000Z"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "alias"),
				ImportStateVerifyIgnore: []string{"file"}, // actual certificate is not returned via the API, so it cannot be "imported"
			},
		},
	})
}

func TestAccCertificate_full(t *testing.T) {
	const certificateFull = `
		resource "artifactory_certificate" "%s" {
			alias   = "%s"
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
		}
	`
	id := test.RandomInt()
	name := fmt.Sprintf("foobar%d", id)
	fqrn := fmt.Sprintf("artifactory_certificate.%s", name)
	subbed := fmt.Sprintf(certificateFull, name, name)
	cleansed := strings.Replace(subbed, "\t", "", -1)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: cleansed,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "alias", name),
					resource.TestCheckResourceAttr(fqrn, "fingerprint", "ED:67:0B:D2:84:C2:93:6D:56:6F:A7:4D:5A:CC:B7:AF:8A:C0:1D:2A:7C:F3:4A:57:31:83:22:30:44:5F:63:9D"),
					resource.TestCheckResourceAttr(fqrn, "issued_by", "Unknown"),
					resource.TestCheckResourceAttr(fqrn, "issued_on", "2019-05-17T10:03:26.000Z"),
					resource.TestCheckResourceAttr(fqrn, "issued_to", "Unknown"),
					resource.TestCheckResourceAttr(fqrn, "valid_until", "2029-05-14T10:03:26.000Z"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "alias"),
				ImportStateVerifyIgnore: []string{"content"}, // actual certificate is not returned via the API, so it cannot be "imported"
			},
		},
	})
}

func testAccCheckCertificateDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		cert, err := security.FindCertificate(id, acctest.Provider.Meta())
		if err != nil {
			return err
		}

		if cert != nil {
			return fmt.Errorf("error: Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
