package user_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccUnmanagedUser_UpgradeFromSDKv2(t *testing.T) {
	id, fqrn, name := testutil.MkNames("test-unmanaged-user-upgrade-", "artifactory_unmanaged_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := fmt.Sprintf(username + "@test.com")

	params := map[string]interface{}{
		"name":  name,
		"email": email,
	}
	config := util.ExecuteTemplate("TestAccUnmanagedUserUpgrade", `
		resource "artifactory_unmanaged_user" "{{ .name }}" {
			name     = "{{ .name }}"
			email 	 = "{{ .email }}"
			password = "Passsw0rd!12"
			disable_ui_access = false
		}
	`, params)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "10.7.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", params["name"].(string)),
					resource.TestCheckResourceAttr(fqrn, "email", params["email"].(string)),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
					resource.TestCheckResourceAttr(fqrn, "internal_password_disabled", "false"),
					resource.TestCheckNoResourceAttr(fqrn, "groups"),
				),
				ConfigPlanChecks: testutil.ConfigPlanChecks,
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				PlanOnly:                 true,
				ConfigPlanChecks:         testutil.ConfigPlanChecks,
			},
		},
	})
}

func TestAccUnmanagedUser_PasswordNotChangeWhenOtherAttributesChangeGH340(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("user-%d", id)
	fqrn := fmt.Sprintf("artifactory_unmanaged_user.%s", name)
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
		resource "artifactory_unmanaged_user" "{{ .name }}" {
			name              = "{{ .username }}"
			email             = "{{ .email }}"
			password          = "{{ .password }}"
			groups            = [ "readers" ]
			disable_ui_access = false
		}
	`, params)
	userUpdated := util.ExecuteTemplate("TestUser", `
		resource "artifactory_unmanaged_user" "{{ .name }}" {
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

func TestAccUnmanagedUser_basic(t *testing.T) {
	const userBasic = `
		resource "artifactory_unmanaged_user" "%s" {
			name  	= "%s"
			password = "Passw0rd!"
			email 	= "dummy_user%d@a.com"
		}
	`
	id := testutil.RandomInt()
	name := fmt.Sprintf("foobar-%d", id)
	fqrn := fmt.Sprintf("artifactory_unmanaged_user.%s", name)
	username := fmt.Sprintf("dummy_user%d", id)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userBasic, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy_user%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
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

func TestAccUnmanagedUser_ShouldCreateUpdateWithoutPassword(t *testing.T) {
	const config = `
		resource "artifactory_unmanaged_user" "%s" {
			name  	= "%s"
			email 	= "dummy_user%d@a.com"
			profile_updatable = true
			disable_ui_access = false
		}
	`

	const updatedConfig = `
		resource "artifactory_unmanaged_user" "%s" {
			name  	= "%s"
			email 	= "dummy_user%d@a.com"
			profile_updatable = true
			disable_ui_access = true
		}
	`

	id := testutil.RandomInt()
	name := fmt.Sprintf("foobar-%d", id)
	fqrn := fmt.Sprintf("artifactory_unmanaged_user.%s", name)
	username := fmt.Sprintf("dummy_user%d", id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy_user%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
					resource.TestCheckNoResourceAttr(fqrn, "password"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(updatedConfig, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy_user%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "true"),
					resource.TestCheckNoResourceAttr(fqrn, "password"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
		},
	})
}

func TestAccUnmanagedUser_internal_password_disabled_changed_error(t *testing.T) {
	const config = `
		resource "artifactory_unmanaged_user" "%s" {
			name  	= "%s"
			email 	= "dummy_user%d@a.com"
			profile_updatable = true
			disable_ui_access = false
			internal_password_disabled = true
		}
	`

	const updatedConfig = `
		resource "artifactory_unmanaged_user" "%s" {
			name  	= "%s"
			email 	= "dummy_user%d@a.com"
			disable_ui_access = true
			internal_password_disabled = false
		}
	`

	id := testutil.RandomInt()
	name := fmt.Sprintf("foobar-%d", id)
	fqrn := fmt.Sprintf("artifactory_unmanaged_user.%s", name)
	username := fmt.Sprintf("dummy_user%d", id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy_user%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "internal_password_disabled", "true"),
					resource.TestCheckNoResourceAttr(fqrn, "password"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
			{
				Config:      fmt.Sprintf(updatedConfig, name, username, id),
				ExpectError: regexp.MustCompile(`Password must be set when internal_password_disabled is changed to 'false'`),
			},
		},
	})
}

func TestAccUnmanagedUser_internal_password_disabled_changed_with_password(t *testing.T) {
	const config = `
		resource "artifactory_unmanaged_user" "%s" {
			name  	= "%s"
			email 	= "dummy_user%d@a.com"
			profile_updatable = true
			disable_ui_access = false
			internal_password_disabled = true
		}
	`

	const updatedConfig = `
		resource "artifactory_unmanaged_user" "%s" {
			name  	= "%s"
			email 	= "dummy_user%d@a.com"
			password = "Password!123"
			disable_ui_access = true
			internal_password_disabled = false
		}
	`

	id := testutil.RandomInt()
	name := fmt.Sprintf("foobar-%d", id)
	fqrn := fmt.Sprintf("artifactory_unmanaged_user.%s", name)
	username := fmt.Sprintf("dummy_user%d", id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy_user%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "internal_password_disabled", "true"),
					resource.TestCheckNoResourceAttr(fqrn, "password"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(updatedConfig, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy_user%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "internal_password_disabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "password", "Password!123"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(config, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy_user%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "internal_password_disabled", "true"),
					resource.TestCheckNoResourceAttr(fqrn, "password"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
		},
	})
}

func TestAccUnmanagedUser_name_change(t *testing.T) {
	const config = `
		resource "artifactory_unmanaged_user" "%s" {
			name  	= "%s"
			email 	= "dummy_user%d@a.com"
			password = "Password!123"
			profile_updatable = true
			disable_ui_access = false
			internal_password_disabled = false
		}
	`

	const nameChangedConfig = `
		resource "artifactory_unmanaged_user" "%s" {
			name  	= "foobar"
			email 	= "dummy_user%d@a.com"
			password = "Password!123"
			profile_updatable = true
			disable_ui_access = false
			internal_password_disabled = false
		}
	`

	id := testutil.RandomInt()
	name := fmt.Sprintf("foobar-%d", id)
	fqrn := fmt.Sprintf("artifactory_unmanaged_user.%s", name)
	username := fmt.Sprintf("dummy_user%d", id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy_user%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "internal_password_disabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "password", "Password!123"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(nameChangedConfig, name, id),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fqrn, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func TestAccUnmanagedUser_full(t *testing.T) {
	id, fqrn, resourceName := testutil.MkNames("test-user-", "artifactory_unmanaged_user")
	_, _, groupName1 := testutil.MkNames("test-group-", "artifactory_group")
	_, _, groupName2 := testutil.MkNames("test-group-", "artifactory_group")

	username := fmt.Sprintf("dummy_user-%d", id)
	email := fmt.Sprintf("dummy%d@a.com", id)

	config := util.ExecuteTemplate(
		"TestAccUnmanagedUser_full",
		`resource "artifactory_group" "{{ .groupName1 }}" {
			name = "{{ .groupName1 }}"
		}

		resource "artifactory_unmanaged_user" "{{ .resourceName }}" {
			name        		= "{{ .username }}"
			email       		= "{{ .email }}"
			password			= "Passw0rd!"
			admin    			= true
			profile_updatable   = true
			disable_ui_access	= false

			groups = [
				artifactory_group.{{ .groupName1 }}.name,
			]
		}`,
		map[string]string{
			"resourceName": resourceName,
			"username":     username,
			"email":        email,
			"groupName1":   groupName1,
		},
	)

	updatedConfig := util.ExecuteTemplate(
		"TestAccUnmanagedUser_full",
		`resource "artifactory_group" "{{ .groupName1 }}" {
			name = "{{ .groupName1 }}"
		}

		resource "artifactory_group" "{{ .groupName2 }}" {
			name = "{{ .groupName2 }}"
		}

		resource "artifactory_unmanaged_user" "{{ .resourceName }}" {
			name        		= "{{ .username }}"
			email       		= "{{ .email }}"
			password			= "Passw0rd!"
			admin    			= false
			profile_updatable   = false
			disable_ui_access	= false

			groups = [
				artifactory_group.{{ .groupName1 }}.name,
				artifactory_group.{{ .groupName2 }}.name,
			]
		}`,
		map[string]string{
			"resourceName": resourceName,
			"username":     username,
			"email":        email,
			"groupName1":   groupName1,
			"groupName2":   groupName2,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", email),
					resource.TestCheckResourceAttr(fqrn, "admin", "true"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "groups.*", groupName1),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "admin", "false"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "false"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "groups.*", groupName1),
					resource.TestCheckTypeSetElemAttr(fqrn, "groups.*", groupName2),
				),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", email),
					resource.TestCheckResourceAttr(fqrn, "admin", "true"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "groups.*", groupName1),
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

func TestAccUnmanagedUser_invalidName(t *testing.T) {
	testCase := []struct {
		name       string
		username   string
		errorRegex string
	}{
		{"Empty", "", `.*Attribute name string length must be at least 1, got: 0.*`},
		{"Uppercase", "test_user_Uppercase", `.*may contain lowercase letters, numbers and symbols: '.-_@'.*`},
		{"Symbols", "test_user_!", `.*may contain lowercase letters, numbers and symbols: '.-_@'.*`},
	}

	for _, tc := range testCase {
		t.Run(tc.name, testAccUnmanagedUserInvalidName(tc.username, tc.errorRegex))
	}
}

func testAccUnmanagedUserInvalidName(username, errorRegex string) func(t *testing.T) {
	return func(t *testing.T) {
		const userNoGroups = `
			resource "artifactory_unmanaged_user" "%s" {
				name  = "%s"
				email = "dummy%d@a.com"
			}
		`
		id, fqrn, name := testutil.MkNames("test-", "artifactory_unmanaged_user")

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

func TestAccUnmanagedUser_EmptyGroups(t *testing.T) {
	const userEmptyGroups = `
		resource "artifactory_unmanaged_user" "%s" {
			name   = "%s"
			email  = "dummy%d@a.com"
			password = "Passw0rd!"
			groups = []
		}
	`
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_unmanaged_user")
	username := fmt.Sprintf("dummy_user%d", id)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userEmptyGroups, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", fmt.Sprintf("dummy_user%d", id)),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
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

func testAccCheckUserDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("resource id[%s] not found", id)
		}

		var resp *resty.Response
		var err error
		// 7.84.3 or later, use Access API
		if ok, e := util.CheckVersion(acctest.Provider.Meta().(util.ProviderMetadata).ArtifactoryVersion, user.AccessAPIArtifactoryVersion); e == nil && ok {
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
