package xray

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testDataSecurity = map[string]string{
	"resource_name":                     "",
	"policy_name":                       "terraform-security-policy",
	"policy_description":                "policy created by xray acceptance tests",
	"rule_name":                         "test-security-rule",
	"cvss_from":                         "1",    // conflicts with min_severity
	"cvss_to":                           "5",    // conflicts with min_severity
	"min_severity":                      "High", // conflicts with cvss_from/cvss_to
	"block_release_bundle_distribution": "true",
	"fail_build":                        "true",
	"notify_watch_recipients":           "true",
	"notify_deployer":                   "true",
	"create_ticket_enabled":             "false",
	"grace_period_days":                 "5",
	"block_unscanned":                   "true",
	"block_active":                      "true",
	"cvssOrSeverity":                    "cvss",
}

// Teh test will try to create a security policy with the type of "license"
// The Policy criteria will be ignored in this case
func TestAccSecurityPolicy_badTypeInSecurityPolicy(t *testing.T) {
	policyName := fmt.Sprintf("terraform-security-policy-1-%d", randomInt())
	policyDesc := "policy created by xray acceptance tests"
	ruleName := fmt.Sprintf("test-security-rule-1-%d", randomInt())
	rangeTo := 5
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_security_policy." + resourceName
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config:      testAccXraySecurityPolicy_badSecurityType(policyName, policyDesc, ruleName, rangeTo),
				ExpectError: regexp.MustCompile("Found Invalid Policy"),
			},
		},
	})
}

// The test will try to use "allowed_licenses" in the security policy criteria
// That field is acceptable only in license policy. No API call, expected to fail on the TF resource verification
func TestAccSecurityPolicy_badSecurityCriteria(t *testing.T) {
	policyName := fmt.Sprintf("terraform-security-policy-2-%d", randomInt())
	policyDesc := "policy created by xray acceptance tests"
	ruleName := fmt.Sprintf("test-security-rule-2-%d", randomInt())
	allowedLicense := "BSD-4-Clause"
	resourceName := "policy-" + strconv.Itoa(randomInt())
	fqrn := "xray_security_policy." + resourceName
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
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
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	testData := make(map[string]string)
	copyStringMap(testDataSecurity, testData)

	testData["resource_name"] = resourceName
	testData["policy_name"] = fmt.Sprintf("terraform-security-policy-3-%d", randomInt())
	testData["rule_name"] = fmt.Sprintf("test-security-rule-3-%d", randomInt())
	testData["fail_build"] = "false"
	testData["grace_period_days"] = "5"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config:      executeTemplate(fqrn, securityPolicyCVSS, testData),
				ExpectError: regexp.MustCompile("Found Invalid Policy"),
			},
		},
	})
}

// CVSS criteria, block downloading of unscanned and active
func TestAccSecurityPolicy_createBlockDownloadTrueCVSS(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	testData := make(map[string]string)
	copyStringMap(testDataSecurity, testData)

	testData["resource_name"] = resourceName
	testData["policy_name"] = fmt.Sprintf("terraform-security-policy-4-%d", randomInt())
	testData["rule_name"] = fmt.Sprintf("test-security-rule-4-%d", randomInt())
	testData["cvssOrSeverity"] = "cvss"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyCVSS, testData),
				Check:  verifySecurityPolicy(fqrn, testData, testData["cvssOrSeverity"]),
			},
		},
	})
}

// CVSS criteria, allow downloading of unscanned and active
func TestAccSecurityPolicy_createBlockDownloadFalseCVSS(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	testData := make(map[string]string)
	copyStringMap(testDataSecurity, testData)

	testData["resource_name"] = resourceName
	testData["policy_name"] = fmt.Sprintf("terraform-security-policy-5-%d", randomInt())
	testData["rule_name"] = fmt.Sprintf("test-security-rule-5-%d", randomInt())
	testData["block_unscanned"] = "false"
	testData["block_active"] = "false"
	testData["cvssOrSeverity"] = "cvss"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyCVSS, testData),
				Check:  verifySecurityPolicy(fqrn, testData, testData["cvssOrSeverity"]),
			},
		},
	})
}

// Min severity criteria, block downloading of unscanned and active
func TestAccSecurityPolicy_createBlockDownloadTrueMinSeverity(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	testData := make(map[string]string)
	copyStringMap(testDataSecurity, testData)

	testData["resource_name"] = resourceName
	testData["policy_name"] = fmt.Sprintf("terraform-security-policy-6-%d", randomInt())
	testData["rule_name"] = fmt.Sprintf("test-security-rule-6-%d", randomInt())
	testData["cvssOrSeverity"] = "severity"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyMinSeverity, testData),
				Check:  verifySecurityPolicy(fqrn, testData, testData["cvssOrSeverity"]),
			},
		},
	})
}

// Min severity criteria, allow downloading of unscanned and active
func TestAccSecurityPolicy_createBlockDownloadFalseMinSeverity(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	testData := make(map[string]string)
	copyStringMap(testDataSecurity, testData)

	testData["resource_name"] = resourceName
	testData["policy_name"] = fmt.Sprintf("terraform-security-policy-7-%d", randomInt())
	testData["rule_name"] = fmt.Sprintf("test-security-rule-7-%d", randomInt())
	testData["block_unscanned"] = "false"
	testData["block_active"] = "false"
	testData["cvssOrSeverity"] = "severity"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyMinSeverity, testData),
				Check:  verifySecurityPolicy(fqrn, testData, testData["cvssOrSeverity"]),
			},
		},
	})
}

