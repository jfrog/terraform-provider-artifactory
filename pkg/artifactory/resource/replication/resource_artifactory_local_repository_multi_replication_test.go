package replication_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalMultiReplication_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("lib-local", "artifactory_local_repository_multi_replication")

	params := map[string]interface{}{
		"url":       acctest.GetArtifactoryUrl(t),
		"username":  acctest.RtDefaultUser,
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccLocalMultiReplication_UpgradeFromSDKv2", `
		resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_local_repository_multi_replication" "{{ .repo_name }}" {
			repo_key = artifactory_local_maven_repository.{{ .repo_name }}.key
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replication {
				url 								= "{{ .url }}"
				username 							= "{{ .username }}"
				socket_timeout_millis 				= 16000
				enabled 							= true
				sync_deletes 						= true
				sync_properties 					= true
				sync_statistics 					= true
			}
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
					resource.TestCheckResourceAttr(fqrn, "replication.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "replication.0.url", acctest.GetArtifactoryUrl(t)),
					resource.TestCheckResourceAttr(fqrn, "replication.0.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "replication.0.check_binary_existence_in_filestore", "false"),
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

func TestAccLocalMultiReplication_InvalidPushCronFails(t *testing.T) {
	_, _, name := testutil.MkNames("lib-local", "artifactory_local_repository_multi_replication")
	params := map[string]interface{}{
		"repo_name": name,
	}
	invalidCron := util.ExecuteTemplate(
		"TestAccLocalMultiReplicationInvalidPushCronFails",
		`resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_local_repository_multi_replication" "{{ .repo_name }}" {
			repo_key = "${artifactory_local_maven_repository.{{ .repo_name }}.key}"
			cron_exp = "0 0 blah foo boo ?"
			enable_event_replication = true

			replication {
				url = "http://localhost:8080"
				username = "test-user"
				password = "test-password"
			}
		}`,
		params)

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

func TestAccLocalMultiReplication_InvalidUrlFails(t *testing.T) {
	_, _, name := testutil.MkNames("lib-local", "artifactory_local_repository_multi_replication")
	params := map[string]interface{}{
		"repo_name": name,
	}
	invalidUrl := util.ExecuteTemplate(
		"TestAccLocalMultiReplicationInvalidUrlFails",
		`resource "artifactory_local_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}

		resource "artifactory_local_repository_multi_replication" "{{ .repo_name }}" {
			repo_key = "${artifactory_local_maven_repository.{{ .repo_name }}.key}"
			cron_exp = "0 0 blah foo boo ?"
			enable_event_replication = true

			replication {
				url = "not a URL"
				username = "test-user"
				password = "test-password"
			}
		}`,
		params,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidUrl,
				ExpectError: regexp.MustCompile(`.*value must be a valid URL with host and http or\n.*https scheme.*`),
			},
		},
	})
}

func TestAccLocalMultiReplication_InvalidRclass_fails(t *testing.T) {
	const testProxy = "test-proxy"
	_, _, name := testutil.MkNames("lib-local", "artifactory_local_repository_multi_replication")
	params := map[string]interface{}{
		"url":       acctest.GetArtifactoryUrl(t),
		"username":  acctest.RtDefaultUser,
		"proxy":     testProxy,
		"repo_name": name,
	}
	replicationConfig := util.ExecuteTemplate("TestAccPushReplication", `
		resource "artifactory_remote_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo1.maven.org/maven2/"
		}

		resource "artifactory_local_repository_multi_replication" "{{ .repo_name }}" {
			repo_key = "${artifactory_remote_maven_repository.{{ .repo_name }}.key}"
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replication {
				url = "{{ .url }}"
				username = "{{ .username }}"
				password = "Passw0rd!"
				proxy = "{{ .proxy }}"
			}
		}
	`, params)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      replicationConfig,
				ExpectError: regexp.MustCompile(`.*source repository rclass is not local.*`),
			},
		},
	})
}

