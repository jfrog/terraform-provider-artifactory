package user_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/provider"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
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
	userNoGroups := utilsdk.ExecuteTemplate("TestAccUserUpgrade", `
		resource "artifactory_managed_user" "{{ .name }}" {
			name        		= "{{ .name }}"
			email 				= "{{ .email }}"
			password			= "Passsw0rd!"
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
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   userNoGroups,
				PlanOnly:                 true,
				ConfigPlanChecks:         acctest.ConfigPlanChecks,
			},
		},
	})
}

func TestAccManagedUser_no_groups(t *testing.T) {
	const userNoGroups = `
		resource "artifactory_managed_user" "%s" {
			name     = "%s"
			email    = "dummy%d@a.com"
			password = "Passsw0rd!"
		}
	`
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_managed_user")
	username := fmt.Sprintf("dummy_user%d", id)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		CheckDestroy: testAccCheckManagedUserDestroy(fqrn),
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
			password			= "Passsw0rd!"
			groups      		= []
		}
	`
	id, fqrn, name := testutil.MkNames("foobar-", "artifactory_managed_user")
	username := fmt.Sprintf("dummy_user%d", id)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		CheckDestroy: testAccCheckManagedUserDestroy(fqrn),
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
		t.Run(tc.name, testAccManagedUserInvalidName(t, tc.username, tc.errorRegex))
	}
}

func testAccManagedUserInvalidName(t *testing.T, username, errorRegex string) func(t *testing.T) {
	return func(t *testing.T) {
		const userNoGroups = `
			resource "artifactory_managed_user" "%s" {
				name  = "%s"
				email = "dummy%d@a.com"
			}
		`
		id, fqrn, name := testutil.MkNames("test-", "artifactory_managed_user")

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.PreCheck(t) },
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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
	const userFull = `
		resource "artifactory_managed_user" "%s" {
			name        		= "%s"
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
			name        		= "%s"
			email       		= "dummy%d@a.com"
			password			= "Passsw0rd!"
			admin    			= false
			profile_updatable   = false
			groups      		= [ "readers" ]
		}
	`
	id, fqrn, name := testutil.MkNames("test-", "artifactory_managed_user")
	username := fmt.Sprintf("dummy_user-@%d.", id)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		CheckDestroy: testAccCheckManagedUserDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(userFull, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "admin", "true"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(userNonAdminNoProfUpd, name, username, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", username),
					resource.TestCheckResourceAttr(fqrn, "email", fmt.Sprintf("dummy%d@a.com", id)),
					resource.TestCheckResourceAttr(fqrn, "admin", "false"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "false"),
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
