package artifactory

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)


func TestAccManagedUser_NoGroups(t *testing.T) {
	const userNoGroups = `
		resource "artifactory_managed_user" "%s" {
			name        		= "dummy_user%d"
			email       		= "dummy%d@a.com"
			password			= "Passsw0rd!"
		}
	`
	id, FQRN, name := utils.MkNames("foobar-", "artifactory_managed_user")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckManagedUserDestroy(FQRN),
		ProviderFactories: utils.TestAccProviders(Provider()),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userNoGroups, name, id, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(FQRN, "name", fmt.Sprintf("dummy_user%d", id)),
					resource.TestCheckResourceAttr(FQRN, "groups.#", "0"),
				),
			},
		},
	})
}

func TestAccManagedUser_EmptyGroups(t *testing.T) {
	const userEmptyGroups = `
		resource "artifactory_managed_user" "%s" {
			name        		= "dummy_user%d"
			email       		= "dummy%d@a.com"
			password			= "Passsw0rd!"
			groups      		= []
		}
	`
	id, FQRN, name := utils.MkNames("foobar-", "artifactory_managed_user")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckManagedUserDestroy(FQRN),
		ProviderFactories: utils.TestAccProviders(Provider()),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userEmptyGroups, name, id, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(FQRN, "name", fmt.Sprintf("dummy_user%d", id)),
					resource.TestCheckResourceAttr(FQRN, "groups.#", "0"),
				),
			},
		},
	})
}

func TestAccManagedUser(t *testing.T) {
	const userFull = `
		resource "artifactory_managed_user" "%s" {
			name        		= "dummy_user%d"
			email       		= "dummy%d@a.com"
			password			= "Passsw0rd!"
			admin    			= true
			profile_updatable   = true
			disable_ui_access	= false
			groups      		= [ "readers" ]
		}
	`
	const userNonAdminNoProfUpd = `
		resource "artifactory_managed_user" "%s" {
			name        		= "dummy_user%d"
			email       		= "dummy%d@a.com"
			password			= "Passsw0rd!"
			admin    			= false
			profile_updatable   = false
			groups      		= [ "readers" ]
		}
	`
	id, FQRN, name := utils.MkNames("foobar-", "artifactory_managed_user")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckManagedUserDestroy(FQRN),
		ProviderFactories: utils.TestAccProviders(Provider()),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userFull, name, id, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(FQRN, "name", fmt.Sprintf("dummy_user%d", id)),
					resource.TestCheckResourceAttr(FQRN, "email", fmt.Sprintf("dummy%d@a.com", id)),
					resource.TestCheckResourceAttr(FQRN, "admin", "true"),
					resource.TestCheckResourceAttr(FQRN, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(FQRN, "disable_ui_access", "false"),
					resource.TestCheckResourceAttr(FQRN, "groups.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(userNonAdminNoProfUpd, name, id, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(FQRN, "name", fmt.Sprintf("dummy_user%d", id)),
					resource.TestCheckResourceAttr(FQRN, "email", fmt.Sprintf("dummy%d@a.com", id)),
					resource.TestCheckResourceAttr(FQRN, "admin", "false"),
					resource.TestCheckResourceAttr(FQRN, "profile_updatable", "false"),
				),
			},
			{
				ResourceName:            FQRN,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func testAccCheckManagedUserDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := utils.TestAccProviders(Provider())["artifactory"]()
		utils.ConfigureProvider(provider)
		client := provider.Meta().(*resty.Client)

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