// CVSS criteria, use float values for CVSS range
func TestAccSecurityPolicy_createCVSSFloat(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	testData := make(map[string]string)
	copyStringMap(testDataSecurity, testData)

	testData["resource_name"] = resourceName
	testData["policy_name"] = fmt.Sprintf("terraform-security-policy-8-%d", randomInt())
	testData["rule_name"] = fmt.Sprintf("test-security-rule-8-%d", randomInt())
	testData["cvss_from"] = "1.5"
	testData["cvss_to"] = "5.3"
	testData["cvssOrSeverity"] = "cvss"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyCVSS, testData),
				Check:  verifySecurityPolicy(fqrn, testData, testData["cvssOrSeverity"]),
			},
		},
	})
}

func TestAccSecurityPolicy_blockMismatchCVSS(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	testData := make(map[string]string)
	copyStringMap(testDataSecurity, testData)

	testData["resource_name"] = resourceName
	testData["policy_name"] = fmt.Sprintf("terraform-security-policy-9-%d", randomInt())
	testData["rule_name"] = fmt.Sprintf("test-security-rule-9-%d", randomInt())
	testData["block_unscanned"] = "true"
	testData["block_active"] = "false"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config:      executeTemplate(fqrn, securityPolicyCVSS, testData),
				ExpectError: regexp.MustCompile("Found Invalid Policy"),
			},
		},
	})
}

func testAccXraySecurityPolicy_badSecurityType(name, description, ruleName string, rangeTo int) string {
	return fmt.Sprintf(`
resource "xray_security_policy" "test" {
	name = "%s"
	description = "%s"
	type = "license"
	rule {
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
	rule {
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

func verifySecurityPolicy(fqrn string, testData map[string]string, cvssOrSeverity string) resource.TestCheckFunc {
	var commonCheckList = resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(fqrn, "name", testData["policy_name"]),
		resource.TestCheckResourceAttr(fqrn, "description", testData["policy_description"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.name", testData["rule_name"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.block_release_bundle_distribution", testData["block_release_bundle_distribution"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.fail_build", testData["fail_build"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.notify_watch_recipients", testData["notify_watch_recipients"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.notify_deployer", testData["notify_deployer"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.create_ticket_enabled", testData["create_ticket_enabled"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.build_failure_grace_period_in_days", testData["grace_period_days"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.block_download.0.active", testData["block_active"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.block_download.0.unscanned", testData["block_unscanned"]),
	)
	if cvssOrSeverity == "cvss" {
		return resource.ComposeTestCheckFunc(
			commonCheckList,
			resource.TestCheckResourceAttr(fqrn, "rule.0.criteria.0.cvss_range.0.from", testData["cvss_from"]),
			resource.TestCheckResourceAttr(fqrn, "rule.0.criteria.0.cvss_range.0.to", testData["cvss_to"]),
		)
	}
	if cvssOrSeverity == "severity" {
		return resource.ComposeTestCheckFunc(
			commonCheckList,
			resource.TestCheckResourceAttr(fqrn, "rule.0.criteria.0.min_severity", testData["min_severity"]),
		)
	}
	return nil
}

const securityPolicyCVSS = `resource "xray_security_policy" "{{ .resource_name }}" {
	name = "{{ .policy_name }}"
	description = "{{ .policy_description }}"
	type = "security"
	rule {
		name = "{{ .rule_name }}"
		priority = 1
		criteria {	
			cvss_range {
				from = {{ .cvss_from }}
				to = {{ .cvss_to }}
			}
		}
		actions {
			block_release_bundle_distribution = {{ .block_release_bundle_distribution }}
			fail_build = {{ .fail_build }}
			notify_watch_recipients = {{ .notify_watch_recipients }}
			notify_deployer = {{ .notify_deployer }}
			create_ticket_enabled = {{ .create_ticket_enabled }}
			build_failure_grace_period_in_days = {{ .grace_period_days }}
			block_download {
				unscanned = {{ .block_unscanned }}
				active = {{ .block_active }}
			}
		}
	}
}`

const securityPolicyMinSeverity = `resource "xray_security_policy" "{{ .resource_name }}" {
	name = "{{ .policy_name }}"
	description = "{{ .policy_description }}"
	type = "security"
	rule {
		name = "{{ .rule_name }}"
		priority = 1
		criteria {
            min_severity = "{{ .min_severity }}"
		}
		actions {
			block_release_bundle_distribution = {{ .block_release_bundle_distribution }}
			fail_build = {{ .fail_build }}
			notify_watch_recipients = {{ .notify_watch_recipients }}
			notify_deployer = {{ .notify_deployer }}
			create_ticket_enabled = {{ .create_ticket_enabled }}
			build_failure_grace_period_in_days = {{ .grace_period_days }}
			block_download {
				unscanned = {{ .block_unscanned }}
				active = {{ .block_active }}
			}
		}
	}
}`
