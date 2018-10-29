package artifactory

import (
	"fmt"
	"testing"
	"time"

	"context"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"net/http"
)

const group_basic = `
resource "artifactory_group" "basic" {
	name  = "tf-group-basic"
}`

func TestAccGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckGroupDestroy("artifactory_group.basic"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: group_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_group.basic", "name", "tf-group-basic"),
				),
			},
		},
	})
}

const group_full = `
resource "artifactory_group" "full" {
	name             = "tf-group-full"
    description 	 = "Test group"
	auto_join        = true
	admin_privileges = false
	realm            = "test"
	realm_attributes = "Some attribute"
}`

func TestAccGroup_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckGroupDestroy("artifactory_group.full"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: group_full,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_group.full", "name", "tf-group-full"),
					resource.TestCheckResourceAttr("artifactory_group.full", "auto_join", "true"),
					resource.TestCheckResourceAttr("artifactory_group.full", "admin_privileges", "false"),
					resource.TestCheckResourceAttr("artifactory_group.full", "realm", "test"),
					resource.TestCheckResourceAttr("artifactory_group.full", "realm_attributes", "Some attribute"),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*artifactory.Client)
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		// It seems artifactory just can't keep up with high requests
		time.Sleep(time.Duration(1 * time.Second))
		_, resp, err := client.Security.GetGroup(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Group %s still exists", rs.Primary.ID)
		}
	}
}
