package artifactory

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const userBasic = `
resource "artifactory_user" "foobar" {
	name  = "the.dude"
    email = "the.dude@domain.com"
	groups      = [ "readers" ]
}`

func TestAccUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckUserDestroy("artifactory_user.foobar"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: userBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_user.foobar", "name", "the.dude"),
					resource.TestCheckResourceAttr("artifactory_user.foobar", "email", "the.dude@domain.com"),
					resource.TestCheckNoResourceAttr("artifactory_user.foobar", "admin"),
					resource.TestCheckNoResourceAttr("artifactory_user.foobar", "profile_updatable"),
				),
			},
		},
	})
}

const userFull = `
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
				Config: userFull,
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
		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.ArtOld

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		user, resp, err := client.V1.Security.GetUser(context.Background(), rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("error: User %s still exists %s", rs.Primary.ID, user)
	}
}
