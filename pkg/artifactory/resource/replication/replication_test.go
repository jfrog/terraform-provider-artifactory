package replication_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/util"

	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/replication"
)

func repConfigExists(id string, m interface{}) (bool, error) {
	_, err := m.(util.ProvderMetadata).Client.R().Head(replication.EndpointPath + id)
	return err == nil, err
}

func testAccCheckPushReplicationDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		exists, _ := repConfigExists(rs.Primary.ID, acctest.Provider.Meta())
		if exists {
			return fmt.Errorf("error: Replication %s still exists", id)
		}
		return nil
	}
}
