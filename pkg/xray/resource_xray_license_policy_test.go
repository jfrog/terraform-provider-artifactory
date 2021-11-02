package xray

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// License policy criteria are different from the security policy criteria
// Test will try to post a new license policy with incorrect body of security policy
// with specified cvss_range. The function expandLicenseCriteria will ignore all the
// fields except of "allow_unknown", "banned_licenses" and "allowed_licenses" if the Policy type is "license"
func TestAccLicensePolicy_badLicenseCriteria(t *testing.T) {
	policyName := "terraform-license-policy-1"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-license-rule-1"
	rangeTo := 5
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_license_policy." + resourceName

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccXrayLicensePolicy_badLicense(resourceName, policyName, policyDesc, ruleName, rangeTo),
				ExpectError: regexp.MustCompile("\"error\":\"Rule " + ruleName + " has empty criteria\""),
			},
		},
	})
}

// This test will try to create a license policy with failure grace period set, but without fail build turned on
func TestAccLicensePolicy_badGracePeriod(t *testing.T) {
	policyName := "terraform-license-policy-2"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-license-rule-2"
	gracePeriod := 5
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_license_policy." + resourceName

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccXrayLicensePolicy_badGracePeriod(resourceName, policyName, policyDesc, ruleName, gracePeriod),
				ExpectError: regexp.MustCompile("Rule " + ruleName + " has failure grace period without fail build"),
			},
		},
	})
}

func TestAccLicensePolicy_createAllowedLic(t *testing.T) {
	policyName := "terraform-license-policy-3"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-license-rule-3"
	gracePeriod := 5
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_license_policy." + resourceName
	multiLicense := "true"
	blockUnscanned := "true"
	blockActive := "true"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayLicensePolicy_createAllowedLic(resourceName, policyName, policyDesc,
					ruleName, multiLicense, gracePeriod, blockUnscanned, blockActive),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", policyDesc),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", strconv.Itoa(gracePeriod)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.custom_severity", "High"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", "true"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", "true"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", "true"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allowed_licenses.0", "Apache-1.0"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allowed_licenses.1", "Apache-2.0"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.multi_license_permissive", "true"),
				),
			},
		},
	})
}

func TestAccLicensePolicy_createBannedLic(t *testing.T) {
	policyName := "terraform-license-policy-4"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-license-rule-4"
	gracePeriod := 5
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_license_policy." + resourceName

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayLicensePolicy_createBanneddLic(resourceName, policyName, policyDesc, ruleName, gracePeriod),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", policyDesc),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", strconv.Itoa(gracePeriod)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.custom_severity", "High"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", "true"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", "true"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", "true"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.banned_licenses.0", "Apache-1.0"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.banned_licenses.1", "Apache-2.0"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.multi_license_permissive", "true"),
				),
			},
		},
	})
}

func TestAccLicensePolicy_createMultiLicensePermissiveFalse(t *testing.T) {
	policyName := "terraform-license-policy-5"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-license-rule-5"
	gracePeriod := 5
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_license_policy." + resourceName
	multiLicense := "false"
	blockUnscanned := "true"
	blockActive := "true"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayLicensePolicy_createAllowedLic(resourceName, policyName, policyDesc,
					ruleName, multiLicense, gracePeriod, blockUnscanned, blockActive),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", policyDesc),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", strconv.Itoa(gracePeriod)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.custom_severity", "High"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", "true"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", "true"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", "true"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allowed_licenses.0", "Apache-1.0"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allowed_licenses.1", "Apache-2.0"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.multi_license_permissive", "false"),
				),
			},
		},
	})
}

func TestAccLicensePolicy_createBlockFalse(t *testing.T) {
	policyName := "terraform-license-policy-6"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-license-rule-6"
	gracePeriod := 5
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_license_policy." + resourceName
	multiLicense := "false"
	blockUnscanned := "false"
	blockActive := "false"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayLicensePolicy_createAllowedLic(resourceName, policyName, policyDesc,
					ruleName, multiLicense, gracePeriod, blockUnscanned, blockActive),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", policyDesc),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", strconv.Itoa(gracePeriod)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.custom_severity", "High"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", "true"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", "false"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", "false"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allowed_licenses.0", "Apache-1.0"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allowed_licenses.1", "Apache-2.0"),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.multi_license_permissive", "false"),
				),
			},
		},
	})
}

func testAccCheckLicensePolicyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: Resource id [%s] not found", id)
		}
		provider, _ := testAccProviders["xray"]()
		_, resp, _ := getPolicy(rs.Primary.ID, provider.Meta().(*resty.Client))

		if resp.StatusCode() == http.StatusOK {
			return fmt.Errorf("error: Policy %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccXrayLicensePolicy_badLicense(resourceName, name, description, ruleName string, rangeTo int) string {
	return fmt.Sprintf(`
resource "xray_security_policy" "%s" {
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
`, resourceName, name, description, ruleName, rangeTo)
}

func testAccXrayLicensePolicy_badGracePeriod(resourceName, name, description, ruleName string, gracePeriod int) string {
	return fmt.Sprintf(`
resource "xray_license_policy" "%s" {
	name = "%s"
	description = "%s"
	type = "license"
	rules {
		name = "%s"
		priority = 1
		criteria {
		  allowed_licenses = ["Apache-1.0","Apache-2.0"]
          allow_unknown = false
          multi_license_permissive = true
		}
		actions {
			fail_build = false
			build_failure_grace_period_in_days = %d
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}
`, resourceName, name, description, ruleName, gracePeriod)
}

func testAccXrayLicensePolicy_createAllowedLic(resourceName, name, description, ruleName string,
	multiLicense string, gracePeriod int, blockUnscanned string, blockActive string) string {
	return fmt.Sprintf(`
resource "xray_license_policy" "%s" {
	name = "%s"
	description = "%s"
	type = "license"
	rules {
		name = "%s"
		priority = 1
		criteria {
		  allowed_licenses = ["Apache-1.0","Apache-2.0"]
          allow_unknown = false
          multi_license_permissive = %s
		}
		actions {
			fail_build = true
			build_failure_grace_period_in_days = %d
			custom_severity = "High"			
			block_download {
				unscanned = %s
				active = %s
			}
		}
	}
}
`, resourceName, name, description, ruleName, multiLicense, gracePeriod, blockUnscanned, blockActive)
}

func testAccXrayLicensePolicy_createBanneddLic(resourceName, name, description, ruleName string, gracePeriod int) string {
	return fmt.Sprintf(`
resource "xray_license_policy" "%s" {
	name = "%s"
	description = "%s"
	type = "license"
	rules {
		name = "%s"
		priority = 1
		criteria {
		  banned_licenses = ["Apache-1.0","Apache-2.0"]
          allow_unknown = true
          multi_license_permissive = true
		}
		actions {
			fail_build = true
			build_failure_grace_period_in_days = %d
			custom_severity = "High"			
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}
`, resourceName, name, description, ruleName, gracePeriod)
}
