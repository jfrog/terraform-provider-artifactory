package user_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func TestAccUserPasswordNotChangeWhenOtherAttributesChangeGH340(t *testing.T) {
	id := utils.RandomInt()
	name := fmt.Sprintf("user-%d", id)
	fqrn := fmt.Sprintf("artifactory_user.%s", name)

	email := fmt.Sprintf("dummy%d@a.com", id)
	password := "Passw0rd!"

	params := map[string]interface{}{
		"name":     name,
		"email":    email,
		"password": password,
	}
	userInitial := acctest.ExecuteTemplate("TestUser", `
		resource "artifactory_user" "{{ .name }}" {
			name              = "{{ .name }}"
			email             = "{{ .email }}"
			password          = "{{ .password }}"
			groups            = [ "readers" ]
			disable_ui_access = false
		}
	`, params)
	userUpdated := acctest.ExecuteTemplate("TestUser", `
		resource "artifactory_user" "{{ .name }}" {
			name              = "{{ .name }}"
			email             = "{{ .email }}"
			password          = "{{ .password }}"
			groups            = [ "readers" ]
			disable_ui_access = true
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: userInitial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "email", email),
					resource.TestCheckResourceAttr(fqrn, "password", password),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
				),
			},
			{
				Config: userUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "email", email),
					resource.TestCheckResourceAttr(fqrn, "password", password),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "true"),
				),
			},
		},
	})
}

func TestAccUser_basic(t *testing.T) {
	const userBasic = `
		resource "artifactory_user" "%s" {
			name  	= "dummy_user%d"
			password = "Passw0rd!"
			email 	= "dummy_user%d@a.com"
			groups  = [ "readers" ]
		}
	`
	id := utils.RandomInt()
	name := fmt.Sprintf("foobar-%d", id)
	fqrn := fmt.Sprintf("artifactory_user.%s", name)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userBasic, name, id, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("dummy_user%d", id)),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy_user%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func TestAccUserShouldCreateWithoutPassword(t *testing.T) {
	const userBasic = `
		resource "artifactory_user" "%s" {
			name  	= "dummy_user%d"
			email 	= "dummy_user%d@a.com"
			groups  = [ "readers" ]
		}
	`
	id := utils.RandomInt()
	name := fmt.Sprintf("foobar-%d", id)
	fqrn := fmt.Sprintf("artifactory_user.%s", name)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userBasic, name, id, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("dummy_user%d", id)),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy_user%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func TestAccUser_full(t *testing.T) {
	const userFull = `
		resource "artifactory_user" "%s" {
			name        		= "dummy_user%d"
			email       		= "dummy%d@a.com"
			password			= "Passw0rd!"
			admin    			= true
			profile_updatable   = true
			disable_ui_access	= false
			groups      		= [ "readers" ]
		}
	`
	const userNonAdminNoProfUpd = `
		resource "artifactory_user" "%s" {
			name        		= "dummy_user%d"
			email       		= "dummy%d@a.com"
			password			= "Passw0rd!"
			admin    			= false
			profile_updatable   = false
			groups      		= [ "readers" ]
		}
	`
	id, FQRN, name := acctest.MkNames("foobar-", "artifactory_user")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy(FQRN),
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

func TestAccUser_NoGroups(t *testing.T) {
	const userNoGroups = `
		resource "artifactory_user" "%s" {
			name        		= "dummy_user%d"
			email       		= "dummy%d@a.com"
		}
	`
	id, FQRN, name := acctest.MkNames("foobar-", "artifactory_user")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy(FQRN),
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

func TestAccUser_EmptyGroups(t *testing.T) {
	const userEmptyGroups = `
		resource "artifactory_user" "%s" {
			name        		= "dummy_user%d"
			email       		= "dummy%d@a.com"
			groups      		= []
		}
	`
	id, FQRN, name := acctest.MkNames("foobar-", "artifactory_user")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy(FQRN),
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

func testAccCheckUserDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(*resty.Client)

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
