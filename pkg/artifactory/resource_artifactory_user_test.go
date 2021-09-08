package artifactory

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUser_basic(t *testing.T) {
	const userBasic = `
		resource "artifactory_user" "%s" {
			name  	= "the.dude%d"
			password = "Password1"
			email 	= "the.dude%d@domain.com"
			groups  = [ "readers" ]
		}
	`
	id := randomInt()
	name := fmt.Sprintf("foobar-%d", id)
	fqrn := fmt.Sprintf("artifactory_user.%s", name)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckUserDestroy(fqrn),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userBasic, name, id, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("the.dude%d", id)),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("the.dude%d@domain.com", id)),
					resource.TestCheckNoResourceAttr(fqrn, "admin"),
					resource.TestCheckNoResourceAttr(fqrn, "profile_updatable"),
				),
			},
		},
	})
}

func TestAccUser_full(t *testing.T) {
	const userFull = `
		resource "artifactory_user" "%s" {
			name        		= "dummy_user%d"
			email       		= "dummy%d@a.com"
			password			= "Password1"
			admin    			= true
			profile_updatable   = true
			groups      		= [ "readers" ]
		}
	`
	id, FQRN, name := mkNames("foobar-", "artifactory_user")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckUserDestroy(FQRN),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userFull, name, id, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(FQRN, "name", fmt.Sprintf("dummy_user%d", id)),
					resource.TestCheckResourceAttr(FQRN, "email", fmt.Sprintf("dummy%d@a.com", id)),
					resource.TestCheckResourceAttr(FQRN, "admin", "true"),
					resource.TestCheckResourceAttr(FQRN, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(FQRN, "groups.#", "1"),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*resty.Client)

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}
		resp, err := client.R().Head("artifactory/api/security/users/" + rs.Primary.ID)

		if err != nil {
			if resp != nil && resp.StatusCode() == http.StatusNotFound {
				return nil
			}
			return err
		}

		return fmt.Errorf("error: User %s still exists", rs.Primary.ID)
	}
}
