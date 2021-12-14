package xray

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var tempStructSecurity = map[string]string{
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
				ExpectError: regexp.MustCompile("Rule " + ruleName + " has empty criteria"),
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
	tempStruct := make(map[string]string)
	copyStringMap(tempStructSecurity, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = fmt.Sprintf("terraform-security-policy-3-%d", randomInt())
	tempStruct["rule_name"] = fmt.Sprintf("test-security-rule-3-%d", randomInt())
	tempStruct["fail_build"] = "false"
	tempStruct["grace_period_days"] = "5"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config:      executeTemplate(fqrn, securityPolicyCVSS, tempStruct),
				ExpectError: regexp.MustCompile("Rule " + tempStruct["rule_name"] + " has failure grace period without fail build"),
			},
		},
	})
}

// CVSS criteria, block downloading of unscanned and active
func TestAccSecurityPolicy_createBlockDownloadTrueCVSS(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructSecurity, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = fmt.Sprintf("terraform-security-policy-4-%d", randomInt())
	tempStruct["rule_name"] = fmt.Sprintf("test-security-rule-4-%d", randomInt())
	tempStruct["cvssOrSeverity"] = "cvss"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyCVSS, tempStruct),
				Check:  verifySecurityPolicy(fqrn, tempStruct, tempStruct["cvssOrSeverity"]),
			},
		},
	})
}

// CVSS criteria, allow downloading of unscanned and active
func TestAccSecurityPolicy_createBlockDownloadFalseCVSS(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructSecurity, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = fmt.Sprintf("terraform-security-policy-5-%d", randomInt())
	tempStruct["rule_name"] = fmt.Sprintf("test-security-rule-5-%d", randomInt())
	tempStruct["block_unscanned"] = "false"
	tempStruct["block_active"] = "false"
	tempStruct["cvssOrSeverity"] = "cvss"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyCVSS, tempStruct),
				Check:  verifySecurityPolicy(fqrn, tempStruct, tempStruct["cvssOrSeverity"]),
			},
		},
	})
}

// Min severity criteria, block downloading of unscanned and active
func TestAccSecurityPolicy_createBlockDownloadTrueMinSeverity(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructSecurity, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = fmt.Sprintf("terraform-security-policy-6-%d", randomInt())
	tempStruct["rule_name"] = fmt.Sprintf("test-security-rule-6-%d", randomInt())
	tempStruct["cvssOrSeverity"] = "severity"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyMinSeverity, tempStruct),
				Check:  verifySecurityPolicy(fqrn, tempStruct, tempStruct["cvssOrSeverity"]),
			},
		},
	})
}

// Min severity criteria, allow downloading of unscanned and active
func TestAccSecurityPolicy_createBlockDownloadFalseMinSeverity(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructSecurity, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = fmt.Sprintf("terraform-security-policy-7-%d", randomInt())
	tempStruct["rule_name"] = fmt.Sprintf("test-security-rule-7-%d", randomInt())
	tempStruct["block_unscanned"] = "false"
	tempStruct["block_active"] = "false"
	tempStruct["cvssOrSeverity"] = "severity"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyMinSeverity, tempStruct),
				Check:  verifySecurityPolicy(fqrn, tempStruct, tempStruct["cvssOrSeverity"]),
			},
		},
	})
}

// CVSS criteria, use float values for CVSS range
func TestAccSecurityPolicy_createCVSSFloat(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructSecurity, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = fmt.Sprintf("terraform-security-policy-8-%d", randomInt())
	tempStruct["rule_name"] = fmt.Sprintf("test-security-rule-8-%d", randomInt())
	tempStruct["cvss_from"] = "1.5"
	tempStruct["cvss_to"] = "5.3"
	tempStruct["cvssOrSeverity"] = "cvss"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyCVSS, tempStruct),
				Check:  verifySecurityPolicy(fqrn, tempStruct, tempStruct["cvssOrSeverity"]),
			},
		},
	})
}

func TestAccSecurityPolicy_blockMismatchCVSS(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_security_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructSecurity, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = fmt.Sprintf("terraform-security-policy-9-%d", randomInt())
	tempStruct["rule_name"] = fmt.Sprintf("test-security-rule-9-%d", randomInt())
	tempStruct["block_unscanned"] = "true"
	tempStruct["block_active"] = "false"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckPolicy),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, securityPolicyCVSS, tempStruct),
				ExpectError: regexp.MustCompile("Rule " + tempStruct["rule_name"] +
					" has block unscanned without block download"),
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

func verifySecurityPolicy(fqrn string, tempStruct map[string]string, cvssOrSeverity string) resource.TestCheckFunc {
	var commonCheckList = resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(fqrn, "name", tempStruct["policy_name"]),
		resource.TestCheckResourceAttr(fqrn, "description", tempStruct["policy_description"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.name", tempStruct["rule_name"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.block_release_bundle_distribution", tempStruct["block_release_bundle_distribution"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.fail_build", tempStruct["fail_build"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.notify_watch_recipients", tempStruct["notify_watch_recipients"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.notify_deployer", tempStruct["notify_deployer"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.create_ticket_enabled", tempStruct["create_ticket_enabled"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.build_failure_grace_period_in_days", tempStruct["grace_period_days"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.block_download.0.active", tempStruct["block_active"]),
		resource.TestCheckResourceAttr(fqrn, "rule.0.actions.0.block_download.0.unscanned", tempStruct["block_unscanned"]),
	)
	if cvssOrSeverity == "cvss" {
		return resource.ComposeTestCheckFunc(
			commonCheckList,
			resource.TestCheckResourceAttr(fqrn, "rule.0.criteria.0.cvss_range.0.from", tempStruct["cvss_from"]),
			resource.TestCheckResourceAttr(fqrn, "rule.0.criteria.0.cvss_range.0.to", tempStruct["cvss_to"]),
		)
	}
	if cvssOrSeverity == "severity" {
		return resource.ComposeTestCheckFunc(
			commonCheckList,
			resource.TestCheckResourceAttr(fqrn, "rule.0.criteria.0.min_severity", tempStruct["min_severity"]),
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
