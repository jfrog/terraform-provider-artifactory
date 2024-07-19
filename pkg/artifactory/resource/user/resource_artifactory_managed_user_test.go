package user_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccManagedUser_UpgradeFromSDKv2(t *testing.T) {
	providerHost := os.Getenv("TF_ACC_PROVIDER_HOST")
	if providerHost == "registry.opentofu.org" {
		t.Skipf("provider host is registry.opentofu.org. Previous version of Artifactory provider is unknown to OpenTofu.")
	}

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
				ConfigPlanChecks: testutil.ConfigPlanChecks(""),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   userNoGroups,
				PlanOnly:                 true,
				ConfigPlanChecks:         testutil.ConfigPlanChecks(""),
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

func TestAccManagedUser_default_password_policy(t *testing.T) {
	testCase := []struct {
		name       string
		password   string
		errorRegex string
	}{
		{"Uppercase", "abcd1234", `.*Attribute password string must have at least 1 uppercase letters.*`},
		{"Lowercase", "ABCD1234", `.*Attribute password string must have at least 1 lowercase letters.*`},
		{"Digit", "ABCDefgh", `.*Attribute password string must have at least 1 digits.*`},
		{"Length", "Abc123", `.*Attribute password string length must be at least.*`},
	}

	for _, tc := range testCase {
		t.Run(tc.name, testAccManagedUserDefaultPasswordPolicy(tc.password, tc.errorRegex))
	}
}

func testAccManagedUserDefaultPasswordPolicy(password, errorRegex string) func(t *testing.T) {
	return func(t *testing.T) {
		id, fqrn, name := testutil.MkNames("test-", "artifactory_managed_user")

		temp := `
			resource "artifactory_managed_user" "{{ .resourceName }}" {
				name  = "{{ .name }}"
				password = "{{ .password }}"
				email = "{{ .email }}"
			}
		`

		config := util.ExecuteTemplate("TestAccManagedUser_password_policy", temp, map[string]string{
			"resourceName": name,
			"name":         fmt.Sprintf("test-%d", id),
			"password":     password,
			"email":        fmt.Sprintf("test-%d@test.com", id),
		})

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.PreCheck(t) },
			ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckUserDestroy(fqrn),
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(errorRegex),
				},
			},
		})
	}
}

func TestAccManagedUser_password_policy(t *testing.T) {
	testCase := []struct {
		name       string
		password   string
		errorRegex string
	}{
		{"Uppercase", "-A1b2c3d4e-", `.*Attribute password string must have at least 2 uppercase letters.*`},
		{"Lowercase", "-A1B2C3D4e-", `.*Attribute password string must have at least 2 lowercase letters.*`},
		{"Special Char", "A1B2CDefgh-", `.*Attribute password string must have at least 2 special characters.*`},
		{"Digit", "-AfBgChDiE1-", `.*Attribute password string must have at least 2 digits.*`},
		{"Length", "-A1B2c3d-", `.*Attribute password string length must be at least.*`},
	}

	for _, tc := range testCase {
		t.Run(tc.name, testAccManagedUserPasswordPolicy(tc.password, tc.errorRegex))
	}
}

func testAccManagedUserPasswordPolicy(password, errorRegex string) func(t *testing.T) {
	return func(t *testing.T) {
		id, fqrn, name := testutil.MkNames("test-", "artifactory_managed_user")

		temp := `
			resource "artifactory_managed_user" "{{ .resourceName }}" {
				name  = "{{ .name }}"
				password = "{{ .password }}"
				password_policy = {
					uppercase = 2
					lowercase = 2
					special_char = 2
					digit = 2
					length = 10
				}
				email = "{{ .email }}"
			}
		`

		config := util.ExecuteTemplate("TestAccManagedUser_password_policy", temp, map[string]string{
			"resourceName": name,
			"name":         fmt.Sprintf("test-%d", id),
			"password":     password,
			"email":        fmt.Sprintf("test-%d@test.com", id),
		})

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.PreCheck(t) },
			ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckUserDestroy(fqrn),
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(errorRegex),
				},
			},
		})
	}
}

