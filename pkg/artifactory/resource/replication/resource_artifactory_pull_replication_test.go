package replication_test

import (
	"fmt"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
)

func mkTclForPullRepConfg(name, cron, url string) string {
	const tcl = `
		resource "artifactory_local_maven_repository" "%s" {
			key = "%s"
		}

		resource "artifactory_pull_replication" "%s" {
			repo_key = "${artifactory_local_maven_repository.%s.key}"
			cron_exp = "%s"
			enable_event_replication = true
			url = "%s"
			username = "%s"
			password = "Passw0rd!"
			check_binary_existence_in_filestore = true
		}
	`
	return fmt.Sprintf(tcl,
		name,
		name,
		name,
		name,
		cron,
		url,
		acctest.RtDefaultUser,
	)
}

func TestAccPullReplicationInvalidCron(t *testing.T) {

	_, fqrn, name := test.MkNames("lib-local", "artifactory_pull_replication")
	var failCron = mkTclForPullRepConfg(name, "0 0 * * * !!", acctest.GetArtifactoryUrl(t))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config:      failCron,
				ExpectError: regexp.MustCompile(`.*syntax error in year field: '!!'.*`),
			},
		},
	})
}

func TestAccPullReplicationLocalRepo(t *testing.T) {
	_, fqrn, name := test.MkNames("lib-local", "artifactory_pull_replication")
	config := mkTclForPullRepConfg(name, "0 0 * * * ?", acctest.GetArtifactoryUrl(t))
	updatedConfig := mkTclForPullRepConfg(name, "1 0 * * * ?", acctest.GetArtifactoryUrl(t))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationDestroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "password", "Passw0rd!"),
					resource.TestCheckResourceAttr(fqrn, "check_binary_existence_in_filestore", "true"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "1 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "password", "Passw0rd!"),
					resource.TestCheckResourceAttr(fqrn, "check_binary_existence_in_filestore", "true"),
				),
			},
		},
	})
}

func TestAccPullReplicationRemoteRepo(t *testing.T) {
	_, fqrn, name := test.MkNames("lib-remote", "artifactory_pull_replication")
	_, fqrepoName, repo_name := test.MkNames("lib-remote", "artifactory_remote_maven_repository")
	var tcl = `
		resource "artifactory_remote_maven_repository" "{{ .remote_name }}" {
			key 				  = "{{ .remote_name }}"
			url                   = "https://repo1.maven.org/maven2/"
			repo_layout_ref       = "maven-2-default"
		}

		resource "artifactory_pull_replication" "{{ .repoconfig_name }}" {
			repo_key = "{{ .remote_name }}"
			cron_exp = "0 0 12 ? * MON *"
			enable_event_replication = false
			depends_on = [artifactory_remote_maven_repository.{{ .remote_name }}]
		}
	`
	tcl = util.ExecuteTemplate("foo", tcl, map[string]string{
		"repoconfig_name": name,
		"remote_name":     repo_name,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(fqrepoName, acctest.CheckRepo),
			testAccCheckReplicationDestroy(fqrn),
		),

		Steps: []resource.TestStep{
			{
				Config: tcl,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", repo_name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 12 ? * MON *"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "false"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_deletes", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_properties", "false"),
					resource.TestCheckResourceAttr(fqrn, "check_binary_existence_in_filestore", "false"),
				),
			},
		},
	})
}
