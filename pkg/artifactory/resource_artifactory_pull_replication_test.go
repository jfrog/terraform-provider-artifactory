package artifactory_test

import (
	"fmt"
	"os"
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

	_, fqrn, name := acctest.MkNames("lib-local", "artifactory_pull_replication")
	var failCron = mkTclForPullRepConfg(name, "0 0 * * * !!", os.Getenv("ARTIFACTORY_URL"))

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
	_, fqrn, name := acctest.MkNames("lib-local", "artifactory_pull_replication")
	config := mkTclForPullRepConfg(name, "0 0 * * * ?", os.Getenv("ARTIFACTORY_URL"))
	updatedConfig := mkTclForPullRepConfg(name, "1 0 * * * ?", os.Getenv("ARTIFACTORY_URL"))
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
				),
			},
		},
	})
}

func TestAccPullReplicationRemoteRepo(t *testing.T) {
	_, fqrn, name := acctest.MkNames("lib-remote", "artifactory_pull_replication")
	_, fqrepoName, repo_name := acctest.MkNames("lib-remote", "artifactory_remote_maven_repository")
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
	tcl = acctest.ExecuteTemplate("foo", tcl, map[string]string{
		"repoconfig_name": name,
		"remote_name":     repo_name,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(fqrepoName, acctest.TestCheckRepo),
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
				),
			},
		},
	})
}
