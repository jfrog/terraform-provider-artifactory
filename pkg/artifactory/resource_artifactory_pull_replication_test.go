package artifactory

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
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
		rtDefaultUser,
	)
}

func TestAccPullReplicationInvalidCron(t *testing.T) {

	_, fqrn, name := utils.MkNames("lib-local", "artifactory_pull_replication")
	var failCron = mkTclForPullRepConfg(name, "0 0 * * * !!", os.Getenv("ARTIFACTORY_URL"))

	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccCheckReplicationDestroy(fqrn),
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      failCron,
				ExpectError: regexp.MustCompile(`.*syntax error in year field: '!!'.*`),
			},
		},
	})
}

func TestAccPullReplicationLocalRepo(t *testing.T) {
	_, fqrn, name := utils.MkNames("lib-local", "artifactory_pull_replication")
	config := mkTclForPullRepConfg(name, "0 0 * * * ?", os.Getenv("ARTIFACTORY_URL"))
	updatedConfig := mkTclForPullRepConfg(name, "1 0 * * * ?", os.Getenv("ARTIFACTORY_URL"))
	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccCheckReplicationDestroy(fqrn),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "username", rtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "password", "Passw0rd!"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "1 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "username", rtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "password", "Passw0rd!"),
				),
			},
		},
	})
}

func compositeCheckDestroy(funcs ...func(state *terraform.State) error) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		var errors []error
		for _, f := range funcs {
			err := f(state)
			if err != nil {
				errors = append(errors, err)
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("%q", errors)
		}
		return nil
	}
}

func TestAccPullReplicationRemoteRepo(t *testing.T) {
	_, fqrn, name := utils.MkNames("lib-remote", "artifactory_pull_replication")
	_, fqrepoName, repo_name := utils.MkNames("lib-remote", "artifactory_remote_maven_repository")
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
	tcl = utils.ExecuteTemplate("foo", tcl, map[string]string{
		"repoconfig_name": name,
		"remote_name":     repo_name,
	})
	resource.Test(t, resource.TestCase{
		CheckDestroy: compositeCheckDestroy(
			verifyDeleted(fqrepoName, testCheckRepo),
			testAccCheckReplicationDestroy(fqrn),
		),

		ProviderFactories: testAccProviders,

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
