package user_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccManagedUser_UpgradeFromSDKv2(t *testing.T) {
	id, fqrn, name := testutil.MkNames("test-user-upgrade-", "artifactory_managed_user")
	username := fmt.Sprintf("dummy_user%d", id)
	email := fmt.Sprintf(username + "@test.com")

	params := map[string]interface{}{
		"name":  name,
		"email": email,
	}
	userNoGroups := util.ExecuteTemplate("TestAccUserUpgrade", `
		resource "artifactory_managed_user" "{{ .name }}" {
			name        		= "{{ .name }}"
			email 				= "{{ .email }}"
			password			= "Passsw0rd!12"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "7.7.0",
						Source:            "jfrog/artifactory",
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
				ConfigPlanChecks: testutil.ConfigPlanChecks,
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   userNoGroups,
				PlanOnly:                 true,
				ConfigPlanChecks:         testutil.ConfigPlanChecks,
			},
		},
	})
}

func TestAccManagedUser_no_groups(t *testing.T) {
	const userNoGroups = `
		resource "artifactory_managed_user" "%s" {
			name     = "%s"
			email    = "dummy%d@a.com"
			password = "Passsw0rd!12"
		}
	`
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_managed_user")
	username := fmt.Sprintf("dummy_user%d", id)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userNoGroups, name, username, id),
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

func TestAccManagedUser_empty_groups(t *testing.T) {
	const userEmptyGroups = `
		resource "artifactory_managed_user" "%s" {
			name        		= "%s"
			email       		= "dummy%d@a.com"
			password			= "Passsw0rd!12"
			groups      		= []
		}
	`
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_managed_user")
	username := fmt.Sprintf("dummy_user%d", id)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userEmptyGroups, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
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

func TestAccManagedUser_invalidName(t *testing.T) {
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
		t.Run(tc.name, testAccManagedUserInvalidName(tc.username, tc.errorRegex))
	}
}

func testAccManagedUserInvalidName(username, errorRegex string) func(t *testing.T) {
	return func(t *testing.T) {
		const userNoGroups = `
			resource "artifactory_managed_user" "%s" {
				name  = "%s"
				email = "dummy%d@a.com"
				password = "Passsw0rd!12"
			}
		`
		id, fqrn, name := testutil.MkNames("test-", "artifactory_managed_user")

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

func TestAccManagedUser_basic(t *testing.T) {
	id, fqrn, name := testutil.MkNames("test-user-", "artifactory_managed_user")
	_, _, groupName := testutil.MkNames("test-group-", "artifactory_group")
	username := fmt.Sprintf("dummy_user%d", id)
	email := fmt.Sprintf(username + "@test.com")

	params := map[string]string{
		"name":      name,
		"username":  username,
		"email":     email,
		"groupName": groupName,
	}

	userFull := util.ExecuteTemplate("TestAccManagedUser", `
		resource "artifactory_group" "{{ .groupName }}" {
			name = "{{ .groupName }}"
		}

		resource "artifactory_managed_user" "{{ .name }}" {
			name        		= "{{ .username }}"
			email       		= "{{ .email }}"
			password			= "Passsw0rd!12"
			admin    			= true
			profile_updatable   = true
			disable_ui_access	= false
			groups      		= [
				artifactory_group.{{ .groupName }}.name,
			]
		}
	`, params)

	userNonAdminNoProfUpd := util.ExecuteTemplate("TestAccManagedUser", `
		resource "artifactory_group" "{{ .groupName }}" {
			name = "{{ .groupName }}"
		}

		resource "artifactory_managed_user" "{{ .name }}" {
			name        		= "{{ .username }}"
			email       		= "{{ .email }}"
			password			= "Passsw0rd!12"
			admin    			= false
			profile_updatable   = false
			groups      		= [
				artifactory_group.{{ .groupName }}.name,
			]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: userFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", params["username"]),
					resource.TestCheckResourceAttr(fqrn, "email", params["email"]),
					resource.TestCheckResourceAttr(fqrn, "admin", "true"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "groups.*", params["groupName"]),
				),
			},
			{
				Config: userNonAdminNoProfUpd,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", params["username"]),
					resource.TestCheckResourceAttr(fqrn, "email", params["email"]),
					resource.TestCheckResourceAttr(fqrn, "admin", "false"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "false"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "groups.*", params["groupName"]),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(params["username"], "name"),
				ImportStateVerifyIgnore: []string{"password"}, // password is never returned via the API, so it cannot be "imported"
			},
		},
	})
}
