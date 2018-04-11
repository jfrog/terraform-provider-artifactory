package artifactory

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"net/http"
	"testing"
)

const permission_basic = `
resource "artifactory_permission_targets" "terraform-test-permission-basic" {
	name 	     = "testpermission"
	repositories = ["not-restricted"]
	users = [
		{
			name = "test_user"
			permissions = [
				"r",
				"w"
			]
		}
    ]
}`

func TestAccPermission_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testPermissionTargetCheckDestroy("artifactory_permission_targets.terraform-test-permission-basic"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: permission_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_permission_targets.terraform-test-permission-basic", "name", "testpermission"),
					resource.TestCheckResourceAttr("artifactory_permission_targets.terraform-test-permission-basic", "repositories.#", "1"),
				),
			},
		},
	})
}

func testPermissionTargetCheckDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*artifactory.Client)
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		permissionTargets, resp, err := client.Security.GetPermissionTargets(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Permission targets %s still exists %s", rs.Primary.ID, permissionTargets)
		}
	}
}
