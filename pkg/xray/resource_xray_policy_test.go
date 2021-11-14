package xray

import (
	"fmt"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory"
	"net/http"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Test was temporarily removed from the test suite
//func TestAccPolicy_basic(t *testing.T) {
//	policyName := "terraform-test-policy"
//	policyDesc := "policy created by xray acceptance tests"
//	ruleName := "test-security-rule"
//	resourceName := "xray_policy.test"
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		CheckDestroy: testAccCheckPolicyDestroy,
//		Providers:    TestAccProviders,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccXrayPolicyBasic(policyName, policyDesc, ruleName),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr(resourceName, "name", policyName),
//					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.min_severity", "High"),
//				),
//			},
//			{
//				ResourceName:      resourceName,
//				ImportState:       true,
//				ImportStateVerify: false, // TODO: figure out why this doesn't work
//			},
//		},
//	})
//}

// Test was temporarily removed from the test suite
//func TestAccPolicy_cvssRange(t *testing.T) {
//	policyName := "terraform-test-policy"
//	policyDesc := "policy created by xray acceptance tests"
//	ruleName := "test-security-rule"
//	rangeTo := 4
//	updatedRangeTo := 2
//	resourceName := "xray_policy.test"
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		CheckDestroy: testAccCheckPolicyDestroy,
//		Providers:    TestAccProviders,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccXrayPolicyCVSSRange(policyName, policyDesc, ruleName, rangeTo),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr(resourceName, "name", policyName),
//					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.cvss_range.0.to", fmt.Sprintf("%d", rangeTo)),
//				),
//			},
//			{
//				ResourceName:      resourceName,
//				ImportState:       true,
//				ImportStateVerify: false,
//			},
//			{
//				Config: testAccXrayPolicyCVSSRange(policyName, policyDesc, ruleName, updatedRangeTo),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr(resourceName, "name", policyName),
//					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.cvss_range.0.to", fmt.Sprintf("%d", updatedRangeTo)),
//				),
//			},
//		},
//	})
//}

// Test was temporarily removed from the test suite
//func TestAccPolicy_allActions(t *testing.T) {
//	policyName := "terraform-test-policy"
//	policyDesc := "policy created by xray acceptance tests"
//	ruleName := "test-security-rule"
//	actionMail := "test@example.com"
//	updatedDesc := "updated policy description"
//	updatedRuleName := "test-updated-rule"
//	updatedMail := "test2@example.com"
//	resourceName := "xray_policy.test"
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		CheckDestroy: testAccCheckPolicyDestroy,
//		Providers:    TestAccProviders,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccXrayPolicyAllActions(policyName, policyDesc, ruleName, actionMail),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr(resourceName, "name", policyName),
//					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.min_severity", "High"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0.fail_build", "true"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0.mails.0", actionMail),
//				),
//			},
//			{
//				ResourceName:      resourceName,
//				ImportState:       true,
//				ImportStateVerify: false,
//			},
//			{
//				Config: testAccXrayPolicyAllActions(policyName, updatedDesc, updatedRuleName, updatedMail),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr(resourceName, "name", policyName),
//					resource.TestCheckResourceAttr(resourceName, "description", updatedDesc),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.name", updatedRuleName),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.min_severity", "High"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0.fail_build", "true"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0.mails.0", updatedMail),
//				),
//			},
//		},
//	})
//}
// Test was temporarily removed from the test suite
//func TestAccPolicy_licenseCriteria(t *testing.T) {
//	policyName := "terraform-test-policy"
//	policyDesc := "policy created by xray acceptance tests"
//	ruleName := "test-security-rule"
//	allowedLicense := "BSD-4-Clause"
//	bannedLicense1 := "0BSD"
//	bannedLicense2 := "diffmark"
//	resourceName := "xray_policy.test"
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		CheckDestroy: testAccCheckPolicyDestroy,
//		Providers:    TestAccProviders,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccXrayPolicyLicense(policyName, policyDesc, ruleName, allowedLicense),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr(resourceName, "name", policyName),
//					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.allowed_licenses.0", allowedLicense),
//				),
//			},
//			{
//				ResourceName:      resourceName,
//				ImportState:       true,
//				ImportStateVerify: false,
//			},
//			{
//				Config: testAccXrayPolicyLicenseBanned(policyName, policyDesc, ruleName, bannedLicense1, bannedLicense2),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr(resourceName, "name", policyName),
//					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.banned_licenses.0", bannedLicense1),
//					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.banned_licenses.1", bannedLicense2),
//				),
//			},
//		},
//	})
//}

func TestAccPolicy_badLicenseCriteria(t *testing.T) {
	policyName := "terraform-test-policy"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule"
	rangeTo := 4

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { artifactory.testAccPreCheck(t) },
		CheckDestroy:      testAccCheckPolicyDestroy,
		ProviderFactories: artifactory.TestAccProviders,
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
		PreCheck:          func() { artifactory.testAccPreCheck(t) },
		CheckDestroy:      testAccCheckPolicyDestroy,
		ProviderFactories: artifactory.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccXrayPolicy_badSecurity(policyName, policyDesc, ruleName, allowedLicense),
				ExpectError: regexp.MustCompile("allow_unknown, banned_licenses, and allowed_licenses are not supported with security policies"),
			},
		},
	})
}

