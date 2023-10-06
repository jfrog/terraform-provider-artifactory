package security_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func TestAccScopedToken_UpgradeFromSDKv2(t *testing.T) {
	// Version 7.11.1 is the last version before we migrated the resource from SDKv2 to Plugin Framework
	version := "7.11.1"
	title := fmt.Sprintf("from_v%s", version)
	t.Run(title, func(t *testing.T) {
		resource.Test(scopedTokenUpgradeTestCase(version, t))
	})
}

func TestAccScopedToken_UpgradeGH_758(t *testing.T) {
	// Version 7.2.0 doesn't have `include_reference_token` attribute
	// This test verifies that there is no state drift on update
	version := "7.2.0"
	title := fmt.Sprintf("from_v%s", version)
	t.Run(title, func(t *testing.T) {
		resource.Test(scopedTokenUpgradeTestCase(version, t))
	})
}

func TestAccScopedToken_UpgradeGH_792(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")
	config := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.test-user.name
		    expires_in  = 31536000
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "7.11.2",
						Source:            "registry.terraform.io/jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "testuser"),
					resource.TestCheckNoResourceAttr(fqrn, "description"),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "expires_in", "31536000"),
					resource.TestCheckNoResourceAttr(fqrn, "audiences"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckNoResourceAttr(fqrn, "refresh_token"),
					resource.TestCheckNoResourceAttr(fqrn, "reference_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
					resource.TestCheckResourceAttrSet(fqrn, "subject"),
					resource.TestCheckResourceAttrSet(fqrn, "expiry"),
					resource.TestCheckResourceAttrSet(fqrn, "issued_at"),
					resource.TestCheckResourceAttrSet(fqrn, "issuer"),
				),
				ConfigPlanChecks: acctest.ConfigPlanChecks,
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5MuxProviderFactories,
				Config:                   config,
				PlanOnly:                 true,
				ConfigPlanChecks:         acctest.ConfigPlanChecks,
			},
		},
	})
}

func TestAccScopedToken_UpgradeGH_818(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-scope-token", "artifactory_scoped_token")
	config := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			scopes   = ["applied-permissions/user"]
			username = artifactory_user.test-user.name
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "7.2.0",
						Source:            "registry.terraform.io/jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "testuser"),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "expires_in", "31536000"),
					resource.TestCheckNoResourceAttr(fqrn, "audiences"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckNoResourceAttr(fqrn, "refresh_token"),
					resource.TestCheckNoResourceAttr(fqrn, "reference_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
					resource.TestCheckResourceAttrSet(fqrn, "subject"),
					resource.TestCheckResourceAttrSet(fqrn, "expiry"),
					resource.TestCheckResourceAttrSet(fqrn, "issued_at"),
					resource.TestCheckResourceAttrSet(fqrn, "issuer"),
				),
				ConfigPlanChecks: acctest.ConfigPlanChecks,
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5MuxProviderFactories,
				Config:                   config,
				PlanOnly:                 true,
				ConfigPlanChecks:         acctest.ConfigPlanChecks,
			},
		},
	})
}

func scopedTokenUpgradeTestCase(version string, t *testing.T) (*testing.T, resource.TestCase) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	config := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.test-user.name
		    expires_in  = 31536000
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	return t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: version,
						Source:            "registry.terraform.io/jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "testuser"),
					resource.TestCheckResourceAttr(fqrn, "description", ""),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "expires_in", "31536000"),
					resource.TestCheckNoResourceAttr(fqrn, "audiences"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckNoResourceAttr(fqrn, "refresh_token"),
					resource.TestCheckNoResourceAttr(fqrn, "reference_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
					resource.TestCheckResourceAttrSet(fqrn, "subject"),
					resource.TestCheckResourceAttrSet(fqrn, "expiry"),
					resource.TestCheckResourceAttrSet(fqrn, "issued_at"),
					resource.TestCheckResourceAttrSet(fqrn, "issuer"),
				),
				ConfigPlanChecks: acctest.ConfigPlanChecks,
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   config,
				PlanOnly:                 true,
				ConfigPlanChecks:         acctest.ConfigPlanChecks,
			},
		},
	}
}

