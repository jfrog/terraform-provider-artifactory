package security_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/security"
)

func TestAccScopedToken_WithDefaults(t *testing.T) {
	_, fqrn, name := acctest.MkNames("test-access-token", "artifactory_scoped_token")

	accessTokenConfig := acctest.ExecuteTemplate("TestAccScopedToken", `
		resource "artifactory_user" "test-user" {
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
		}
	`, map[string]interface{}{
		"name": name,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, security.CheckAccessToken),
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "testuser"),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/user"),
					resource.TestCheckResourceAttr(fqrn, "expires_in", "31536000"),
					resource.TestCheckResourceAttr(fqrn, "refreshable", "false"),
					resource.TestCheckResourceAttr(fqrn, "description", "test description"),
					resource.TestCheckNoResourceAttr(fqrn, "audiences"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
				),
			},
		},
	})
}

func TestAccScopedToken_WithAttributes(t *testing.T) {
	_, fqrn, name := acctest.MkNames("test-access-token", "artifactory_scoped_token")

	accessTokenConfig := acctest.ExecuteTemplate("TestAccScopedToken", `
		resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.test-user.name
			scopes      = ["applied-permissions/admin", "system:metrics:r"]
			description = "test description"
			refreshable = true
			audiences   = ["admin"]
		}
	`, map[string]interface{}{
		"name": name,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accessTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "username", "testuser"),
					resource.TestCheckResourceAttr(fqrn, "scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "applied-permissions/admin"),
					resource.TestCheckTypeSetElemAttr(fqrn, "scopes.*", "system:metrics:r"),
					resource.TestCheckResourceAttr(fqrn, "refreshable", "true"),
					resource.TestCheckResourceAttr(fqrn, "description", "test description"),
					resource.TestCheckResourceAttr(fqrn, "audiences.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "audiences.0", "admin"),
					resource.TestCheckResourceAttrSet(fqrn, "access_token"),
					resource.TestCheckResourceAttr(fqrn, "token_type", "Bearer"),
				),
			},
		},
	})
}

func TestAccScopedToken_WithInvalidScopes(t *testing.T) {
	_, _, name := acctest.MkNames("test-scoped-token", "artifactory_scoped_token")

	scopedTokenConfig := acctest.ExecuteTemplate("TestAccScopedToken", `
		resource "artifactory_local_generic_repository" "generic-local-1" {
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

		resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.test-user.name
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
		}
	`, map[string]interface{}{
		"name": name,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      scopedTokenConfig,
				ExpectError: regexp.MustCompile(".*Total combined length of scopes field exceeds 500 characters:.*"),
			},
		},
	})
}

func TestAccScopedToken_WithInvalidAudiences(t *testing.T) {
	_, _, name := acctest.MkNames("test-scoped-token", "artifactory_scoped_token")

	audences := []string{}
	for i := 0; i < 100; i++ {
		audences = append(audences, fmt.Sprintf("audence-%d", i))
	}

	scopedTokenConfig := acctest.ExecuteTemplate("TestAccScopedToken", `
		resource "artifactory_user" "test-user" {
			name              = "testuser"
		    email             = "testuser@tempurl.org"
			admin             = true
			disable_ui_access = false
			groups            = ["readers"]
			password          = "Passw0rd!"
		}

		resource "artifactory_scoped_token" "{{ .name }}" {
			username    = artifactory_user.test-user.name
			scopes      = [
				"applied-permissions/admin",
				"applied-permissions/user",
			]

			audiences    = [
				{{range .audiences}}"{{.}}",{{end}}
			]
		}
	`, map[string]interface{}{
		"name":      name,
		"audiences": audences,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      scopedTokenConfig,
				ExpectError: regexp.MustCompile(".*Total combined length of audences field exceeds 255 characters:.*"),
			},
		},
	})
}
