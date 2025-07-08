package replication_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalSingleReplication_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-local-single-replication", "artifactory_local_repository_single_replication")

	params := map[string]string{
		"url":       acctest.GetArtifactoryUrl(t),
		"username":  acctest.RtDefaultUser,
		"proxy":     "test-proxy",
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccLocalSingleReplication_UpgradeFromSDKv2", `
		resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_local_repository_single_replication" "{{ .repo_name }}" {
			repo_key 							= artifactory_local_maven_repository.{{ .repo_name }}.key
			cron_exp 							= "0 0 * * * ?"
			enable_event_replication 			= true
			url 								= "{{ .url }}"
 			username 							= "{{ .username }}"
			socket_timeout_millis 				= 16000
			enabled 							= true
			sync_deletes 						= true
			sync_properties 					= true
			sync_statistics 					= true
			check_binary_existence_in_filestore = true
		}
	`, params)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "11.7.0",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "url", params["url"]),
					resource.TestCheckResourceAttr(fqrn, "username", params["username"]),
					resource.TestCheckResourceAttr(fqrn, "socket_timeout_millis", "16000"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_deletes", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_properties", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_statistics", "true"),
					resource.TestCheckResourceAttr(fqrn, "check_binary_existence_in_filestore", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccLocalSingleReplicationInvalidPushCron_fails(t *testing.T) {
	const invalidCron = `
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_local_repository_single_replication" "lib-local" {
			repo_key 							= artifactory_local_maven_repository.lib-local.key
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
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
			repo_key 							= artifactory_local_maven_repository.lib-local.key
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidUrl,
				ExpectError: regexp.MustCompile(`.*must be a valid URL with host and http or https scheme.*`),
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
			repo_key 							= artifactory_remote_maven_repository.lib-remote.key
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
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidUrl,
				ExpectError: regexp.MustCompile(`.*source repository rclass is not local.*`),
			},
		},
	})
}

func TestAccLocalSingleReplication_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-local-single-replication", "artifactory_local_repository_single_replication")

	params := map[string]string{
		"url":          acctest.GetArtifactoryUrl(t),
		"username":     acctest.RtDefaultUser,
		"proxy":        "test-proxy",
		"disableProxy": "false",
		"repo_name":    name,
	}

	config := util.ExecuteTemplate("TestAccPushReplication", `
		resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_proxy" "{{ .proxy }}" {
			key  = "{{ .proxy }}"
			host = "https://fake-proxy.org"
			port = 8080
		}

		resource "artifactory_local_repository_single_replication" "{{ .repo_name }}" {
			repo_key 							= artifactory_local_maven_repository.{{ .repo_name }}.key
			cron_exp 							= "0 0 * * * ?"
			enable_event_replication 			= true
			url 								= "{{ .url }}"
 			username 							= "{{ .username }}"
			password 							= "Passw0rd!"
			proxy 								= artifactory_proxy.{{ .proxy }}.key
			disable_proxy 						= {{ .disableProxy }}
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

	updateConfig := util.ExecuteTemplate("TestAccPushReplication", `
		resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_proxy" "{{ .proxy }}" {
			key  = "{{ .proxy }}"
			host = "https://fake-proxy.org"
			port = 8080
		}

		resource "artifactory_local_repository_single_replication" "{{ .repo_name }}" {
			repo_key 							= artifactory_local_maven_repository.{{ .repo_name }}.key
			cron_exp 							= "0 0 0 * * ?"
			enable_event_replication 			= false
			url 								= "{{ .url }}"
 			username 							= "{{ .username }}"
			password 							= "Passw0rd!"
			proxy 								= artifactory_proxy.{{ .proxy }}.key
			socket_timeout_millis 				= 17000
			enabled 							= false
			sync_deletes 						= false
			sync_properties 					= false
			sync_statistics 					= false
			include_path_prefix_pattern 		= "/some-repo-modified/"
			exclude_path_prefix_pattern 		= "/some-other-repo-modified/"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: func() func(*terraform.State) error {
			return testAccCheckPushReplicationDestroy(fqrn)
		}(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "url", params["url"]),
					resource.TestCheckResourceAttr(fqrn, "username", params["username"]),
					resource.TestCheckResourceAttr(fqrn, "proxy", params["proxy"]),
					resource.TestCheckResourceAttr(fqrn, "disable_proxy", params["disableProxy"]),
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
				Config: updateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 0 * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "false"),
					resource.TestCheckResourceAttr(fqrn, "url", params["url"]),
					resource.TestCheckResourceAttr(fqrn, "username", params["username"]),
					resource.TestCheckResourceAttr(fqrn, "proxy", params["proxy"]),
					resource.TestCheckResourceAttr(fqrn, "disable_proxy", params["disableProxy"]),
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
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccLocalSingleReplicationDisableProxy(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-local-single-replication-disable-proxy", "artifactory_local_repository_single_replication")

	params := map[string]string{
		"url":       acctest.GetArtifactoryUrl(t),
		"username":  acctest.RtDefaultUser,
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccLocalSingleReplicationDisableProxy", `
		resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_local_repository_single_replication" "{{ .repo_name }}" {
			repo_key 							= artifactory_local_maven_repository.{{ .repo_name }}.key
			cron_exp 							= "0 0 * * * ?"
			enable_event_replication 			= true
			url 								= "{{ .url }}"
 			username 							= "{{ .username }}"
			password 							= "Passw0rd!"
			disable_proxy 						= true
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: func() func(*terraform.State) error {
			return testAccCheckPushReplicationDestroy(fqrn)
		}(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "proxy", ""),
					resource.TestCheckResourceAttr(fqrn, "disable_proxy", "true"),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "url", params["url"]),
					resource.TestCheckResourceAttr(fqrn, "username", params["username"]),
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
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "replication_key"},
			},
		},
	})
}