func TestAccScopedToken_WithDefaults(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	template := `resource "artifactory_user" "test-user" {
		name              = "testuser"
		email             = "testuser@tempurl.org"
		admin             = true
		disable_ui_access = false
		groups            = ["readers"]
		password          = "Passw0rd!"
	}

	resource "artifactory_scoped_token" "{{ .name }}" {
		username    = artifactory_user.test-user.name
		description = "{{ .description }}"
	}`

	accessTokenConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		template,
		map[string]interface{}{
			"name":        name,
			"description": "",
		},
	)

	accessTokenUpdatedConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		template,
		map[string]interface{}{
			"name":        name,
			"description": "test updated description",
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, security.CheckAccessToken),
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "testuser"),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/user"),
					resource.TestCheckResourceAttr(fqrn, "refreshable", "false"),
					resource.TestCheckResourceAttr(fqrn, "description", ""),
					resource.TestCheckNoResourceAttr(fqrn, "audiences"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckNoResourceAttr(fqrn, "refresh_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
					resource.TestCheckResourceAttrSet(fqrn, "subject"),
					resource.TestCheckResourceAttrSet(fqrn, "expiry"),
					resource.TestCheckResourceAttrSet(fqrn, "issued_at"),
					resource.TestCheckResourceAttrSet(fqrn, "issuer"),
					resource.TestCheckNoResourceAttr(fqrn, "reference_token"),
				),
			},
			{
				Config: accessTokenUpdatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "description", "test updated description"),
				),
			},
			{
				ResourceName: fqrn,
				ImportState:  true,
				ExpectError:  regexp.MustCompile("resource artifactory_scoped_token doesn't support import"),
			},
		},
	})
}

func TestAccScopedToken_WithAttributes(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")
	projectKey := fmt.Sprintf("test-project-%d", testutil.RandomInt())

	accessTokenConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.test-user.name
			project_key = "{{ .projectKey }}"
			scopes      = ["applied-permissions/admin", "system:metrics:r"]
			description = "test description"
			refreshable = true
			expires_in  = 0
			audiences   = ["jfrt@1", "jfxr@*"]
		}`,
		map[string]interface{}{
			"name":       name,
			"projectKey": projectKey,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: acctest.VerifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return security.CheckAccessToken(id, request)
		}),
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "testuser"),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/admin"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "system:metrics:r"),
					resource.TestCheckResourceAttr(fqrn, "refreshable", "true"),
					resource.TestCheckResourceAttr(fqrn, "expires_in", "0"),
					resource.TestCheckResourceAttr(fqrn, "description", "test description"),
					resource.TestCheckResourceAttr(fqrn, "audiences.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "audiences.*", "jfrt@1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "audiences.*", "jfxr@*"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckResourceAttrSet(fqrn, "refresh_token"),
					resource.TestCheckNoResourceAttr(fqrn, "reference_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
					resource.TestCheckResourceAttrSet(fqrn, "subject"),
					resource.TestCheckResourceAttrSet(fqrn, "expiry"),
					resource.TestCheckResourceAttrSet(fqrn, "issued_at"),
					resource.TestCheckResourceAttrSet(fqrn, "issuer"),
					resource.TestCheckResourceAttr(fqrn, "include_reference_token", "false"),
				),
			},
			{
				ResourceName: fqrn,
				ImportState:  true,
				ExpectError:  regexp.MustCompile("resource artifactory_scoped_token doesn't support import"),
			},
		},
	})
}

func TestAccScopedToken_WithGroupScope(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	accessTokenConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_group" "test-group" {
			name = "{{ .groupName }}"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_group.test-group.name
			scopes      = ["applied-permissions/groups:{{ .groupName }}"]
		}`,
		map[string]interface{}{
			"name":      name,
			"groupName": "test-group",
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "test-group"),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/groups:test-group"),
				),
			},
			{
				ResourceName: fqrn,
				ImportState:  true,
				ExpectError:  regexp.MustCompile("resource artifactory_scoped_token doesn't support import"),
			},
		},
	})
}

func TestAccScopedToken_WithInvalidScopes(t *testing.T) {
	_, _, name := testutil.MkNames("test-scoped-token", "artifactory_scoped_token")

	scopedTokenConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_scoped_token" "{{ .name }}" {
			scopes      = ["foo"]
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      scopedTokenConfig,
				ExpectError: regexp.MustCompile(`.*Invalid Attribute Value Match.*`),
			},
		},
	})
}

