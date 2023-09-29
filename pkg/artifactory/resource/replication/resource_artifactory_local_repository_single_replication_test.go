package replication_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccLocalSingleReplicationInvalidPushCron_fails(t *testing.T) {
	const invalidCron = `
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_local_repository_single_replication" "lib-local" {
			repo_key 							= "${artifactory_local_maven_repository.lib-local.key}"
			cron_exp 							= "0 0 xoxo xoxoxo * ?"
			enable_event_replication 			= true
			url 								= "http://localhost:8080"
 			username 							= "admin"
			password 							= "Passw0rd!"
			socket_timeout_millis 				= 16000
			enabled 							= true
			sync_deletes 						= true
			sync_properties 					= true
			sync_statistics 					= true
			include_path_prefix_pattern 		= "/some-repo/"
			exclude_path_prefix_pattern 		= "/some-other-repo/"
			check_binary_existence_in_filestore = true
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

func TestAccLocalSingleReplicationInvalidUrl_fails(t *testing.T) {
	const invalidUrl = `
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_local_repository_single_replication" "lib-local" {
			repo_key 							= "${artifactory_local_maven_repository.lib-local.key}"
			cron_exp 							= "0 0 xoxo xoxoxo * ?"
			enable_event_replication 			= true
			url 								= "not a URL"
 			username 							= "admin"
			password 							= "Passw0rd!"
			socket_timeout_millis 				= 16000
			enabled 							= true
			sync_deletes 						= true
			sync_properties 					= true
			sync_statistics 					= true
			include_path_prefix_pattern 		= "/some-repo/"
			exclude_path_prefix_pattern 		= "/some-other-repo/"
			check_binary_existence_in_filestore = true
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

func TestAccLocalSingleReplicationInvalidRclass_fails(t *testing.T) {
	const invalidUrl = `
		resource "artifactory_remote_maven_repository" "lib-remote" {
			key = "lib-remote"
			url = "https://repo1.maven.org/maven2/"
		}

		resource "artifactory_local_repository_single_replication" "lib-remote" {
			repo_key 							= "${artifactory_remote_maven_repository.lib-remote.key}"
			cron_exp 							= "0 0 * * * ?"
			enable_event_replication 			= true
			url 								= "http://localhost:8080"
 			username 							= "admin"
			password 							= "Passw0rd!"
			socket_timeout_millis 				= 16000
			enabled 							= true
			sync_deletes 						= true
			sync_properties 					= true
			sync_statistics 					= true
			include_path_prefix_pattern 		= "/some-repo/"
			exclude_path_prefix_pattern 		= "/some-other-repo/"
			check_binary_existence_in_filestore = true
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidUrl,
				ExpectError: regexp.MustCompile(`.*source repository rclass is not local.*`),
			},
		},
	})
}

func TestAccLocalSingleReplication_full(t *testing.T) {
	const testProxy = "test-proxy"
	_, fqrn, name := testutil.MkNames("lib-local", "artifactory_local_repository_single_replication")
	params := map[string]string{
		"url":       acctest.GetArtifactoryUrl(t),
		"username":  acctest.RtDefaultUser,
		"proxy":     testProxy,
		"repo_name": name,
	}
	replicationConfig := utilsdk.ExecuteTemplate("TestAccPushReplication", `
		resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_local_repository_single_replication" "{{ .repo_name }}" {
			repo_key 							= "${artifactory_local_maven_repository.{{ .repo_name }}.key}"
			cron_exp 							= "0 0 * * * ?"
			enable_event_replication 			= true
			url 								= "{{ .url }}"
 			username 							= "{{ .username }}"
			password 							= "Passw0rd!"
			proxy 								= "{{ .proxy }}"
			socket_timeout_millis 				= 16000
			enabled 							= true
			sync_deletes 						= true
			sync_properties 					= true
			sync_statistics 					= true
			include_path_prefix_pattern 		= "/some-repo/"
			exclude_path_prefix_pattern 		= "/some-other-repo/"
			check_binary_existence_in_filestore = true
		}
	`, params)

	replicationUpdateConfig := utilsdk.ExecuteTemplate("TestAccPushReplication", `
		resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_local_repository_single_replication" "{{ .repo_name }}" {
			repo_key 							= "${artifactory_local_maven_repository.{{ .repo_name }}.key}"
			cron_exp 							= "0 0 0 * * ?"
			enable_event_replication 			= false
			url 								= "{{ .url }}"
 			username 							= "{{ .username }}"
			password 							= "Passw0rd!"
			proxy 								= "{{ .proxy }}"
			socket_timeout_millis 				= 17000
			enabled 							= false
			sync_deletes 						= false
			sync_properties 					= false
			sync_statistics 					= false
			include_path_prefix_pattern 		= "/some-repo-modified/"
			exclude_path_prefix_pattern 		= "/some-other-repo-modified/"
			check_binary_existence_in_filestore = false
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
					resource.TestCheckResourceAttr(fqrn, "url", params["url"]),
					resource.TestCheckResourceAttr(fqrn, "username", params["username"]),
					resource.TestCheckResourceAttr(fqrn, "proxy", params["proxy"]),
					resource.TestCheckResourceAttr(fqrn, "socket_timeout_millis", "16000"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_deletes", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_properties", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_statistics", "true"),
					resource.TestCheckResourceAttr(fqrn, "include_path_prefix_pattern", "/some-repo/"),
					resource.TestCheckResourceAttr(fqrn, "exclude_path_prefix_pattern", "/some-other-repo/"),
					resource.TestCheckResourceAttr(fqrn, "check_binary_existence_in_filestore", "true"),
				),
			},
			{
				Config: replicationUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 0 * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "false"),
					resource.TestCheckResourceAttr(fqrn, "url", params["url"]),
					resource.TestCheckResourceAttr(fqrn, "username", params["username"]),
					resource.TestCheckResourceAttr(fqrn, "proxy", params["proxy"]),
					resource.TestCheckResourceAttr(fqrn, "socket_timeout_millis", "17000"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_deletes", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_properties", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_statistics", "false"),
					resource.TestCheckResourceAttr(fqrn, "include_path_prefix_pattern", "/some-repo-modified/"),
					resource.TestCheckResourceAttr(fqrn, "exclude_path_prefix_pattern", "/some-other-repo-modified/"),
					resource.TestCheckResourceAttr(fqrn, "check_binary_existence_in_filestore", "false"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "repo_key"),
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}
