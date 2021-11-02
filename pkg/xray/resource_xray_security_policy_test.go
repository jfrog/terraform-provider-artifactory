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

// Teh test will try to create a security policy with the type of "license"
// The Policy criteria will be ignored in this case
func TestAccSecurityPolicy_badTypeInSecurityPolicy(t *testing.T) {
	policyName := "terraform-security-policy-1"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule-1"
	rangeTo := 5
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_security_policy." + resourceName
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckSecurityPolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccXraySecurityPolicy_badSecurityType(policyName, policyDesc, ruleName, rangeTo),
				ExpectError: regexp.MustCompile("Rule " + ruleName + " has empty criteria"),
			},
		},
	})
}

// The test will try to use "allowed_licenses" in the security policy criteria
// That field is acceptable only in license policy. No API call, expected to fail on the TF resource verification
func TestAccSecurityPolicy_badSecurityCriteria(t *testing.T) {
	policyName := "terraform-security-policy-2"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule-2"
	allowedLicense := "BSD-4-Clause"
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_security_policy." + resourceName
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckSecurityPolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccXraySecurityPolicy_badSecurity(policyName, policyDesc, ruleName, allowedLicense),
				ExpectError: regexp.MustCompile("An argument named \"allow_unknown\" is not expected here."),
			},
		},
	})
}

// This test will try to create a security policy with "build_failure_grace_period_in_days" set,
// but with "fail_build" set to false, which conflicts with the field mentioned above.
func TestAccSecurityPolicy_badGracePeriod(t *testing.T) {
	policyName := "terraform-security-policy-3"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule-3"
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_security_policy." + resourceName
	cvssFrom := float64(1)
	cvssTo := float64(5)
	failBuild := "false"
	gracePeriod := 5
	blockUnscanned := "true"
	blockActive := "true"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckSecurityPolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXraySecurityPolicy_createSecurityPolicyCVSS(resourceName, policyName, policyDesc,
					ruleName, cvssFrom, cvssTo, failBuild, gracePeriod, blockUnscanned, blockActive),
				ExpectError: regexp.MustCompile("Rule " + ruleName + " has failure grace period without fail build"),
			},
		},
	})
}

func TestAccSecurityPolicy_createBlockDownloadTrueCVSS(t *testing.T) {
	policyName := "terraform-security-policy-4"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule-4"
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_security_policy." + resourceName
	cvssFrom := float64(1)
	cvssTo := float64(5)
	failBuild := "true"
	gracePeriod := 5
	blockUnscanned := "true"
	blockActive := "true"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckSecurityPolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXraySecurityPolicy_createSecurityPolicyCVSS(resourceName, policyName, policyDesc,
					ruleName, cvssFrom, cvssTo, failBuild, gracePeriod, blockUnscanned, blockActive),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", policyDesc),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", strconv.Itoa(gracePeriod)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", failBuild),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", blockActive),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", blockUnscanned),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.cvss_range.0.from", fmt.Sprint(cvssFrom)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.cvss_range.0.to", fmt.Sprint(cvssTo)),
				),
			},
		},
	})
}

func TestAccSecurityPolicy_createBlockDownloadFalseCVSS(t *testing.T) {
	policyName := "terraform-security-policy-5"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule-5"
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_security_policy." + resourceName
	cvssFrom := float64(1)
	cvssTo := float64(5)
	failBuild := "true"
	gracePeriod := 5
	blockUnscanned := "false"
	blockActive := "false"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckSecurityPolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXraySecurityPolicy_createSecurityPolicyCVSS(resourceName, policyName, policyDesc,
					ruleName, cvssFrom, cvssTo, failBuild, gracePeriod, blockUnscanned, blockActive),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", policyDesc),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", strconv.Itoa(gracePeriod)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", failBuild),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", blockActive),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", blockUnscanned),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.cvss_range.0.from", fmt.Sprint(cvssFrom)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.cvss_range.0.to", fmt.Sprint(cvssTo)),
				),
			},
		},
	})
}