// This should be uncommented when someone figures out how to deal with this (see comment in expandActions in the provider)
/*func TestAccPolicy_missingBlockDownloads(t *testing.T) {
	policyName   := "terraform-test-policy"
	policyDesc   := "policy created by xray acceptance tests"
	ruleName     := "test-security-rule"
	resourceName := "xray_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckPolicyDestroy,
		Providers:    TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayPolicy_missingBlockDownloads(policyName, policyDesc, ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}*/

func testAccCheckPolicyDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "xray_policy" {
			provider, _ := artifactory.TestAccProviders["artifactory"]()
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

func testAccXrayPolicyBasic(name, description, ruleName string) string {
	return fmt.Sprintf(`
		resource "artifactory_xray_policy" "test" {
			name  = "%s"
			description = "%s"
			type = "security"
	
			rules {
				name = "%s"
				priority = 1
				criteria {
					min_severity = "High"
				}
				actions {
					block_download {
						unscanned = true
						active = true
					}
				}
			}
		}
`, name, description, ruleName)
}

func testAccXrayPolicyCVSSRange(name, description, ruleName string, rangeTo int) string {
	return fmt.Sprintf(`
		resource "artifactory_xray_policy" "test" {
			name  = "%s"
			description = "%s"
			type = "security"
		
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

func testAccXrayPolicyAllActions(name, description, ruleName, email string) string {
	// Except for webhooks, because the API won't let you test with junk urls: Error: {"error":"Rule test-security-rule triggers an unrecognized webhook https://example.com"}
	return fmt.Sprintf(`
		resource "artifactory_xray_policy" "test" {
			name  = "%s"
			description = "%s"
			type = "security"
		
			rules {
				name = "%s"
				priority = 1
				criteria {
					min_severity = "High"
				}
				actions {
					fail_build = true
					block_download {
						unscanned = false
						active = false
					}
					mails = ["%s"]
					custom_severity = "High"
				}
			}
		}
`, name, description, ruleName, email)
}

func testAccXrayPolicyLicense(name, description, ruleName, allowedLicense string) string {
	return fmt.Sprintf(`
		resource "artifactory_xray_policy" "test" {
			name = "%s"
			description = "%s"
			type = "license"
		
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

func testAccXrayPolicyLicenseBanned(name, description, ruleName, bannedLicense1, bannedLicense2 string) string {
	return fmt.Sprintf(`
		resource "artifactory_xray_policy" "test" {
			name = "%s"
			description = "%s"
			type = "license"
		
			rules {
				name = "%s"
				priority = 1
				criteria {
					allow_unknown = true
					banned_licenses = ["%s", "%s"]
				}
				actions {
					block_download {
						unscanned = true
						active = true
					}
				}
			}
		}
`, name, description, ruleName, bannedLicense1, bannedLicense2)
}

func testAccXrayPolicy_badLicense(name, description, ruleName string, rangeTo int) string {
	return fmt.Sprintf(`
resource "artifactory_xray_policy" "test" {
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
resource "artifactory_xray_policy" "test" {
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
