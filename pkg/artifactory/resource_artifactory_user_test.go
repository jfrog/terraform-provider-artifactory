package artifactory

import (
	"fmt"
	"testing"

	"context"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"net/http"
)

const user_basic = `
resource "artifactory_user" "foobar" {
	name  = "the.dude"
    email = "the.dude@domain.com"
}`

func TestAccUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckUserDestroy("artifactory_user.foobar"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: user_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_user.foobar", "name", "the.dude"),
					resource.TestCheckResourceAttr("artifactory_user.foobar", "email", "the.dude@domain.com"),
					resource.TestCheckResourceAttr("artifactory_user.foobar", "admin", "false"),
					resource.TestCheckResourceAttr("artifactory_user.foobar", "profile_updatable", "true"),
				),
			},
		},
	})
}

const user_full = `
resource "artifactory_user" "foobar" {
	name        		= "dummy_user"
    email       		= "dummy@a.com"
    admin    			= true
    profile_updatable   = true
    groups      = [ "readers" ]
}`

func TestAccUser_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckUserDestroy("artifactory_user.foobar"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: user_full,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_user.foobar", "name", "dummy_user"),
					resource.TestCheckResourceAttr("artifactory_user.foobar", "email", "dummy@a.com"),
					resource.TestCheckResourceAttr("artifactory_user.foobar", "admin", "true"),
					resource.TestCheckResourceAttr("artifactory_user.foobar", "profile_updatable", "true"),
					resource.TestCheckResourceAttr("artifactory_user.foobar", "groups.#", "1"),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*artifactory.Client)
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		user, resp, err := client.Security.GetUser(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: User %s still exists %s", rs.Primary.ID, user)
		}
	}
}
