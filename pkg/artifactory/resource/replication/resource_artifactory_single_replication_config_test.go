package replication_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func mkTclForRepConfg(name, cron, url, proxy string) string {
	const tcl = `
		resource "artifactory_local_maven_repository" "%s" {
			key = "%s"
		}

		resource "artifactory_single_replication_config" "%s" {
			repo_key = "${artifactory_local_maven_repository.%s.key}"
			cron_exp = "%s"
			enable_event_replication = true
			url = "%s"
			username = "%s"
			proxy = "%s"
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
		proxy,
	)
}

func TestAccSingleReplicationInvalidCron(t *testing.T) {

	_, _, name := testutil.MkNames("lib-local", "artifactory_single_replication_config")
	var failCron = mkTclForRepConfg(name, "0 0 * * * !!", acctest.GetArtifactoryUrl(t), "")

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

func TestAccSingleReplicationInvalidUrl(t *testing.T) {

	_, fqrn, name := testutil.MkNames("lib-local", "artifactory_single_replication_config")
	var failCron = mkTclForRepConfg(name, "0 0 * * * ?", "bad_url", "")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config:      failCron,
				ExpectError: regexp.MustCompile(`.*expected "url" to have a host, got bad_url.*`),
			},
		},
	})
}

func TestAccSingleReplication_full(t *testing.T) {
	const testProxy = "testProxy"
	_, fqrn, name := testutil.MkNames("lib-local", "artifactory_single_replication_config")
	config := mkTclForRepConfg(name, "0 0 * * * ?", acctest.GetArtifactoryUrl(t), testProxy)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProxy(t, testProxy)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func() func(*terraform.State) error {
			acctest.DeleteProxy(t, testProxy)
			return testAccCheckReplicationDestroy(fqrn)
		}(),

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "url", acctest.GetArtifactoryUrl(t)),
					resource.TestCheckResourceAttr(fqrn, "username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "proxy", testProxy),
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

func TestAccSingleReplication_withDelRepo(t *testing.T) {
	_, fqrn, name := testutil.MkNames("lib-local", "artifactory_single_replication_config")
	config := mkTclForRepConfg(name, "0 0 * * * ?", acctest.GetArtifactoryUrl(t), "")
	var deleteRepo = func() {
		restyClient := acctest.GetTestResty(t)
		_, err := restyClient.R().Delete("artifactory/api/repositories/" + name)
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("Delete repo %s done.", name)
	}
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
					resource.TestCheckResourceAttr(fqrn, "url", acctest.GetArtifactoryUrl(t)),
					resource.TestCheckResourceAttr(fqrn, "username", acctest.RtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "proxy", ""),
				),
			},
			{
				PreConfig: deleteRepo,
				Config:    config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
				),
			},
		},
	})
}

func TestAccSingleReplicationRemoteRepo(t *testing.T) {
	_, fqrn, name := testutil.MkNames("lib-remote", "artifactory_single_replication_config")
	_, fqrepoName, repoName := testutil.MkNames("lib-remote", "artifactory_remote_maven_repository")
	var tcl = `
		resource "artifactory_remote_maven_repository" "{{ .remote_name }}" {
			key 				  = "{{ .remote_name }}"
			url                   = "https://repo1.maven.org/maven2/"
			repo_layout_ref       = "maven-2-default"
		}

		resource "artifactory_single_replication_config" "{{ .repoconfig_name }}" {
			repo_key = "{{ .remote_name }}"
			cron_exp = "0 0 12 ? * MON *"
			enable_event_replication = false
			url = "https://repo1.maven.org/maven2/"
			username = "{{ .username }}"
			depends_on = [artifactory_remote_maven_repository.{{ .remote_name }}]
		}
	`
	tcl = utilsdk.ExecuteTemplate("foo", tcl, map[string]string{
		"repoconfig_name": name,
		"remote_name":     repoName,
		"username":        acctest.RtDefaultUser,
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
				),
			},
		},
	})
}
