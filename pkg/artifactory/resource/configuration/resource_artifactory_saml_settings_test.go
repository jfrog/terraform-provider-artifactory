package configuration_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/configuration"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
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
    use_encrypted_assertion      = false
}`

func TestAccSamlSettings_full(t *testing.T) {
	fqrn := "artifactory_saml_settings.saml"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccSamlSettingsDestroy("artifactory_saml_settings.saml"),

		Steps: []resource.TestStep{
			{
				Config: SamlSettingsTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enable", "true"),
					resource.TestCheckResourceAttr(fqrn, "certificate", "test-certificate"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "email"),
					resource.TestCheckResourceAttr(fqrn, "group_attribute", "group"),
					resource.TestCheckResourceAttr(fqrn, "login_url", "test-login-url"),
					resource.TestCheckResourceAttr(fqrn, "logout_url", "test-logout-url"),
					resource.TestCheckResourceAttr(fqrn, "no_auto_user_creation", "false"),
					resource.TestCheckResourceAttr(fqrn, "service_provider_name", "okta"),
					resource.TestCheckResourceAttr(fqrn, "allow_user_to_access_profile", "true"),
					resource.TestCheckResourceAttr(fqrn, "auto_redirect", "true"),
					resource.TestCheckResourceAttr(fqrn, "sync_groups", "true"),
					resource.TestCheckResourceAttr(fqrn, "verify_audience_restriction", "true"),
					resource.TestCheckResourceAttr(fqrn, "use_encrypted_assertion", "false"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSamlSettingsDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		c := acctest.Provider.Meta().(utilsdk.ProvderMetadata).Client

		_, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		samlSettings := configuration.SamlSettings{}

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
