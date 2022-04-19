package artifactory_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory"
)

func TestAccPushReplicationInvalidPushCronFails(t *testing.T) {
	const invalidCron = `
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_push_replication" "lib-local" {
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

func TestAccPushReplicationInvalidUrlFails(t *testing.T) {
	const invalidUrl = `
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_push_replication" "lib-local" {
			repo_key = "${artifactory_local_maven_repository.lib-local.key}"
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replications {
				url = "not a URL"
				username = "%s"
				password = "Passw0rd!"
			}
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidUrl,
				ExpectError: regexp.MustCompile(`.*expected "url" to have a host, got not a URL.*`),
			},
		},
	})
}

func TestAccPushReplication_full(t *testing.T) {
	const testProxy = "test-proxy"

	params := map[string]interface{}{
		"url":      os.Getenv("ARTIFACTORY_URL"),
		"username": acctest.RtDefaultUser,
		"proxy":    testProxy,
	}
	replicationConfig := acctest.ExecuteTemplate("TestAccPushReplication", `
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_push_replication" "lib-local" {
			repo_key = "${artifactory_local_maven_repository.lib-local.key}"
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replications {
				url = "{{ .url }}"
				username = "{{ .username }}"
				password = "Passw0rd!"
				proxy = "{{ .proxy }}"
			}
		}
	`, params)

	replicationUpdateConfig := acctest.ExecuteTemplate("TestAccPushReplication", `
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_push_replication" "lib-local" {
			repo_key = "${artifactory_local_maven_repository.lib-local.key}"
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replications {
				url = "{{ .url }}"
				username = "{{ .username }}"
				password = "Passw0rd!"
				proxy = "{{ .proxy }}"
				enabled = true
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProxy(t, testProxy)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func() func(*terraform.State) error {
			acctest.DeleteProxy(t, testProxy)
			return testAccCheckPushReplicationDestroy("artifactory_push_replication.lib-local")
		}(),

		Steps: []resource.TestStep{
			{
				Config: replicationConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "repo_key", "lib-local"),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "enable_event_replication", "true"),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "replications.#", "1"),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "replications.0.url", os.Getenv("ARTIFACTORY_URL")),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "replications.0.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "replications.0.password", "Passw0rd!"),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "replications.0.proxy", testProxy),
				),
			},
			{
				Config: replicationUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "repo_key", "lib-local"),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "enable_event_replication", "true"),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "replications.#", "1"),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "replications.0.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "replications.0.password", "Passw0rd!"),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "replications.0.proxy", testProxy),
					resource.TestCheckResourceAttr("artifactory_push_replication.lib-local", "replications.0.enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckPushReplicationDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		exists, _ := artifactory.RepConfigExists(rs.Primary.ID, acctest.Provider.Meta())
		if exists {
			return fmt.Errorf("error: Replication %s still exists", id)
		}
		return nil
	}
}
