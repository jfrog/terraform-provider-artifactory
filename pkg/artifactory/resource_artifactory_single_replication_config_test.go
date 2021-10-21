package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func mkTclForRepConfg(name, cron, url string) string {
	const tcl = `
		resource "artifactory_local_repository" "%s" {
			key = "%s"
			package_type = "maven"
		}
		
		resource "artifactory_single_replication_config" "%s" {
			repo_key = "${artifactory_local_repository.%s.key}"
			cron_exp = "%s" 
			enable_event_replication = true
			url = "%s"
			username = "%s"
			password = "%s"
		}
	`
	return fmt.Sprintf(tcl,
		name,
		name,
		name,
		name,
		cron,
		url,
		os.Getenv("ARTIFACTORY_USERNAME"),
		os.Getenv("ARTIFACTORY_PASSWORD"),
	)
}
func TestInvalidCronSingleReplication(t *testing.T) {

	_, fqrn, name := mkNames("lib-local", "artifactory_single_replication_config")
	var failCron = mkTclForRepConfg(name, "0 0 * * * !!", os.Getenv("ARTIFACTORY_URL"))

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
	var failCron = mkTclForRepConfg(name, "0 0 * * * ?", "bad_url")

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

// Test was temporarily removed from the test suite
//func TestAccSingleReplication_full(t *testing.T) {
//	_, fqrn, name := mkNames("lib-local", "artifactory_single_replication_config")
//	config := mkTclForRepConfg(name, "0 0 * * * ?", os.Getenv("ARTIFACTORY_URL"))
//	resource.Test(t, resource.TestCase{
//		CheckDestroy: testAccCheckReplicationDestroy(fqrn),
//		Providers:    testAccProviders,
//
//		Steps: []resource.TestStep{
//			{
//				Config: config,
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
//					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
//					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
//					resource.TestCheckResourceAttr(fqrn, "url", os.Getenv("ARTIFACTORY_URL")),
//					resource.TestCheckResourceAttr(fqrn, "username", os.Getenv("ARTIFACTORY_USERNAME")),
//					// artifactory is sending us back a scrambled password and because we can't compute it, we can't
//					// store it's state. I am going to leave this test broken specifically to draw attention to this
//					// because local state will never match remote state and TF will have issues
//					// we send: password
//					// we get back: JE2fNsEThvb1buiH7h7S2RDsGWSdp2EcuG9Pky5AFyRMwE4UzG
//					//resource.TestCheckResourceAttr(fqrn, "password", os.Getenv("ARTIFACTORY_PASSWORD")),
//					resource.TestCheckResourceAttr(fqrn, "password", "Known issue in RT"),
//				),
//			},
//		},
//	})
//}

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
func TestAccSingleReplicationRemoteRepo(t *testing.T) {
	_, fqrn, name := mkNames("lib-remote", "artifactory_single_replication_config")
	_, fqrepoName, repo_name := mkNames("lib-remote", "artifactory_remote_repository")
	var tcl = `
		resource "artifactory_remote_repository" "{{ .remote_name }}" {
			key 				  = "{{ .remote_name }}"
			package_type          = "maven"
			url                   = "https://repo1.maven.org/maven2/"
			repo_layout_ref       = "maven-2-default"
		}

		resource "artifactory_single_replication_config" "{{ .repoconfig_name }}" {
			repo_key = "{{ .remote_name }}"
			cron_exp = "0 0 12 ? * MON *" 
			enable_event_replication = false
			url = "https://repo1.maven.org/maven2/"
			username = "christianb"
			password = "password"
			depends_on = [artifactory_remote_repository.{{ .remote_name }}]
		}
	`
	tcl = executeTemplate("foo", tcl, map[string]string{
		"repoconfig_name": name,
		"remote_name":     repo_name,
	})
	resource.Test(t, resource.TestCase{
		CheckDestroy: compositeCheckDestroy(
			testAccCheckRepositoryDestroy(fqrepoName),
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