func TestAccManagedUser_password_policy_interpolated(t *testing.T) {
	testCase := []struct {
		name             string
		passwordCriteria string
		errorRegex       string
	}{
		{
			"Uppercase",
			`length = 10
			min_lower = 2
			upper = false
			min_numeric = 2
			min_special = 2
			override_special = "!"#$%%&'()*+,-./:;<=>?@[\]^_\x60{|}~"`,
			`.*Attribute password string must have at least 2 uppercase letters.*`,
		},
		{
			"Lowercase",
			`length = 10
			lower = false
			min_upper = 2
			min_numeric = 2
			min_special = 2
			override_special = "!"#$%%&'()*+,-./:;<=>?@[\]^_\x60{|}~"`,
			`.*Attribute password string must have at least 2 lowercase letters.*`,
		},
		{
			"Special Char",
			`length = 10
			min_lower = 2
			min_upper = 2
			min_numeric = 2
			special = false
			override_special = "!"#$%%&'()*+,-./:;<=>?@[\]^_\x60{|}~"`,
			`.*Attribute password string must have at least 2 special characters.*`,
		},
		{
			"Digit",
			`length = 10
			min_lower = 2
			min_upper = 2
			numeric = false
			min_special = 2
			override_special = "!"#$%%&'()*+,-./:;<=>?@[\]^_\x60{|}~"`,
			`.*Attribute password string must have at least 2 digits.*`,
		},
		{
			"Length",
			`length = 9
			min_lower = 2
			min_upper = 2
			min_numeric = 2
			min_special = 2
			override_special = "!"#$%%&'()*+,-./:;<=>?@[\]^_\x60{|}~"`,
			`.*Attribute password string length must be at least.*`,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, testAccManagedUserPasswordPolicyInterpolated(tc.passwordCriteria, tc.errorRegex))
	}
}

func testAccManagedUserPasswordPolicyInterpolated(passwordCriteria, errorRegex string) func(t *testing.T) {
	return func(t *testing.T) {
		id, _, name := testutil.MkNames("test-", "artifactory_user")

		temp := `
		resource "random_password" "test" {
			{{ .passwordCriteria }}
		}

		resource "artifactory_managed_user" "{{ .resourceName }}" {
			name  = "{{ .name }}"
			password = random_password.test.result
			password_policy = {
				uppercase = 2
				lowercase = 2
				special_char = 2
				digit = 2
				length = 10
			}
			email = "{{ .email }}"
		}`

		config := util.ExecuteTemplate("TestAccUser_password_policy", temp, map[string]string{
			"resourceName":     name,
			"name":             fmt.Sprintf("test-%d", id),
			"email":            fmt.Sprintf("test-%d@test.com", id),
			"passwordCriteria": passwordCriteria,
		})

		resource.Test(t, resource.TestCase{
			PreCheck: func() { acctest.PreCheck(t) },
			ExternalProviders: map[string]resource.ExternalProvider{
				"random": {
					Source: "hashicorp/random",
				},
			},
			ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      config,
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

func TestAccManagedUser_name_change(t *testing.T) {
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
		resource "artifactory_managed_user" "{{ .name }}" {
			name        		= "{{ .username }}"
			email       		= "{{ .email }}"
			password			= "Passsw0rd!12"
			admin    			= true
			profile_updatable   = true
			disable_ui_access	= false
		}
	`, params)

	usernameChangedConfig := util.ExecuteTemplate("TestAccManagedUser", `
		resource "artifactory_managed_user" "{{ .name }}" {
			name        		= "foobar"
			email       		= "{{ .email }}"
			password			= "Passsw0rd!12"
			admin    			= false
			profile_updatable   = false
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
					resource.TestCheckResourceAttr(fqrn, "groups.#", "0"),
				),
			},
			{
				Config: usernameChangedConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fqrn, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}
