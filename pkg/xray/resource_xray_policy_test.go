package xray

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPolicy_badLicenseCriteria(t *testing.T) {
	policyName := "terraform-test-policy"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule"
	rangeTo := 4

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckPolicyDestroy,
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccXrayPolicy_badLicense(policyName, policyDesc, ruleName, rangeTo),
				ExpectError: regexp.MustCompile("min_severity and cvvs_range are not supported with license policies"),
			},
		},
	})
}

func TestAccPolicy_badSecurityCriteria(t *testing.T) {
	policyName := "terraform-test-policy"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule"
	allowedLicense := "BSD-4-Clause"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckPolicyDestroy,
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccXrayPolicy_badSecurity(policyName, policyDesc, ruleName, allowedLicense),
				ExpectError: regexp.MustCompile("An argument named \"allow_unknown\" is not expected here."),
			},
		},
	})
}

func testAccCheckPolicyDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "xray_security_policy" {
			provider, _ := testAccProviders["xray"]()
			policy, resp, err := getPolicy(rs.Primary.ID, provider.Meta().(*resty.Client))

			if err != nil {
				if resp != nil {
					if resp.StatusCode() == http.StatusInternalServerError &&
						err.Error() == fmt.Sprintf("{\"error\":\"Failed to find Policy %s\"}", rs.Primary.ID) {
						continue
					}
					if resp.StatusCode() == http.StatusNotFound {
						continue
					}
				}
				return err
			}
			return fmt.Errorf("error: Policy %s still exists %s", rs.Primary.ID, *policy.Name)
		}
	}
	return nil
}

func testAccXrayPolicy_badLicense(name, description, ruleName string, rangeTo int) string {
	return fmt.Sprintf(`
resource "xray_security_policy" "test" {
	name = "%s"
	description = "%s"
	type = "license"
	rules {
		name = "%s"
		priority = 1
		criteria {
			cvss_range {
				from = 1
				to = %d
			}
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}
`, name, description, ruleName, rangeTo)
}

func testAccXrayPolicy_badSecurity(name, description, ruleName, allowedLicense string) string {
	return fmt.Sprintf(`
resource "xray_security_policy" "test" {
	name = "%s"
	description = "%s"
	type = "security"
	rules {
		name = "%s"
		priority = 1
		criteria {
			allow_unknown = true
			allowed_licenses = ["%s"]
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}
`, name, description, ruleName, allowedLicense)
}

//TODO: Add test to create a policy
//func TestAccPolicy_create(t *testing.T) {
//	policyName := "terraform-test-policy"
//	policyDesc := "policy created by xray acceptance tests"
//	ruleName := "test-security-rule"
//	rangeTo := 4
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:          func() { testAccPreCheck(t) },
//		CheckDestroy:      testAccCheckPolicyDestroy,
//		ProviderFactories: testAccProviders,
//		Steps: []resource.TestStep{
//			{
//				Config:      testAccXrayPolicy_valid(policyName, policyDesc, ruleName, rangeTo),
//				Check: 		resource.TestCheckResourceAttr("name", "issued_by", "Unknown"),
//			},
//		},
//	})
//}
//
//func testAccXrayPolicy_valid(name, description, ruleName string, rangeTo int) string {
//	return fmt.Sprintf(`
//resource "xray_security_policy" "test" {
//	name = "%s"
//	description = "%s"
//	type = "security"
//	rules {
//		name = "%s"
//		priority = 1
//		criteria {
//			cvss_range {
//				from = 1
//				to = %d
//			}
//		}
//		actions {
//			block_download {
//				unscanned = true
//				active = true
//			}
//		}
//	}
//}
//`, name, description, ruleName, rangeTo)
//}
