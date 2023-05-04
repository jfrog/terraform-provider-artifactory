package user_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/provider"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccUser_basic_groups(t *testing.T) {
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := fmt.Sprintf(username + "@test.com")

	params := map[string]interface{}{
		"name":     fmt.Sprintf("foobar-%d", id),
		"username": username,
		"email":    email,
	}
	userNoGroups := utilsdk.ExecuteTemplate("TestAccUserBasic", `
		resource "artifactory_user" "{{ .name }}" {
			name        		= "{{ .name }}"
			email 				= "{{ .email }}"
			password			= "Passsw0rd!"
			admin 				= false
			groups      		= [ "readers" ]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: userNoGroups,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("foobar-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "id"),
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func TestAccUser_no_password(t *testing.T) {
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := fmt.Sprintf(username + "@test.com")

	params := map[string]interface{}{
		"name":     fmt.Sprintf("foobar-%d", id),
		"username": username,
		"email":    email,
	}
	userNoGroups := utilsdk.ExecuteTemplate("TestAccUserBasic", `
		resource "artifactory_user" "{{ .name }}" {
			name        		= "{{ .name }}"
			email 				= "{{ .email }}"
			admin 				= false
			groups      		= [ "readers" ]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: userNoGroups,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("foobar-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "id"),
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func TestAccUser_default_group(t *testing.T) {
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := fmt.Sprintf(username + "@test.com")

	params := map[string]interface{}{
		"name":     fmt.Sprintf("foobar-%d", id),
		"username": username,
		"email":    email,
	}
	userNoGroups := utilsdk.ExecuteTemplate("TestAccUserBasic", `
		resource "artifactory_user" "{{ .name }}" {
			name        		= "{{ .name }}"
			email 				= "{{ .email }}"
			password			= "Passsw0rd!"
			admin 				= false
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: userNoGroups,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("foobar-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "groups.0", "readers"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "id"),
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func TestAccUser_empty_groups(t *testing.T) {
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := fmt.Sprintf(username + "@test.com")

	params := map[string]interface{}{
		"name":     fmt.Sprintf("foobar-%d", id),
		"username": username,
		"email":    email,
	}
	userNoGroups := utilsdk.ExecuteTemplate("TestAccUserBasic", `
		resource "artifactory_user" "{{ .name }}" {
			name        		= "{{ .name }}"
			email 				= "{{ .email }}"
			password			= "Passsw0rd!"
			admin 				= false
			groups      		= []
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: userNoGroups,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("foobar-%d", id)),
					//resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "id"),
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func TestAccUser_all_attributes(t *testing.T) {
	const userFull = `
		resource "artifactory_user" "%s" {
			name        				= "dummy_user%d"
			email       				= "dummy%d@a.com"
			password					= "Passw0rd!"
			admin    					= true
			profile_updatable   		= true
			disable_ui_access			= false
			internal_password_disabled 	= false
			groups      				= [ "readers" ]
		}
	`
	const userUpdated = `
		resource "artifactory_user" "%s" {
			name        				= "dummy_user%d"
			email       				= "dummy%d@a.com"
			password					= "Passw0rd!"
			admin    					= false
			profile_updatable   		= true
			internal_password_disabled 	= true
			groups      				= [ "readers" ]
		}
	`
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		CheckDestroy: testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userFull, name, id, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "admin", "true"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
					resource.TestCheckResourceAttr(fqrn, "internal_password_disabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(userUpdated, name, id, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "admin", "false"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "internal_password_disabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(username, "name"),
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func TestAccUser_PasswordNotChangeWhenOtherAttributesChangeGH340(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("user-%d", id)
	fqrn := fmt.Sprintf("artifactory_user.%s", name)
	username := fmt.Sprintf("dummy_user%d", id)
	email := fmt.Sprintf("dummy%d@a.com", id)
	password := "Passw0rd!"

	params := map[string]interface{}{
		"name":     name,
		"username": username,
		"email":    email,
		"password": password,
	}
	userInitial := utilsdk.ExecuteTemplate("TestUser", `
		resource "artifactory_user" "{{ .name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			password          = "{{ .password }}"
			groups            = [ "readers" ]
			disable_ui_access = false
		}
	`, params)
	userUpdated := utilsdk.ExecuteTemplate("TestUser", `
		resource "artifactory_user" "{{ .name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			password          = "{{ .password }}"
			groups            = [ "readers" ]
			disable_ui_access = true
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		CheckDestroy: testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: userInitial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", email),
					resource.TestCheckResourceAttr(fqrn, "password", password),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
				),
			},
			{
				Config: userUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", email),
					resource.TestCheckResourceAttr(fqrn, "password", password),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "true"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(username, "name"),
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func testAccCheckManagedUserDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(utilsdk.ProvderMetadata).Client

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
