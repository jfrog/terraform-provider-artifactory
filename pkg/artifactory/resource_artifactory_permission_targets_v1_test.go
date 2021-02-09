package artifactory

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const permissionV1Basic = `
resource "artifactory_permission_target" "test-perm" {
	name 	     = "test-perm"
	repositories = ["example-repo-local"]
	users {
		name = "anonymous"
		permissions = [
			"r",
			"w"
		]
	}
}`

func TestAccPermissionV1_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testPermissionTargetV1CheckDestroy("artifactory_permission_target.test-perm"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: permissionV1Basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_permission_target.test-perm", "name", "test-perm"),
				),
			},
		},
	})
}

func testPermissionTargetV1CheckDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.ArtOld

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		permissionTargets, resp, err := client.V1.Security.GetPermissionTargets(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Permission targets %s still exists %s", rs.Primary.ID, permissionTargets)
		}
	}
}
