package security

import (
	"fmt"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccOauthSettingsDestroy("artifactory_oauth_settings.oauth"),
		ProviderFactories: artifactory.TestAccProviders,

		Steps: []resource.TestStep{
			{
				Config: OauthSettingsTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_oauth_settings.oauth", "enable", "true"),
					resource.TestCheckResourceAttr("artifactory_oauth_settings.oauth", "persist_users", "true"),
					resource.TestCheckResourceAttr("artifactory_oauth_settings.oauth", "allow_user_to_access_profile", "true"),
					resource.TestCheckResourceAttr("artifactory_oauth_settings.oauth", "oauth_provider.#", "1"),
				),
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
	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccOauthSettingsDestroy("artifactory_oauth_settings.oauth"),
		ProviderFactories: artifactory.TestAccProviders,

		Steps: []resource.TestStep{
			{
				Config: OauthSettingsTemplateMultipleProviders,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_oauth_settings.oauth", "enable", "true"),
					resource.TestCheckResourceAttr("artifactory_oauth_settings.oauth", "persist_users", "false"),
					resource.TestCheckResourceAttr("artifactory_oauth_settings.oauth", "allow_user_to_access_profile", "false"),
					resource.TestCheckResourceAttr("artifactory_oauth_settings.oauth", "oauth_provider.#", "2"),
				),
			},
		},
	})
}

func testAccOauthSettingsDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := artifactory.TestAccProviders["artifactory"]()
		client := provider.Meta().(*resty.Client)

		_, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		oauthSettings := OauthSettings{}
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
