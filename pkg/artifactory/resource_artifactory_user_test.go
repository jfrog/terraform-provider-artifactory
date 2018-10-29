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

const user_basic = `
resource "artifactory_user" "basic" {
	name   = "tf-user-test"
    email  = "tf-user-test@domain.com"
	groups = [ "logged-in-user", "readers" ]
}`

func TestAccUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckUserDestroy("artifactory_user.basic"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: user_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_user.basic", "name", "tf-user-test"),
					resource.TestCheckResourceAttr("artifactory_user.basic", "email", "tf-user-test@domain.com"),
					resource.TestCheckResourceAttr("artifactory_user.basic", "admin", "false"),
					resource.TestCheckResourceAttr("artifactory_user.basic", "profile_updatable", "true"),
				),
			},
		},
	})
}

const user_full = `
resource "artifactory_user" "full" {
	name        		= "tf-user-test"
    email       		= "terraform_test@a.com"
    admin    			= true
    profile_updatable   = true
    groups              = [ "logged-in-user", "readers" ]
}`

func TestAccUser_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckUserDestroy("artifactory_user.full"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: user_full,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_user.full", "name", "tf-user-test"),
					resource.TestCheckResourceAttr("artifactory_user.full", "email", "terraform_test@a.com"),
					resource.TestCheckResourceAttr("artifactory_user.full", "admin", "true"),
					resource.TestCheckResourceAttr("artifactory_user.full", "profile_updatable", "true"),
					resource.TestCheckResourceAttr("artifactory_user.full", "groups.#", "2"),
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

		// It seems artifactory just can't keep up with high requests
		time.Sleep(time.Duration(1 * time.Second))
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
