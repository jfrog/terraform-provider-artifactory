package artifactory

import (
	"fmt"
	"os"
	"testing"
	"time"

	"context"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"net/http"
)

const replicationConfigTemplate = `
resource "artifactory_local_repository" "rep-src" {
	key = "rep-src"
	package_type = "maven"
}

resource "artifactory_local_repository" "rep-dest" {
	key = "rep-dest"
	package_type = "maven"
}

resource "artifactory_replication_config" "foo-rep" {
	repo_key = "${artifactory_local_repository.rep-src.key}"
	cron_exp = "0 0 * * * ?"
	enable_event_replication = true
	
	replications = [
		{
			url = "%s"
			username = "%s"
			password = "%s"
		}
	]
}
`

func TestAccReplication_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() {},
		CheckDestroy: testAccCheckReplicationDestroy("artifactory_replication_config.foo-rep"),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(replicationConfigTemplate, os.Getenv("ARTIFACTORY_URL"), os.Getenv("ARTIFACTORY_USERNAME"), os.Getenv("ARTIFACTORY_PASSWORD")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_replication_config.foo-rep", "repo_key", "rep-src"),
					resource.TestCheckResourceAttr("artifactory_replication_config.foo-rep", "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr("artifactory_replication_config.foo-rep", "enable_event_replication", "true"),
					resource.TestCheckResourceAttr("artifactory_replication_config.foo-rep", "replications.#", "1"),
				),
			},
		},
	})
}

func testAccCheckReplicationDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*artifactory.Client)
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		// It seems artifactory just can't keep up with high requests
		time.Sleep(time.Duration(1 * time.Second))
		replica, resp, err := client.Artifacts.GetRepositoryReplicationConfig(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Replication %s still exists", replica)
		}
	}
}
