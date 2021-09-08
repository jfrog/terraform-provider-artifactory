package artifactory

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)


func mkTclForRepConfg(name, cron, url string) string{
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
	var failCron = mkTclForRepConfg(name, "0 0 * * * !!",os.Getenv("ARTIFACTORY_URL"))

	resource.Test(t, resource.TestCase{
		CheckDestroy: testAccCheckReplicationDestroy(fqrn),
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
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
	var failCron = mkTclForRepConfg(name, "0 0 * * * ?","bad_url")

	resource.Test(t, resource.TestCase{
		CheckDestroy: testAccCheckReplicationDestroy(fqrn),
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      failCron,
				ExpectError: regexp.MustCompile(`.*expected "url" to have a host, got bad_url.*`),
			},
		},
	})
}

func TestAccSingleReplication_full(t *testing.T) {
	_, fqrn, name := mkNames("lib-local", "artifactory_single_replication_config")
	config := mkTclForRepConfg(name,"0 0 * * * ?",os.Getenv("ARTIFACTORY_URL"))
	resource.Test(t, resource.TestCase{
		CheckDestroy: testAccCheckReplicationDestroy(fqrn),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", name),
					resource.TestCheckResourceAttr(fqrn, "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr(fqrn, "enable_event_replication", "true"),
					resource.TestCheckResourceAttr(fqrn, "url", os.Getenv("ARTIFACTORY_URL")),
					resource.TestCheckResourceAttr(fqrn, "username", os.Getenv("ARTIFACTORY_USERNAME")),
					// artifactory is sending us back a scrambled password and because we can't compute it, we can't
					// store it's state. I am going to leave this test broken specifically to draw attention to this
					// because local state will never match remote state and TF will have issues
					// we send: password
					// we get back: JE2fNsEThvb1buiH7h7S2RDsGWSdp2EcuG9Pky5AFyRMwE4UzG
					//resource.TestCheckResourceAttr(fqrn, "password", os.Getenv("ARTIFACTORY_PASSWORD")),
					resource.TestCheckResourceAttr(fqrn, "password", "Known issue in RT"),
				),
			},
		},
	})
}
