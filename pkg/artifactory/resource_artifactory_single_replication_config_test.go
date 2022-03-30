package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		rtDefaultUser,
		proxy,
	)
}
func TestInvalidCronSingleReplication(t *testing.T) {

	_, fqrn, name := mkNames("lib-local", "artifactory_single_replication_config")
	var failCron = mkTclForRepConfg(name, "0 0 * * * !!", os.Getenv("ARTIFACTORY_URL"), "")

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

func TestInvalidUrlSingleReplication(t *testing.T) {

	_, fqrn, name := mkNames("lib-local", "artifactory_single_replication_config")
	var failCron = mkTclForRepConfg(name, "0 0 * * * ?", "bad_url", "")

	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccCheckReplicationDestroy(fqrn),
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
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
	_, fqrn, name := mkNames("lib-local", "artifactory_single_replication_config")
	config := mkTclForRepConfg(name, "0 0 * * * ?", os.Getenv("ARTIFACTORY_URL"), testProxy)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			createProxy(t, testProxy)
		},
		CheckDestroy: func() func(*terraform.State) error {
			deleteProxy(t, testProxy)
			return testAccCheckReplicationDestroy(fqrn)
		}(),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "url", os.Getenv("ARTIFACTORY_URL")),
					resource.TestCheckResourceAttr(fqrn, "username", rtDefaultUser),
					resource.TestCheckResourceAttr(fqrn, "proxy", testProxy),
				),
			},
		},
	})
}

func TestAccSingleReplication_withDelRepo(t *testing.T) {
	_, fqrn, name := mkNames("lib-local", "artifactory_single_replication_config")
	config := mkTclForRepConfg(name, "0 0 * * * ?", os.Getenv("ARTIFACTORY_URL"), "")
	var deleteRepo = func() {
		restyClient := getTestResty(t)
		_, err := restyClient.R().Delete("artifactory/api/repositories/" + name)
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("Delete repo %s done.", name)
	}
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
					resource.TestCheckResourceAttr(fqrn, "url", os.Getenv("ARTIFACTORY_URL")),
					resource.TestCheckResourceAttr(fqrn, "username", rtDefaultUser),
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
	_, fqrn, name := mkNames("lib-remote", "artifactory_single_replication_config")
	_, fqrepoName, repo_name := mkNames("lib-remote", "artifactory_remote_maven_repository")
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
	tcl = executeTemplate("foo", tcl, map[string]string{
		"repoconfig_name": name,
		"remote_name":     repo_name,
		"username":        os.Getenv("ARTIFACTORY_USERNAME"),
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
