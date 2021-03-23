package artifactory

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xero-oss/go-xray/xray"
)

func TestAccPolicy_basic(t *testing.T) {
	policyName := "terraform-test-policy"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule"
	resourceName := "xray_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckPolicyDestroy,
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayPolicy_basic(policyName, policyDesc, ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.min_severity", "High"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false, // TODO: figure out why this doesn't work
			},
		},
	})
}

func TestAccPolicy_cvssRange(t *testing.T) {
	policyName := "terraform-test-policy"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule"
	rangeTo := 4
	updatedRangeTo := 2
	resourceName := "xray_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckPolicyDestroy,
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayPolicy_cvssRange(policyName, policyDesc, ruleName, rangeTo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.cvss_range.0.to", fmt.Sprintf("%d", rangeTo)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: testAccXrayPolicy_cvssRange(policyName, policyDesc, ruleName, updatedRangeTo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.cvss_range.0.to", fmt.Sprintf("%d", updatedRangeTo)),
				),
			},
		},
	})
}

func TestAccPolicy_allActions(t *testing.T) {
	policyName := "terraform-test-policy"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule"
	actionMail := "test@example.com"
	updatedDesc := "updated policy description"
	updatedRuleName := "test-updated-rule"
	updatedMail := "test2@example.com"
	resourceName := "xray_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckPolicyDestroy,
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayPolicy_allActions(policyName, policyDesc, ruleName, actionMail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.min_severity", "High"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0.fail_build", "true"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0.mails.0", actionMail),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: testAccXrayPolicy_allActions(policyName, updatedDesc, updatedRuleName, updatedMail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", updatedRuleName),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.min_severity", "High"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0.fail_build", "true"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0.mails.0", updatedMail),
				),
			},
		},
	})
}

func TestAccPolicy_licenseCriteria(t *testing.T) {
	policyName := "terraform-test-policy"
	policyDesc := "policy created by xray acceptance tests"
	ruleName := "test-security-rule"
	allowedLicense := "BSD-4-Clause"
	bannedLicense1 := "0BSD"
	bannedLicense2 := "diffmark"
	resourceName := "xray_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckPolicyDestroy,
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayPolicy_license(policyName, policyDesc, ruleName, allowedLicense),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.allowed_licenses.0", allowedLicense),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: testAccXrayPolicy_licenseBanned(policyName, policyDesc, ruleName, bannedLicense1, bannedLicense2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.banned_licenses.0", bannedLicense1),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.0.banned_licenses.1", bannedLicense2),
				),
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
		Providers:    testAccProviders,
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
	conn := testAccProvider.Meta().(*xray.Xray)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "xray_policy" {
			continue
		}

		policy, resp, err := conn.V1.Policies.GetPolicy(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound {
			continue
		} else if resp.StatusCode == http.StatusInternalServerError && err.Error() == fmt.Sprintf("{\"error\":\"Failed to find Policy %s\"}", rs.Primary.ID) {
			continue
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Policy %s still exists %s", rs.Primary.ID, *policy.Name)
		}
	}
	return nil
}

func testAccXrayPolicy_basic(name, description, ruleName string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
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

func testAccXrayPolicy_cvssRange(name, description, ruleName string, rangeTo int) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
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

func testAccXrayPolicy_allActions(name, description, ruleName, email string) string {
	// Except for webhooks, because the API won't let you test with junk urls: Error: {"error":"Rule test-security-rule triggers an unrecognized webhook https://example.com"}
	return fmt.Sprintf(`
resource "xray_policy" "test" {
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

func testAccXrayPolicy_license(name, description, ruleName, allowedLicense string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
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

func testAccXrayPolicy_licenseBanned(name, description, ruleName, bannedLicense1, bannedLicense2 string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
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

func testAccXrayPolicy_missingBlockDownloads(name, description, ruleName string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
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
		}
	}
}
`, name, description, ruleName)
}