func TestAccScopedToken_WithTooLongScopes(t *testing.T) {
	_, _, name := testutil.MkNames("test-scoped-token", "artifactory_scoped_token")

	scopedTokenConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_local_generic_repository" "generic-local-1" {
			key = "generic-local-1"
		}

		resource "artifactory_local_generic_repository" "generic-local-2" {
			key = "generic-local-2"
		}

		resource "artifactory_local_generic_repository" "generic-local-3" {
			key = "generic-local-3"
		}

		resource "artifactory_local_generic_repository" "generic-local-4" {
			key = "generic-local-4"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			scopes      = [
				"applied-permissions/admin",
				"applied-permissions/user",
				"system:metrics:r",
				"system:livelogs:r",
				"artifact:generic-local-1:r",
				"artifact:generic-local-1:w",
				"artifact:generic-local-1:d",
				"artifact:generic-local-1:a",
				"artifact:generic-local-1:m",
				"artifact:generic-local-2:r",
				"artifact:generic-local-2:w",
				"artifact:generic-local-2:d",
				"artifact:generic-local-2:a",
				"artifact:generic-local-2:m",
				"artifact:generic-local-3:r",
				"artifact:generic-local-3:w",
				"artifact:generic-local-3:d",
				"artifact:generic-local-3:a",
				"artifact:generic-local-3:m",
				"artifact:generic-local-4:r",
				"artifact:generic-local-4:w",
				"artifact:generic-local-4:d",
				"artifact:generic-local-4:a",
				"artifact:generic-local-4:m",
			]

			depends_on = [
				artifactory_local_generic_repository.generic-local-1,
				artifactory_local_generic_repository.generic-local-2,
				artifactory_local_generic_repository.generic-local-3,
				artifactory_local_generic_repository.generic-local-4,
			]
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      scopedTokenConfig,
				ExpectError: regexp.MustCompile(".*Scopes length exceeds 500 characters.*"),
			},
		},
	})
}

func TestAccScopedToken_WithAudience(t *testing.T) {

	for _, prefix := range []string{"jfrt", "jfxr", "jfpip", "jfds", "jfmc", "jfac", "jfevt", "jfmd", "jfcon", "*"} {
		t.Run(prefix, func(t *testing.T) {
			resource.Test(mkAudienceTestCase(prefix, t))
		})
	}
}

func mkAudienceTestCase(prefix string, t *testing.T) (*testing.T, resource.TestCase) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	accessTokenConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_scoped_token" "{{ .name }}" {
			audiences = ["{{ .prefix }}@*"]
		}`,
		map[string]interface{}{
			"name":   name,
			"prefix": prefix,
		},
	)

	return t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "audiences.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "audiences.*", fmt.Sprintf("%s@*", prefix)),
				),
			},
			{
				ResourceName: fqrn,
				ImportState:  true,
				ExpectError:  regexp.MustCompile("resource artifactory_scoped_token doesn't support import"),
			},
		},
	}
}

func TestAccScopedToken_WithInvalidAudiences(t *testing.T) {
	_, _, name := testutil.MkNames("test-scoped-token", "artifactory_scoped_token")

	scopedTokenConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_scoped_token" "{{ .name }}" {
			audiences = ["foo@*"]
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      scopedTokenConfig,
				ExpectError: regexp.MustCompile(`.*must either begin with jfrt, jfxr, jfpip,.*`),
			},
		},
	})
}

func TestAccScopedToken_WithTooLongAudiences(t *testing.T) {
	_, _, name := testutil.MkNames("test-scoped-token", "artifactory_scoped_token")

	var audiences []string
	for i := 0; i < 100; i++ {
		audiences = append(audiences, fmt.Sprintf("jfrt@%d", i))
	}

	scopedTokenConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_scoped_token" "{{ .name }}" {
			audiences    = [
				{{range .audiences}}"{{.}}",{{end}}
			]
		}`,
		map[string]interface{}{
			"name":      name,
			"audiences": audiences,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      scopedTokenConfig,
				ExpectError: regexp.MustCompile(".*Audiences length exceeds 255 characters.*"),
			},
		},
	})
}

func TestAccScopedToken_WithExpiresInLessThanPersistencyThreshold(t *testing.T) {
	_, _, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	accessTokenConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.test-user.name
			description = "test description"
			expires_in  = {{ .expires_in }}
		}`,
		map[string]interface{}{
			"name":       name,
			"expires_in": 600, // any value > 0 and less than default persistency threshold (10800) will result in token not being saved.
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      accessTokenConfig,
				ExpectError: regexp.MustCompile("Unable to Create Resource"),
			},
		},
	})
}

func TestAccScopedToken_WithExpiresInSetToZeroForNonExpiringToken(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-access-token", "artifactory_scoped_token")

	accessTokenConfig := utilsdk.ExecuteTemplate(
		"TestAccScopedToken",
		`resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.test-user.name
			description = "test description"
			expires_in  = 0
		}`,
		map[string]interface{}{
			"name": name,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check:  resource.TestCheckResourceAttr(fqrn, "expires_in", "0"),
			},
		},
	})
}
