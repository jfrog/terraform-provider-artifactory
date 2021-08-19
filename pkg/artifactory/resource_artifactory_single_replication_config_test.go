package artifactory

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const singleReplicationConfigTemplate = `
resource "artifactory_local_repository" "lib-local" {
	key = "lib-local"
	package_type = "maven"
}

resource "artifactory_single_replication_config" "lib-local" {
	repo_key = "${artifactory_local_repository.lib-local.key}"
	cron_exp = "0 0 * * * ?"
	enable_event_replication = true
	url = "%s"
	username = "%s"
	password = "%s"
}
`

func TestAccSingleReplication_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		CheckDestroy: testAccCheckSingleReplicationDestroy("artifactory_single_replication_config.lib-local"),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					singleReplicationConfigTemplate,
					os.Getenv("ARTIFACTORY_URL"),
					os.Getenv("ARTIFACTORY_USERNAME"),
					os.Getenv("ARTIFACTORY_PASSWORD"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_single_replication_config.lib-local", "repo_key", "lib-local"),
					resource.TestCheckResourceAttr("artifactory_single_replication_config.lib-local", "cron_exp", "0 0 * * * ?"),
					resource.TestCheckResourceAttr("artifactory_single_replication_config.lib-local", "enable_event_replication", "true"),
					resource.TestCheckResourceAttr("artifactory_single_replication_config.lib-local", "url", os.Getenv("ARTIFACTORY_URL")),
					resource.TestCheckResourceAttr("artifactory_single_replication_config.lib-local", "username", os.Getenv("ARTIFACTORY_USERNAME")),
					resource.TestCheckResourceAttr("artifactory_single_replication_config.lib-local", "password", os.Getenv("ARTIFACTORY_PASSWORD")),
				),
			},
		},
	})
}

func testAccCheckSingleReplicationDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.ArtOld

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		replica, resp, err := client.V1.Artifacts.GetRepositoryReplicationConfig(context.Background(), rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
			return nil
		}
		return fmt.Errorf("error: Replication %s still exists", replica)
	}
}
