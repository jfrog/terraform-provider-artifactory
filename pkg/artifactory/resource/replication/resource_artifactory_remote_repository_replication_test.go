package replication_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccRemoteRepositoryReplication_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-remote-repo-replication", "artifactory_remote_repository_replication")

	params := map[string]interface{}{
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccPushSingleRemoteReplication", `
		resource "artifactory_remote_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo1.maven.org/maven2/"
		}

		resource "artifactory_remote_repository_replication" "{{ .repo_name }}" {
			repo_key 							= artifactory_remote_maven_repository.{{ .repo_name }}.key
			cron_exp 							= "0 0 * * * ?"
			enable_event_replication 			= true
			enabled 							= true
			sync_deletes 						= false
			sync_properties 					= true
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
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_deletes", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_properties", "true"),
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

func TestAccRemoteRepositoryReplication_InvalidPushCron_fails(t *testing.T) {
	const invalidCron = `
		resource "artifactory_remote_maven_repository" "lib-remote" {
			key = "lib-remote"
			url = "https://repo1.maven.org/maven2/"
		}

		resource "artifactory_remote_repository_replication" "lib-remote" {
			repo_key 							= "${artifactory_remote_maven_repository.lib-remote.key}"
			cron_exp 							= "0 0 xoxo xoxoxo * ?"
			enable_event_replication 			= true
			enabled 							= true
			sync_deletes 						= false
			sync_properties 					= true
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

func TestAccRemoteRepositoryReplication_InvalidRclass_fails(t *testing.T) {
	const invalidRclass = `
		resource "artifactory_local_maven_repository" "lib-local" {
			key = "lib-local"
		}

		resource "artifactory_remote_repository_replication" "lib-local" {
			repo_key 							= "${artifactory_local_maven_repository.lib-local.key}"
			cron_exp 							= "0 0 * * * ?"
			enable_event_replication 			= true
			enabled 							= true
			sync_deletes 						= false
			sync_properties 					= true
			include_path_prefix_pattern 		= "/some-repo/"
			exclude_path_prefix_pattern 		= "/some-other-repo/"
			check_binary_existence_in_filestore = false
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidRclass,
				ExpectError: regexp.MustCompile(`.*source repository rclass is not.*`),
			},
		},
	})
}

func TestAccRemoteRepositoryReplication_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-remote-repo-replication", "artifactory_remote_repository_replication")

	params := map[string]interface{}{
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccPushSingleRemoteReplication", `
		resource "artifactory_remote_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo1.maven.org/maven2/"
		}

		resource "artifactory_remote_repository_replication" "{{ .repo_name }}" {
			repo_key 							= artifactory_remote_maven_repository.{{ .repo_name }}.key
			cron_exp 							= "0 0 * * * ?"
			enable_event_replication 			= true
			enabled 							= true
			sync_deletes 						= false
			sync_properties 					= true
			include_path_prefix_pattern 		= "/some-repo/"
			exclude_path_prefix_pattern 		= "/some-other-repo/"
			check_binary_existence_in_filestore = true
		}
	`, params)

	updateConfig := util.ExecuteTemplate("TestAccPushSingleRemoteReplication", `
		resource "artifactory_remote_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo1.maven.org/maven2/"
		}

		resource "artifactory_remote_repository_replication" "{{ .repo_name }}" {
			repo_key 							= artifactory_remote_maven_repository.{{ .repo_name }}.key
			cron_exp 							= "0 0 0 * * ?"
			enable_event_replication 			= false
			enabled 							= false
			sync_deletes 						= true
			sync_properties 					= false
			include_path_prefix_pattern 		= "/some-repo-modified/"
			exclude_path_prefix_pattern 		= "/some-other-repo-modified/"
			check_binary_existence_in_filestore = false
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
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_deletes", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_properties", "true"),
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
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_deletes", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_properties", "false"),
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
