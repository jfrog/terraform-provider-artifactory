package artifactory

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"net/http"
	"testing"
	"time"
)

const permission_full = `
resource "artifactory_permission_targets" "full" {
	name 	     = "tf-permission-full"
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

func TestAccPermission_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testPermissionTargetCheckDestroy("artifactory_permission_targets.full"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: permission_full,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_permission_targets.full", "name", "tf-permission-full"),
					resource.TestCheckResourceAttr("artifactory_permission_targets.full", "repositories.#", "1"),
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
			fmt.Printf("%v\n", s.RootModule().Resources)
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		// It seems artifactory just can't keep up with high requests
		time.Sleep(time.Duration(1 * time.Second))
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
