package artifactory

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

const GeneralSecurityTemplateFull = `
resource "artifactory_general_security" "security" {
	enable_anonymous_access = true
}`

func TestAccGeneralSecurity_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccGeneralSecurityDestroy("artifactory_general_security.security"),
		ProviderFactories: utils.TestAccProviders(Provider()),

		Steps: []resource.TestStep{
			{
				Config: GeneralSecurityTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_general_security.security", "enable_anonymous_access", "true"),
				),
			},
		},
	})
}

func testAccGeneralSecurityDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := utils.TestAccProviders(Provider())["artifactory"]()
		provider, err := utils.ConfigureProvider(provider)
		if err != nil {
			return err
		}

		client := provider.Meta().(*resty.Client)

		_, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		generalSettings := GeneralSettings{}
		_, err = client.R().SetResult(&generalSettings).Get("artifactory/api/securityconfig")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from <base_url>/artifactory/api/securityconfig during Read")
		}
		if generalSettings.AnonAccessEnabled != false {
			return fmt.Errorf("error: general security setting to allow anonymous access is still enabled")
		}

		return nil
	}
}
