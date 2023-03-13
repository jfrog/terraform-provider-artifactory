package replication_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccReplicationConfigInvalidCronFails(t *testing.T) {
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
			}
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidCron,
				ExpectError: regexp.MustCompile(`.*Invalid cronExp.*`),
			},
		},
	})
}

func TestAccReplicationConfigInvalidUrlFails(t *testing.T) {
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

func TestAccReplicationConfig_full(t *testing.T) {
	const testProxy = "test-proxy"
	_, fqrn, name := test.MkNames("lib-local", "artifactory_replication_config")
	params := map[string]interface{}{
		"url":       acctest.GetArtifactoryUrl(t),
		"username":  acctest.RtDefaultUser,
		"proxy":     testProxy,
		"repo_name": name,
	}
	replicationUpdateConfig := util.ExecuteTemplate("ReplicationConfigTemplate", `
		resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_replication_config" "{{ .repo_name }}" {
			repo_key = "${artifactory_local_maven_repository.{{ .repo_name }}.key}"
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replications {
				url = "{{ .url }}"
				username = "{{ .username }}"
				proxy = "{{ .proxy }}"
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
			return testAccCheckReplicationDestroy(fqrn)
		}(),

		Steps: []resource.TestStep{
			{
				Config: replicationUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "replications.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "replications.0.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "replications.0.proxy", testProxy),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replications.0.password"}, // this attribute is not being sent via API, can't be imported
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