func TestAccSecurityPolicy_createBlockDownloadTrueMinSeverity(t *testing.T) {
	policyName := "terraform-security-policy-6"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule-6"
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_security_policy." + resourceName
	minSeverity := "Medium"
	failBuild := "true"
	gracePeriod := 5
	blockUnscanned := "true"
	blockActive := "true"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckSecurityPolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXraySecurityPolicy_createSecurityPolicyMinSeverity(resourceName, policyName, policyDesc,
					ruleName, minSeverity, failBuild, gracePeriod, blockUnscanned, blockActive),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", policyDesc),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", strconv.Itoa(gracePeriod)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", failBuild),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", blockActive),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", blockUnscanned),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.min_severity", minSeverity),
				),
			},
		},
	})
}

func TestAccSecurityPolicy_createBlockDownloadFalseMinSeverity(t *testing.T) {
	policyName := "terraform-security-policy-7"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule-7"
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_security_policy." + resourceName
	minSeverity := "Medium"
	failBuild := "true"
	gracePeriod := 5
	blockUnscanned := "false"
	blockActive := "false"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckSecurityPolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXraySecurityPolicy_createSecurityPolicyMinSeverity(resourceName, policyName, policyDesc,
					ruleName, minSeverity, failBuild, gracePeriod, blockUnscanned, blockActive),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", policyDesc),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", strconv.Itoa(gracePeriod)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", failBuild),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", blockActive),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", blockUnscanned),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.min_severity", minSeverity),
				),
			},
		},
	})
}

func TestAccSecurityPolicy_createCVSSFloat(t *testing.T) {
	policyName := "terraform-security-policy-8"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule-8"
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_security_policy." + resourceName
	cvssFrom := 1.5
	cvssTo := 5.3
	failBuild := "true"
	gracePeriod := 5
	blockUnscanned := "true"
	blockActive := "true"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckSecurityPolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXraySecurityPolicy_createSecurityPolicyCVSS(resourceName, policyName, policyDesc,
					ruleName, cvssFrom, cvssTo, failBuild, gracePeriod, blockUnscanned, blockActive),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", policyDesc),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", strconv.Itoa(gracePeriod)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", failBuild),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", blockActive),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", blockUnscanned),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.cvss_range.0.from", fmt.Sprint(cvssFrom)),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.cvss_range.0.to", fmt.Sprint(cvssTo)),
				),
			},
		},
	})
}

func testAccCheckSecurityPolicyDestroy(id string) func(*terraform.State) error {
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

func testAccXraySecurityPolicy_badSecurityType(name, description, ruleName string, rangeTo int) string {
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

func testAccXraySecurityPolicy_badSecurity(name, description, ruleName, allowedLicense string) string {
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

func testAccXraySecurityPolicy_createSecurityPolicyCVSS(resourceName, name, description, ruleName string,
	cvssFrom float64, cvssTo float64, failBuild string, gracePeriod int, blockUnscanned string,
	blockActive string) string {
	return fmt.Sprintf(`
resource "xray_security_policy" "%s" {
	name = "%s"
	description = "%s"
	type = "security"
	rules {
		name = "%s"
		priority = 1
		criteria {	
			cvss_range {
				from = %f
				to = %f
			}
		}
		actions {
			fail_build = %s
			build_failure_grace_period_in_days = %d
			block_download {
				unscanned = %s
				active = %s
			}
		}
	}
}
`, resourceName, name, description, ruleName, cvssFrom, cvssTo, failBuild, gracePeriod, blockUnscanned, blockActive)
}

func testAccXraySecurityPolicy_createSecurityPolicyMinSeverity(resourceName, name, description, ruleName string,
	minSeverity string, failBuild string, gracePeriod int, blockUnscanned string,
	blockActive string) string {
	return fmt.Sprintf(`
resource "xray_security_policy" "%s" {
	name = "%s"
	description = "%s"
	type = "security"
	rules {
		name = "%s"
		priority = 1
		criteria {
			min_severity = "%s"
		}
		actions {
			fail_build = %s
			build_failure_grace_period_in_days = %d
			block_download {
				unscanned = %s
				active = %s
			}
		}
	}
}
`, resourceName, name, description, ruleName, minSeverity, failBuild, gracePeriod, blockUnscanned, blockActive)
}
