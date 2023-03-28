package configuration_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/util"
)

const OauthSettingsTemplateFull = `
resource "artifactory_oauth_settings" "oauth" {
	enable 						 = true
	persist_users 				 = true
	allow_user_to_access_profile = true

	oauth_provider {
		name 	          = "okta"
		enabled           = false
		type 	          = "openId"
		client_id 		  = "foo"
		client_secret 	  = "bar"
		api_url           = "https://organization.okta.com/oauth2/v1/userinfo"
		auth_url          = "https://organization.okta.com/oauth2/v1/authorize"
		token_url         = "https://organization.okta.com/oauth2/v1/token"
    }
}`

func TestAccOauthSettings_full(t *testing.T) {
	fqrn := "artifactory_oauth_settings.oauth"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccOauthSettingsDestroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: OauthSettingsTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enable", "true"),
					resource.TestCheckResourceAttr(fqrn, "persist_users", "true"),
					resource.TestCheckResourceAttr(fqrn, "allow_user_to_access_profile", "true"),
					resource.TestCheckResourceAttr(fqrn, "oauth_provider.#", "1"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth_provider.0.client_secret"},
			},
		},
	})
}

const OauthSettingsTemplateMultipleProviders = `
resource "artifactory_oauth_settings" "oauth" {
	enable 						 = true
	persist_users 				 = false
	allow_user_to_access_profile = false

	oauth_provider {
		name 	          = "okta"
		enabled           = true
		type 	          = "openId"
		client_id 		  = "foo"
		client_secret 	  = "bar"
		api_url           = "https://organization.okta.com/oauth2/v1/userinfo"
		auth_url          = "https://organization.okta.com/oauth2/v1/authorize"
		token_url         = "https://organization.okta.com/oauth2/v1/token"
    }

	oauth_provider {
		name 	          = "keycloak"
		enabled           = true
		type 	          = "openId"
		client_id 		  = "foo"
		client_secret 	  = "bar"
		api_url           = "https://keycloak.organization.com/auth/realms/test-realm/protocol/openid-connect/userinfo"
		auth_url          = "https://keycloak.organization.com/auth/realms/test-realm/protocol/openid-connect/auth"
		token_url         = "https://keycloak.organization.com/auth/realms/test-realm/protocol/openid-connect/token"
    }
}
`

func TestAccOauthSettings_multipleProviders(t *testing.T) {
	fqrn := "artifactory_oauth_settings.oauth"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccOauthSettingsDestroy("artifactory_oauth_settings.oauth"),

		Steps: []resource.TestStep{
			{
				Config: OauthSettingsTemplateMultipleProviders,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enable", "true"),
					resource.TestCheckResourceAttr(fqrn, "persist_users", "false"),
					resource.TestCheckResourceAttr(fqrn, "allow_user_to_access_profile", "false"),
					resource.TestCheckResourceAttr(fqrn, "oauth_provider.#", "2"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth_provider.0.client_secret", "oauth_provider.1.client_secret"},
			},
		},
	})
}

func testAccOauthSettingsDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProvderMetadata).Client

		_, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		oauthSettings := configuration.OauthSettings{}
		_, err := client.R().SetResult(&oauthSettings).Get("artifactory/api/oauth")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from <base_url>/artifactory/api/oauth during Read")
		}

		if len(oauthSettings.OauthProvidersSettings) > 0 {
			return fmt.Errorf("error: OAuth SSO providers still exist")
		}
		if oauthSettings.AllowUserToAccessProfile != false {
			return fmt.Errorf("error: OAuth SSO setting, allow user to access profile, is still enabled")
		}
		if oauthSettings.PersistUsers != false {
			return fmt.Errorf("error: OAuth SSO setting, persist users, is still enabled")
		}
		if oauthSettings.EnableIntegration != false {
			return fmt.Errorf("error: OAuth SSO integration is still enabled")
		}

		return nil
	}
}
