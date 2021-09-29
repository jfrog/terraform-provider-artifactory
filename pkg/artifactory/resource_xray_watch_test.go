package artifactory

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Test was temporarily removed from the test suite
//func TestAccWatch_basic(t *testing.T) {
//	watchName := "test-watch"
//	policyName := "test-policy"
//	watchDesc := "watch created by xray acceptance tests"
//	resourceName := "xray_watch.test"
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		CheckDestroy: testAccCheckWatchDestroy,
//		Providers:    testAccProviders,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccXrayWatchBasic(watchName, watchDesc, policyName),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr(resourceName, "name", watchName),
//					resource.TestCheckResourceAttr(resourceName, "description", watchDesc),
//					resource.TestCheckResourceAttr(resourceName, "resources.0.type", "all-repos"),
//					resource.TestCheckResourceAttr(resourceName, "assigned_policies.0.name", policyName),
//					resource.TestCheckResourceAttr(resourceName, "assigned_policies.0.type", "security"),
//				),
//			},
//			{
//				ResourceName:      resourceName,
//				ImportState:       true,
//				ImportStateVerify: false,
//			},
//			{
//				Config: testAccXrayWatchUnassigned(policyName),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckWatchDoesntExist(resourceName),
//				),
//			},
//		},
//	})
//}

// These two tests are commented out because repoName and binMgrId must be real values but neither are terraformable so can't be put into these tests
// I have tested this with some real values, but for obvious privacy reasons am not leaving those real values in here
/*func TestAccWatch_filters(t *testing.T) {
	watchName := "test-watch"
	watchDesc := "watch created by xray acceptance tests"
	repoName := "repo-name"
	binMgrId := "artifactory-id"
	policyName := "test-policy"
	filterValue := "Debian"
	updatedDesc := "updated watch description"
	updatedValue := "Docker"
	resourceName := "xray_watch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy,
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayWatchFilters(watchName, watchDesc, repoName, binMgrId, policyName, filterValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", watchName),
					resource.TestCheckResourceAttr(resourceName, "description", watchDesc),
					resource.TestCheckResourceAttr(resourceName, "resources.0.filters.0.type", "package-type"),
					resource.TestCheckResourceAttr(resourceName, "resources.0.filters.0.value", filterValue),
					resource.TestCheckResourceAttr(resourceName, "resources.0.type", "repository"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: testAccXrayWatchFilters(watchName, updatedDesc, repoName, binMgrId, policyName, updatedValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", watchName),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDesc),
					resource.TestCheckResourceAttr(resourceName, "resources.0.filters.0.type", "package-type"),
					resource.TestCheckResourceAttr(resourceName, "resources.0.filters.0.value", updatedValue),
					resource.TestCheckResourceAttr(resourceName, "resources.0.type", "repository"),
				),
			},
			{
				Config: testAccXrayWatchUnassigned(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWatchDoesntExist(resourceName),
				),
			},
		},
	})
}

func TestAccWatch_builds(t *testing.T) {
	watchName := "test-watch"
	policyName := "test-policy"
	watchDesc := "watch created by xray acceptance tests"
	binMgrId := "artifactory-id"
	resourceName := "xray_watch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy,
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayWatchBuilds(watchName, watchDesc, policyName, binMgrId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", watchName),
					resource.TestCheckResourceAttr(resourceName, "description", watchDesc),
					resource.TestCheckResourceAttr(resourceName, "resources.0.type", "all-builds"),
					resource.TestCheckResourceAttr(resourceName, "assigned_policies.0.name", policyName),
					resource.TestCheckResourceAttr(resourceName, "assigned_policies.0.type", "security"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: testAccXrayWatchUnassigned(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWatchDoesntExist(resourceName),
				),
			},
		},
	})
}*/

func testAccCheckWatchDoesntExist(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if ok {
			return fmt.Errorf("watch %s exists when it shouldn't", resourceName)
		}
		return nil
	}
}

func testAccCheckWatchDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*resty.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "xray_watch" {
			watch := Watch{}
			resp, err := client.R().SetResult(&watch).Get("xray/api/v2/watches/" + rs.Primary.ID)
			if err != nil {
				if resp != nil && resp.StatusCode() == http.StatusNotFound {
					continue
				}
				return err
			}

			return fmt.Errorf("error: Watch %s still exists %s", rs.Primary.ID, *watch.GeneralData.Name)

		}
		if rs.Type == "xray_policy" {
			policy, resp, err := getPolicy(rs.Primary.ID, client)

			if err != nil {
				if resp != nil && resp.StatusCode() == http.StatusInternalServerError &&
					err.Error() != fmt.Sprintf("{\"error\":\"Failed to find Policy %s\"}", rs.Primary.ID) {
					continue
				}
				return err
			}
			return fmt.Errorf("error: Policy %s still exists %s", rs.Primary.ID, *policy.Name)
		}
	}

	return nil
}

func testAccXrayWatchBasic(name, description, policyName string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
	name  = "%s"
	description = "test policy description"
	type = "security"

	rules {
		name = "rule-name"
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

resource "xray_watch" "test" {
	name  = "%s"
	description = "%s"
	resources {
		type = "all-repos"
		name = "All Repositories"
	}
	assigned_policies {
		name = xray_policy.test.name
		type = "security"
	}
	watch_recipients = ["test@example.com"]
}
`, policyName, name, description)
}

// Since policies can't be deleted if they have a watch assigned, we need to force terraform to delete the watch first
// by removing it from the code at the end of every test step
func testAccXrayWatchUnassigned(policyName string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
	name  = "%s"
	description = "test policy description"
	type = "security"

	rules {
		name = "rule-name"
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
`, policyName)
}

// You seemingly can't do filters with all-repos - it's an example in the docs but doesn't seem possible via the web ui
//func testAccXrayWatchFilters(name, description, repoName, binMgrId, policyName, filterValue string) string {
//	return fmt.Sprintf(`
//resource "xray_policy" "test" {
//	name  = "%s"
//	description = "test policy description"
//	type = "security"
//
//	rules {
//		name = "rule-name"
//		priority = 1
//		criteria {
//			min_severity = "High"
//		}
//		actions {
//			block_download {
//				unscanned = true
//				active = true
//			}
//		}
//	}
//}
//
//resource "xray_watch" "test" {
//	name  = "%s"
//	description = "%s"
//	resources {
//		type = "repository"
//		name = "%s"
//		bin_mgr_id = "%s"
//		filters {
//			type = "package-type"
//			value = "%s"
//		}
//	}
//	assigned_policies {
//		name = xray_policy.test.name
//		type = "security"
//	}
//}
//`, policyName, name, description, repoName, binMgrId, filterValue)
//}
//
//func testAccXrayWatchBuilds(name, description, policyName, binMgrId string) string {
//	return fmt.Sprintf(`
//resource "xray_policy" "test" {
//	name  = "%s"
//	description = "test policy description"
//	type = "security"
//
//	rules {
//		name = "rule-name"
//		priority = 1
//		criteria {
//			min_severity = "High"
//		}
//		actions {
//			block_download {
//				unscanned = true
//				active = true
//			}
//		}
//	}
//}
//
//resource "xray_watch" "test" {
//	name = "%s"
//	description = "%s"
//	resources {
//		type = "all-builds"
//		name = "All Builds"
//		bin_mgr_id = "%s"
//	}
//	assigned_policies {
//		name = xray_policy.test.name
//		type = "security"
//	}
//}
//`, policyName, name, description, binMgrId)
//}

// TODO for bonus points - test builds with complex filters eg "filters":[{"type":"ant-patterns","value":{"ExcludePatterns":[],"IncludePatterns":["*"]}
