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

func TestAccRemoteReplicationInvalidPushCron_fails(t *testing.T) {
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

func TestAccRemoteReplicationInvalidRclass_fails(t *testing.T) {
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidRclass,
				ExpectError: regexp.MustCompile(`.*source repository rclass is not.*`),
			},
		},
	})
}

func TestAccRemoteReplicationRepo_full(t *testing.T) {
	const testProxy = "test-proxy"
	_, fqrn, name := testutil.MkNames("lib-remote", "artifactory_remote_repository_replication")
	params := map[string]interface{}{
		"repo_name": name,
	}
	replicationConfig := utilsdk.ExecuteTemplate("TestAccPushSingleRemoteReplication", `
		resource "artifactory_remote_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo1.maven.org/maven2/"
		}

		resource "artifactory_remote_repository_replication" "{{ .repo_name }}" {
			repo_key 							= "${artifactory_remote_maven_repository.{{ .repo_name }}.key}"
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

	replicationUpdateConfig := utilsdk.ExecuteTemplate("TestAccPushSingleRemoteReplication", `
		resource "artifactory_remote_maven_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo1.maven.org/maven2/"
		}

		resource "artifactory_remote_repository_replication" "{{ .repo_name }}" {
			repo_key 							= "${artifactory_remote_maven_repository.{{ .repo_name }}.key}"
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
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_deletes", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_properties", "true"),
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
