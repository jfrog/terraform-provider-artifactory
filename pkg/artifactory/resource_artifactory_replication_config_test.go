package artifactory

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"net/http"
	"os"
	"testing"
)

const replicationConfigTemplate = `
resource "artifactory_local_repository" "lib-local" {
	key = "lib-local"
	package_type = "maven"
}

resource "artifactory_replication_config" "lib-local" {
	repo_key = "${artifactory_local_repository.lib-local.key}"
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
		CheckDestroy: testAccCheckReplicationDestroy("artifactory_replication_config.lib-local"),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					replicationConfigTemplate,
					os.Getenv("ARTIFACTORY_URL"),
					os.Getenv("ARTIFACTORY_USERNAME"),
					os.Getenv("ARTIFACTORY_PASSWORD"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "repo_key", "lib-local"),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "enable_event_replication", "true"),
					resource.TestCheckResourceAttr("artifactory_replication_config.lib-local", "replications.#", "1"),
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