func TestAccLocalMultiReplication_full(t *testing.T) {
	const testProxy = "test-proxy"
	_, fqrn, name := testutil.MkNames("lib-local", "artifactory_local_repository_multi_replication")

	params := map[string]interface{}{
		"url":       acctest.GetArtifactoryUrl(t),
		"username":  acctest.RtDefaultUser,
		"proxy":     testProxy,
		"repo_name": name,
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

		resource "artifactory_local_repository_multi_replication" "{{ .repo_name }}" {
			repo_key = artifactory_local_maven_repository.{{ .repo_name }}.key
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replication {
				url 								= "{{ .url }}"
				username 							= "{{ .username }}"
				password 							= "Passw0rd!"
				proxy 								= artifactory_proxy.{{ .proxy }}.key
				socket_timeout_millis 				= 16000
				enabled 							= true
				sync_deletes 						= true
				sync_properties 					= true
				sync_statistics 					= true
				include_path_prefix_pattern 		= "/some-repo/"
				exclude_path_prefix_pattern 		= "/some-other-repo/"
			}
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

		resource "artifactory_local_repository_multi_replication" "{{ .repo_name }}" {
			repo_key = artifactory_local_maven_repository.{{ .repo_name }}.key
			cron_exp = "0 0 * * * ?"
			enable_event_replication = true

			replication {
				url 								= "{{ .url }}"
				username 							= "{{ .username }}"
				password 							= "Passw0rd!"
				proxy 								= artifactory_proxy.{{ .proxy }}.key
				enabled 							= false
				socket_timeout_millis 				= 16000
				sync_deletes 						= true
				sync_properties 					= true
				sync_statistics 					= true
				include_path_prefix_pattern 		= "/some-repo/"
				exclude_path_prefix_pattern 		= "/some-other-repo/"
				check_binary_existence_in_filestore = true

			}
			replication {
				url 								= "https://dummyurl.com/"
				username 							= "{{ .username }}"
				password 							= "Passw0rd!"
				proxy 								= artifactory_proxy.{{ .proxy }}.key
				enabled 							= false
				socket_timeout_millis 				= 16000
				sync_deletes 						= true
				sync_properties 					= true
				sync_statistics 					= true
				include_path_prefix_pattern 		= "/some-repo/"
				exclude_path_prefix_pattern 		= "/some-other-repo/"
				check_binary_existence_in_filestore = true
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccCheckPushReplicationDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "replication.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "replication.0.url", acctest.GetArtifactoryUrl(t)),
					resource.TestCheckResourceAttr(fqrn, "replication.0.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "replication.0.password", "Passw0rd!"),
					resource.TestCheckResourceAttr(fqrn, "replication.0.proxy", testProxy),
					resource.TestCheckResourceAttr(fqrn, "replication.0.check_binary_existence_in_filestore", "false"),
				),
			},
			{
				Config: updateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "replication.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "replication.0.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "replication.0.password", "Passw0rd!"),
					resource.TestCheckResourceAttr(fqrn, "replication.0.proxy", testProxy),
					resource.TestCheckResourceAttr(fqrn, "replication.0.enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "replication.0.check_binary_existence_in_filestore", "true"),
					resource.TestCheckResourceAttr(fqrn, "replication.1.username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "replication.1.password", "Passw0rd!"),
					resource.TestCheckResourceAttr(fqrn, "replication.1.proxy", testProxy),
					resource.TestCheckResourceAttr(fqrn, "replication.1.enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "replication.1.check_binary_existence_in_filestore", "true"),
					resource.TestCheckTypeSetElemAttr(fqrn, "replication.*.*", acctest.GetArtifactoryUrl(t)),
					resource.TestCheckTypeSetElemAttr(fqrn, "replication.*.*", "https://dummyurl.com/"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication.0.password", "replication.1.password"}, // this attribute is not being sent via API, can't be imported
			},
		},
	})
}
