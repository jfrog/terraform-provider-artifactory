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

var tempStructLicense = map[string]string{
	"resource_name":                     "",
	"policy_name":                       "terraform-license-policy",
	"policy_description":                "policy created by xray acceptance tests",
	"rule_name":                         "test-license-rule",
	"license_0":                         "Apache-1.0",
	"license_1":                         "Apache-2.0",
	"mails_0":                           "test0@email.com",
	"mails_1":                           "test1@email.com",
	"allow_unknown":                     "true",
	"multi_license_permissive":          "false",
	"block_release_bundle_distribution": "true",
	"fail_build":                        "true",
	"notify_watch_recipients":           "true",
	"notify_deployer":                   "true",
	"create_ticket_enabled":             "false",
	"custom_severity":                   "High",
	"grace_period_days":                 "5",
	"block_unscanned":                   "true",
	"block_active":                      "true",
}

// License policy criteria are different from the security policy criteria
// Test will try to post a new license policy with incorrect body of security policy
// with specified cvss_range. The function expandLicenseCriteria will ignore all the
// fields except of "allow_unknown", "banned_licenses" and "allowed_licenses" if the Policy type is "license"
func TestAccLicensePolicy_badLicenseCriteria(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_license_policy")
	policyName := "terraform-license-policy-1"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-license-rule-1"
	rangeTo := 5

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
	_, fqrn, resourceName := mkNames("policy-", "xray_license_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructLicense, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = "terraform-security-policy-2"
	tempStruct["rule_name"] = "test-license-rule-2"
	tempStruct["fail_build"] = "false"
	tempStruct["grace_period_days"] = "5"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      executeTemplate(fqrn, licensePolicyAllowed, tempStruct),
				ExpectError: regexp.MustCompile("Rule " + tempStruct["rule_name"] + " has failure grace period without fail build"),
			},
		},
	})
}

