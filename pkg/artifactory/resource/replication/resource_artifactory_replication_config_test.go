package replication_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
)

func TestInvalidCronFails(t *testing.T) {
	const invalidCron = `
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_replication_config" "lib-local" {
			repo_key = "${artifactory_local_maven_repository.lib-local.key}"
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
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
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_replication_config" "lib-local" {
			repo_key = "${artifactory_local_maven_repository.lib-local.key}"
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replications {
				url = "not a URL"
				username = "%s"
			}
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidUrl,
				ExpectError: regexp.MustCompile(`.*expected "replications.0.url" to have a host, got not a URL.*`),
			},
		},
	})
}

func TestAccReplication_full(t *testing.T) {
	const testProxy = "test-proxy"
	const replicationConfigTemplate = `
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_replication_config" "lib-local" {
			repo_key = "${artifactory_local_maven_repository.lib-local.key}"
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replications {
				url = "%s"
				username = "%s"
				proxy = "%s"
			}
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProxy(t, testProxy)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func() func(*terraform.State) error {
			acctest.DeleteProxy(t, testProxy)
			return testAccCheckReplicationDestroy("artifactory_replication_config.lib-local")
		}(),

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					replicationConfigTemplate,
					acctest.GetArtifactoryUrl(t),
					acctest.RtDefaultUser,
					testProxy,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "repo_key", "lib-local"),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "enable_event_replication", "true"),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "replications.#", "1"),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "replications.0.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "replications.0.proxy", testProxy),
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

		exists, _ := repConfigExists(rs.Primary.ID, acctest.Provider.Meta())
		if exists {
			return fmt.Errorf("error: Replication %s still exists", id)
		}
		return nil
	}
}
