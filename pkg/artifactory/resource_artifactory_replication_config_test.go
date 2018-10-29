package artifactory

import (
	"os"
	"fmt"
	"testing"
	
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"net/http"
	"context"
)

const replicationConfigTemplate = `
resource "artifactory_local_repository" "provider_test_source" {
	key = "provider_test_source"
	package_type = "maven"
}

resource "artifactory_local_repository" "provider_test_dest" {
	key = "provider_test_dest"
	package_type = "maven"
}

resource "artifactory_replication_config" "foo-rep" {
	repo_key = "${artifactory_local_repository.provider_test_source.key}"
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
		PreCheck: func(){},
		CheckDestroy:testAccCheckReplicationDestroy("artifactory_replication_config.foo"),
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(replicationConfigTemplate, os.Getenv("ARTIFACTORY_URL"), os.Getenv("ARTIFACTORY_USERNAME"), os.Getenv("ARTIFACTORY_PASSWORD")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_replication_config.foo-rep", "repo_key", "provider_test_source"),
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

		replica, resp, err := client.Artifacts.GetRepositoryReplicationConfig(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Replication %s still exists", replica)
		}
	}
}

