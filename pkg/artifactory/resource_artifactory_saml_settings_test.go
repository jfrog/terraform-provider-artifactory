package artifactory

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const SamlSettingsTemplateFull = `
resource "artifactory_saml_settings" "saml" {
	enable 					     = true
	certificate                  = "test-certificate"
	email_attribute              = "email"
	group_attribute              = "group"
	login_url                    = "test-login-url"
	logout_url                   = "test-logout-url"
	no_auto_user_creation        = false
	service_provider_name        = "okta"
	allow_user_to_access_profile = true
	auto_redirect                = true
	sync_groups                  = true
	verify_audience_restriction  = true
}`

func TestAccSamlSettings_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		CheckDestroy: testAccSamlSettingsDestroy("artifactory_saml_settings.saml"),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: SamlSettingsTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "enable", "true"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "certificate", "test-certificate"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "email_attribute", "email"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "group_attribute", "group"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "login_url", "test-login-url"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "logout_url", "test-logout-url"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "no_auto_user_creation", "false"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "service_provider_name", "okta"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "allow_user_to_access_profile", "true"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "auto_redirect", "true"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "sync_groups", "true"),
					resource.TestCheckResourceAttr("artifactory_saml_settings.saml", "verify_audience_restriction", "true"),
				),
			},
		},
	})
}

func testAccSamlSettingsDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		c := testAccProvider.Meta().(*resty.Client)


		_, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		samlSettings := SamlSettings{}

		_, err := c.R().SetResult(&samlSettings).Get("artifactory/api/saml/config")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from <base_url>/artifactory/api/saml/config during Read")
		}

		if samlSettings.AllowUserToAccessProfile != false {
			return fmt.Errorf("error: SAML SSO setting, allow user to access profile, is still enabled")
		}
		if samlSettings.SyncGroups != false {
			return fmt.Errorf("error: SAML SSO setting, sync groups, is still enabled")
		}
		if samlSettings.NoAutoUserCreation != false {
			return fmt.Errorf("error: SAML SSO setting, no auto user creation, is still enabled")
		}
		if samlSettings.EnableIntegration != false {
			return fmt.Errorf("error: SAML SSO integration is still enabled")
		}

		return nil
	}
}
