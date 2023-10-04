package replication_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
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

	_, _, name := testutil.MkNames("lib-local", "artifactory_pull_replication")
	var failCron = mkTclForPullRepConfg(name, "0 0 * * * !!", acctest.GetArtifactoryUrl(t))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      failCron,
				ExpectError: regexp.MustCompile(`.*Invalid cronExp.*`),
			},
		},
	})
}

func TestPullReplicationLocalRepoCron(t *testing.T) {
	cronExpressions := [...]string{
		"10/20 15 14 5-10 * ? *",
		"* 5,7,9 14-16 * * ? *",
		"* 5,7,9 14/2 ? * WED,Sat *",
		"* * * * * ? *",
		"* * 14/2 ? * mon/3 *",
		"* 5-9 14/2 ? * 1-3 *",
		"*/3 */51 */12 */2 */4 ? *",
		"* 5 22-23 ? * Sun *",
		"0/5 14,18,3-39,52 * ? JAN,MAR,SEP MON-FRI 2002-2010",
		"0 15 10 * * ? *",
		"0 15 10 ? * 6#2",
		"0 15 10 15 * ?",
		"0 0 2 ? * MON-SAT",
	}
	for _, cron := range cronExpressions {
		t.Run(cron, func(t *testing.T) {
			resource.Test(pullReplicationLocalRepoTestCase(cron, t))
		})
	}
}

func pullReplicationLocalRepoTestCase(cronExpression string, t *testing.T) (*testing.T, resource.TestCase) {
	_, fqrn, name := testutil.MkNames("lib-local", "artifactory_pull_replication")
	config := mkTclForPullRepConfg(name, cronExpression, acctest.GetArtifactoryUrl(t))

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationDestroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", cronExpression),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "password", "Passw0rd!"),
					resource.TestCheckResourceAttr(fqrn, "check_binary_existence_in_filestore", "true"),
				),
			},
		},
	}
}

func TestAccPullReplicationLocalRepo(t *testing.T) {
	_, fqrn, name := testutil.MkNames("lib-local", "artifactory_pull_replication")
	url := acctest.GetArtifactoryUrl(t)
	config := mkTclForPullRepConfg(name, "0 0 * * * ?", url)
	updatedConfig := mkTclForPullRepConfg(name, "1 0 * * * ?", url)

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
					resource.TestCheckResourceAttr(fqrn, "url", url),
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
					resource.TestCheckResourceAttr(fqrn, "url", url),
					resource.TestCheckResourceAttr(fqrn, "username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "password", "Passw0rd!"),
					resource.TestCheckResourceAttr(fqrn, "check_binary_existence_in_filestore", "true"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // this attribute is not being sent via API, can't be imported
			},
		},
	})
}

func TestAccPullReplicationRemoteRepo(t *testing.T) {
	_, fqrn, name := testutil.MkNames("lib-remote", "artifactory_pull_replication")
	_, fqrepoName, repoName := testutil.MkNames("lib-remote", "artifactory_remote_maven_repository")
	var tcl = `
		resource "artifactory_remote_maven_repository" "{{ .remote_name }}" {
			key 				  = "{{ .remote_name }}"
			url                   = "https://repo1.maven.org/maven2/"
			repo_layout_ref       = "maven-2-default"
		}

		resource "artifactory_pull_replication" "{{ .repoconfig_name }}" {
			repo_key 				 = "{{ .remote_name }}"
			cron_exp 				 = "0 0 12 ? * MON *"
			enable_event_replication = false
			depends_on 				 = [artifactory_remote_maven_repository.{{ .remote_name }}]
		}
	`
	tcl = utilsdk.ExecuteTemplate("foo", tcl, map[string]string{
		"repoconfig_name": name,
		"remote_name":     repoName,
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
					resource.TestCheckResourceAttr(fqrn, "repo_key", repoName),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 12 ? * MON *"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "false"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_deletes", "false"),
					resource.TestCheckResourceAttr(fqrn, "sync_properties", "false"),
					resource.TestCheckResourceAttr(fqrn, "check_binary_existence_in_filestore", "false"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // this attribute is not being sent via API, can't be imported
			},
		},
	})
}
