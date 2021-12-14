package artifactory

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestInvalidCronFails(t *testing.T) {
	const invalidCron = `
		resource "artifactory_local_repository" "lib-local" {
			key = "lib-local"
			package_type = "maven"
		}

		resource "artifactory_replication_config" "lib-local" {
			repo_key = "${artifactory_local_repository.lib-local.key}"
			cron_exp = "0 0 blah foo boo ?"
			enable_event_replication = true

			replications {
				url = "http://localhost:8080"
				username = "%s"
				password = "%s"
			}
		}
	`
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      invalidCron,
				ExpectError: regexp.MustCompile(`.*syntax error in day-of-month field.*`),
			},
		},
	})
}

func TestInvalidReplicationUrlFails(t *testing.T) {
	const invalidUrl = `
		resource "artifactory_local_repository" "lib-local" {
			key = "lib-local"
			package_type = "maven"
		}

		resource "artifactory_replication_config" "lib-local" {
			repo_key = "${artifactory_local_repository.lib-local.key}"
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replications {
				url = "not a URL"
				username = "%s"
			}
		}
	`
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      invalidUrl,
				ExpectError: regexp.MustCompile(`.*expected "replications.0.url" to have a host, got not a URL.*`),
			},
		},
	})
}

func TestAccReplication_full(t *testing.T) {
	const replicationConfigTemplate = `
		resource "artifactory_local_repository" "lib-local" {
			key = "lib-local"
			package_type = "maven"
		}

		resource "artifactory_replication_config" "lib-local" {
			repo_key = "${artifactory_local_repository.lib-local.key}"
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replications {
				url = "%s"
				username = "%s"
			}
		}
	`

	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccCheckReplicationDestroy("artifactory_replication_config.lib-local"),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					replicationConfigTemplate,
					os.Getenv("ARTIFACTORY_URL"),
					os.Getenv("ARTIFACTORY_USERNAME"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "repo_key", "lib-local"),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "enable_event_replication", "true"),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "replications.#", "1"),
				),
			},
		},
	})
}

func testAccCheckReplicationDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}
		provider, _ := testAccProviders["artifactory"]()
		exists, _ := repConfigExists(rs.Primary.ID, provider.Meta())
		if exists {
			return fmt.Errorf("error: Replication %s still exists", id)
		}
		return nil
	}
}