func TestAccLicensePolicy_createAllowedLic(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_license_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructLicense, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = "terraform-license-policy-3"
	tempStruct["rule_name"] = "test-license-rule-3"
	tempStruct["multi_license_permissive"] = "true"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, licensePolicyAllowed, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", tempStruct["policy_name"]),
					resource.TestCheckResourceAttr(fqrn, "description", tempStruct["policy_description"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", tempStruct["rule_name"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allow_unknown", tempStruct["allow_unknown"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.multi_license_permissive", tempStruct["multi_license_permissive"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allowed_licenses.0", tempStruct["license_0"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allowed_licenses.1", tempStruct["license_1"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.mails.0", tempStruct["mails_0"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.mails.1", tempStruct["mails_1"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_release_bundle_distribution", tempStruct["block_release_bundle_distribution"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", tempStruct["fail_build"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.notify_watch_recipients", tempStruct["notify_watch_recipients"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.notify_deployer", tempStruct["notify_deployer"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.create_ticket_enabled", tempStruct["create_ticket_enabled"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", tempStruct["grace_period_days"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", tempStruct["block_active"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", tempStruct["block_unscanned"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.custom_severity", tempStruct["custom_severity"]),
				),
			},
		},
	})
}

func TestAccLicensePolicy_createBannedLic(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_license_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructLicense, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = "terraform-license-policy-4"
	tempStruct["rule_name"] = "test-license-rule-4"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, licensePolicyBanned, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", tempStruct["policy_name"]),
					resource.TestCheckResourceAttr(fqrn, "description", tempStruct["policy_description"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", tempStruct["rule_name"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allow_unknown", tempStruct["allow_unknown"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.multi_license_permissive", tempStruct["multi_license_permissive"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.banned_licenses.0", tempStruct["license_0"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.banned_licenses.1", tempStruct["license_1"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.mails.0", tempStruct["mails_0"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.mails.1", tempStruct["mails_1"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_release_bundle_distribution", tempStruct["block_release_bundle_distribution"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", tempStruct["fail_build"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.notify_watch_recipients", tempStruct["notify_watch_recipients"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.notify_deployer", tempStruct["notify_deployer"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.create_ticket_enabled", tempStruct["create_ticket_enabled"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", tempStruct["grace_period_days"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", tempStruct["block_active"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", tempStruct["block_unscanned"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.custom_severity", tempStruct["custom_severity"]),
				),
			},
		},
	})
}

func TestAccLicensePolicy_createMultiLicensePermissiveFalse(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_license_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructLicense, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = "terraform-license-policy-5"
	tempStruct["rule_name"] = "test-license-rule-5"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, licensePolicyBanned, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", tempStruct["policy_name"]),
					resource.TestCheckResourceAttr(fqrn, "description", tempStruct["policy_description"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", tempStruct["rule_name"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allow_unknown", tempStruct["allow_unknown"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.multi_license_permissive", tempStruct["multi_license_permissive"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.banned_licenses.0", tempStruct["license_0"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.banned_licenses.1", tempStruct["license_1"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.mails.0", tempStruct["mails_0"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.mails.1", tempStruct["mails_1"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_release_bundle_distribution", tempStruct["block_release_bundle_distribution"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", tempStruct["fail_build"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.notify_watch_recipients", tempStruct["notify_watch_recipients"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.notify_deployer", tempStruct["notify_deployer"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.create_ticket_enabled", tempStruct["create_ticket_enabled"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", tempStruct["grace_period_days"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", tempStruct["block_active"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", tempStruct["block_unscanned"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.custom_severity", tempStruct["custom_severity"]),
				),
			},
		},
	})
}

func TestAccLicensePolicy_createBlockFalse(t *testing.T) {
	_, fqrn, resourceName := mkNames("policy-", "xray_license_policy")
	tempStruct := make(map[string]string)
	copyStringMap(tempStructLicense, tempStruct)

	tempStruct["resource_name"] = resourceName
	tempStruct["policy_name"] = "terraform-license-policy-6"
	tempStruct["rule_name"] = "test-license-rule-6"
	tempStruct["block_unscanned"] = "true"
	tempStruct["block_active"] = "true"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckLicensePolicyDestroy(fqrn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, licensePolicyBanned, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", tempStruct["policy_name"]),
					resource.TestCheckResourceAttr(fqrn, "description", tempStruct["policy_description"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.name", tempStruct["rule_name"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.allow_unknown", tempStruct["allow_unknown"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.multi_license_permissive", tempStruct["multi_license_permissive"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.banned_licenses.0", tempStruct["license_0"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.criteria.0.banned_licenses.1", tempStruct["license_1"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.mails.0", tempStruct["mails_0"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.mails.1", tempStruct["mails_1"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_release_bundle_distribution", tempStruct["block_release_bundle_distribution"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.fail_build", tempStruct["fail_build"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.notify_watch_recipients", tempStruct["notify_watch_recipients"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.notify_deployer", tempStruct["notify_deployer"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.create_ticket_enabled", tempStruct["create_ticket_enabled"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.build_failure_grace_period_in_days", tempStruct["grace_period_days"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.active", tempStruct["block_active"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.block_download.0.unscanned", tempStruct["block_unscanned"]),
					resource.TestCheckResourceAttr(fqrn, "rules.0.actions.0.custom_severity", tempStruct["custom_severity"]),
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

const licensePolicyAllowed = `resource "xray_license_policy" "{{ .resource_name }}" {
	name = "{{ .policy_name }}"
	description = "{{ .policy_description }}"
	type = "license"
	rules {
		name = "{{ .rule_name }}"
		priority = 1
		criteria {	
          allowed_licenses = ["{{ .license_0 }}","{{ .license_1 }}"]
          allow_unknown = {{ .allow_unknown }}
          multi_license_permissive = {{ .multi_license_permissive }}
		}
		actions {
          webhooks = []
          mails = ["{{ .mails_0 }}", "{{ .mails_1 }}"]
          block_download {
				unscanned = {{ .block_unscanned }}
				active = {{ .block_active }}
          }
          block_release_bundle_distribution = {{ .block_release_bundle_distribution }}
          fail_build = {{ .fail_build }}
          notify_watch_recipients = {{ .notify_watch_recipients }}
          notify_deployer = {{ .notify_deployer }}
          create_ticket_enabled = {{ .create_ticket_enabled }}           
          custom_severity = "{{ .custom_severity }}"
          build_failure_grace_period_in_days = {{ .grace_period_days }}  
		}
	}
}`

const licensePolicyBanned = `resource "xray_license_policy" "{{ .resource_name }}" {
	name = "{{ .policy_name }}"
	description = "{{ .policy_description }}"
	type = "license"
	rules {
		name = "{{ .rule_name }}"
		priority = 1
		criteria {	
          banned_licenses = ["{{ .license_0 }}","{{ .license_1 }}"]
          allow_unknown = {{ .allow_unknown }}
          multi_license_permissive = {{ .multi_license_permissive }}
		}
		actions {
          webhooks = []
          mails = ["{{ .mails_0 }}", "{{ .mails_1 }}"]
          block_download {
				unscanned = {{ .block_unscanned }}
				active = {{ .block_active }}
          }
          block_release_bundle_distribution = {{ .block_release_bundle_distribution }}
          fail_build = {{ .fail_build }}
          notify_watch_recipients = {{ .notify_watch_recipients }}
          notify_deployer = {{ .notify_deployer }}
          create_ticket_enabled = {{ .create_ticket_enabled }}           
          custom_severity = "{{ .custom_severity }}"
          build_failure_grace_period_in_days = {{ .grace_period_days }}  
		}
	}
}`
