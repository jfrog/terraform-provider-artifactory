package replication_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
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
				ExpectError: regexp.MustCompile(`.*Invalid cronExp.*`),
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
	_, fqrn, name := testutil.MkNames("lib-local", "artifactory_push_replication")
	params := map[string]interface{}{
		"url":       acctest.GetArtifactoryUrl(t),
		"username":  acctest.RtDefaultUser,
		"proxy":     testProxy,
		"repo_name": name,
	}
	replicationConfig := utilsdk.ExecuteTemplate("TestAccPushReplication", `
		resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_push_replication" "{{ .repo_name }}" {
			repo_key = "${artifactory_local_maven_repository.{{ .repo_name }}.key}"
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

	replicationUpdateConfig := utilsdk.ExecuteTemplate("TestAccPushReplication", `
		resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_push_replication" "{{ .repo_name }}" {
			repo_key = "${artifactory_local_maven_repository.{{ .repo_name }}.key}"
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replications {
				url = "{{ .url }}"
				username = "{{ .username }}"
				password = "Passw0rd!"
				proxy = "{{ .proxy }}"
				enabled = true
			}
			replications {
				url = "https://dummyurl.com/"
				username = "{{ .username }}"
				password = "Passw0rd!"
				proxy = "{{ .proxy }}"
				enabled = false
				check_binary_existence_in_filestore = true
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
			return testAccCheckPushReplicationDestroy(fqrn)
		}(),

		Steps: []resource.TestStep{
			{
				Config: replicationConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "replications.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "replications.0.url", acctest.GetArtifactoryUrl(t)),
					resource.TestCheckResourceAttr(fqrn, "replications.0.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "replications.0.password", "Passw0rd!"),
					resource.TestCheckResourceAttr(fqrn, "replications.0.proxy", testProxy),
					resource.TestCheckResourceAttr(fqrn, "replications.0.check_binary_existence_in_filestore", "false"),
				),
			},
			{
				Config: replicationUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "replications.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "replications.0.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "replications.0.password", "Passw0rd!"),
					resource.TestCheckResourceAttr(fqrn, "replications.0.proxy", testProxy),
					resource.TestCheckResourceAttr(fqrn, "replications.0.enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "replications.0.check_binary_existence_in_filestore", "false"),
					resource.TestCheckResourceAttr(fqrn, "replications.1.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "replications.1.password", "Passw0rd!"),
					resource.TestCheckResourceAttr(fqrn, "replications.1.proxy", testProxy),
					resource.TestCheckResourceAttr(fqrn, "replications.1.enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "replications.1.check_binary_existence_in_filestore", "true"),
					resource.TestCheckTypeSetElemAttr(fqrn, "replications.*.*", acctest.GetArtifactoryUrl(t)),
					resource.TestCheckTypeSetElemAttr(fqrn, "replications.*.*", "https://dummyurl.com/"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replications.0.password", "replications.1.password"}, // this attribute is not being sent via API, can't be imported
			},
		},
	})
}
