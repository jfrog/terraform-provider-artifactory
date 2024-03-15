package replication_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/replication"
	"github.com/jfrog/terraform-provider-shared/util"
)

func repConfigExists(id string, m interface{}) (bool, error) {
	resp, err := m.(util.ProvderMetadata).Client.R().Head(replication.EndpointPath + id)
	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, fmt.Errorf("%s", resp.String())
	}

	return true, nil
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
