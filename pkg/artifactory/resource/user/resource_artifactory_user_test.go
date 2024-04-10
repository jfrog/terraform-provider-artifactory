package user_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccUser_UpgradeFromSDKv2(t *testing.T) {
	id, fqrn, name := testutil.MkNames("test-user-upgrade-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := fmt.Sprintf(username + "@test.com")

	params := map[string]interface{}{
		"name":  name,
		"email": email,
	}
	userNoGroups := util.ExecuteTemplate("TestAccUserUpgrade", `
		resource "artifactory_user" "{{ .name }}" {
			name     = "{{ .name }}"
			email 	 = "{{ .email }}"
			password = "Passsw0rd!12"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "7.7.0",
						Source:            "registry.terraform.io/jfrog/artifactory",
					},
				},
				Config: userNoGroups,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", params["name"].(string)),
					resource.TestCheckResourceAttr(fqrn, "email", params["email"].(string)),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "true"),
					resource.TestCheckResourceAttr(fqrn, "internal_password_disabled", "false"),
					resource.TestCheckNoResourceAttr(fqrn, "groups"),
				),
				ConfigPlanChecks: acctest.ConfigPlanChecks,
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   userNoGroups,
				PlanOnly:                 true,
				ConfigPlanChecks:         acctest.ConfigPlanChecks,
			},
		},
	})
}

func TestAccUser_full_groups(t *testing.T) {
	id, fqrn, name := testutil.MkNames("test-user-", "artifactory_user")
	_, _, groupName1 := testutil.MkNames("test-group-", "artifactory_group")
	_, _, groupName2 := testutil.MkNames("test-group-", "artifactory_group")

	params := map[string]string{
		"name":       name,
		"email":      fmt.Sprintf("dummy_user%d@test.com", id),
		"groupName1": groupName1,
		"groupName2": groupName2,
	}
	config := util.ExecuteTemplate("TestAccUserBasic", `
		resource "artifactory_group" "{{ .groupName1 }}" {
			name = "{{ .groupName1 }}"
		}

		resource "artifactory_user" "{{ .name }}" {
			name     = "{{ .name }}"
			email 	 = "{{ .email }}"
			password = "Passsw0rd!12"
			admin 	 = false
			groups   = [
				artifactory_group.{{ .groupName1 }}.name,
			]
		}
	`, params)

	updatedConfig := util.ExecuteTemplate("TestAccUserBasic", `
		resource "artifactory_group" "{{ .groupName1 }}" {
			name = "{{ .groupName1 }}"
		}

		resource "artifactory_group" "{{ .groupName2 }}" {
			name = "{{ .groupName2 }}"
		}

		resource "artifactory_user" "{{ .name }}" {
			name     = "{{ .name }}"
			email 	 = "{{ .email }}"
			password = "Passsw0rd!12"
			admin 	 = false
			groups   = [
				artifactory_group.{{ .groupName1 }}.name,
				artifactory_group.{{ .groupName2 }}.name,
			]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", params["name"]),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "groups.*", params["groupName1"]),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", params["name"]),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "groups.*", params["groupName1"]),
					resource.TestCheckTypeSetElemAttr(fqrn, "groups.*", params["groupName2"]),
				),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", params["name"]),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "groups.*", params["groupName1"]),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "name"),
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
	config := util.ExecuteTemplate("TestAccUserBasic", `
		resource "artifactory_user" "{{ .name }}" {
			name   = "{{ .name }}"
			email  = "{{ .email }}"
			admin  = false
			groups = [ "readers" ]
		}
	`, params)

	updatedConfig := util.ExecuteTemplate("TestAccUserBasic", `
		resource "artifactory_user" "{{ .name }}" {
			name   = "{{ .name }}"
			email  = "{{ .email }}"
			admin  = false
			profile_updatable = false
			groups = [ "readers" ]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("foobar-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("foobar-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "false"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "name"),
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func TestAccUser_no_groups(t *testing.T) {
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := fmt.Sprintf(username + "@test.com")

	params := map[string]interface{}{
		"name":     fmt.Sprintf("foobar-%d", id),
		"username": username,
		"email":    email,
	}
	userEmptyGroups := util.ExecuteTemplate("TestAccUserBasic", `
		resource "artifactory_user" "{{ .name }}" {
			name        		= "{{ .name }}"
			email 				= "{{ .email }}"
			password			= "Passsw0rd!12"
			admin 				= false
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: userEmptyGroups,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("foobar-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "name"),
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
	userEmptyGroups := util.ExecuteTemplate("TestAccUserBasic", `
		resource "artifactory_user" "{{ .name }}" {
			name        		= "{{ .name }}"
			email 				= "{{ .email }}"
			password			= "Passsw0rd!12"
			admin 				= false
			groups      		= []
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: userEmptyGroups,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("foobar-%d", id)),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "name"),
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}

func TestAccUser_invalidName(t *testing.T) {
	testCase := []struct {
		name       string
		username   string
		errorRegex string
	}{
		{"Empty", "", `.*Invalid Attribute Value Length.*`},
		{"Uppercase", "test_user_Uppercase", `.*may contain lowercase letters, numbers and symbols: '.-_@'.*`},
		{"Symbols", "test_user_!", `.*may contain lowercase letters, numbers and symbols: '.-_@'.*`},
	}

	for _, tc := range testCase {
		t.Run(tc.name, testAccUserInvalidName(tc.username, tc.errorRegex))
	}
}

func testAccUserInvalidName(username, errorRegex string) func(t *testing.T) {
	return func(t *testing.T) {
		const userNoGroups = `
			resource "artifactory_user" "%s" {
				name  = "%s"
				email = "dummy%d@a.com"
			}
		`
		id, fqrn, name := testutil.MkNames("test-", "artifactory_user")

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.PreCheck(t) },
			ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckUserDestroy(fqrn),
			Steps: []resource.TestStep{
				{
					Config:      fmt.Sprintf(userNoGroups, name, username, id),
					ExpectError: regexp.MustCompile(errorRegex),
				},
			},
		})
	}
}

func TestAccUser_all_attributes(t *testing.T) {
	const userFull = `
		resource "artifactory_user" "%s" {
			name        				= "%s"
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
			name        				= "%s"
			email       				= "dummy%d@a.com"
			password					= "Passw0rd!"
			admin    					= false
			profile_updatable   		= true
			internal_password_disabled 	= false
			groups      				= [ "readers" ]
		}
	`
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_user")
	username := fmt.Sprintf("dummy_user-@%d.", id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userFull, name, username, id),
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
				Config: fmt.Sprintf(userUpdated, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "admin", "false"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "internal_password_disabled", "false"),
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
	userInitial := util.ExecuteTemplate("TestUser", `
		resource "artifactory_user" "{{ .name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			password          = "{{ .password }}"
			groups            = [ "readers" ]
			disable_ui_access = false
		}
	`, params)
	userUpdated := util.ExecuteTemplate("TestUser", `
		resource "artifactory_user" "{{ .name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			password          = "{{ .password }}"
			groups            = [ "readers" ]
			disable_ui_access = true
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(fqrn),
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
		client := acctest.Provider.Meta().(util.ProvderMetadata).Client

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("resource id[%s] not found", id)
		}

		var resp *resty.Response
		var err error
		// 7.49.3 or later, use Access API
		if ok, e := util.CheckVersion(acctest.Provider.Meta().(util.ProvderMetadata).ArtifactoryVersion, "7.49.3"); e == nil && ok {
			r, er := client.R().Get("access/api/v2/users/" + rs.Primary.ID)
			resp = r
			err = er
		} else {
			r, er := client.R().Get("artifactory/api/security/users/" + rs.Primary.ID)
			resp = r
			err = er
		}

		if err != nil {
			return err
		}

		if resp.StatusCode() == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("user %s still exists", rs.Primary.ID)
	}
}
