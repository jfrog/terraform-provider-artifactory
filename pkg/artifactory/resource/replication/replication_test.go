package replication_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/replication"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func repConfigExists(id string, m interface{}) (bool, error) {
	_, err := m.(utilsdk.ProvderMetadata).Client.R().Head(replication.EndpointPath + id)
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
