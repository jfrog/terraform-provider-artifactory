package artifactory

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const GeneralSecurityTemplateFull = `
resource "artifactory_general_security" "security" {
	enable_anonymous_access = true
}`

func TestAccGeneralSecurity_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		CheckDestroy: testAccGeneralSecurityDestroy("artifactory_general_security.security"),
		Providers:    testAccProviders,

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
		apis := testAccProvider.Meta().(*ArtClient)
		c := apis.ArtNew

		serviceDetails := c.GetConfig().GetServiceDetails()
		httpClientDetails := serviceDetails.CreateHttpClientDetails()

		_, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		_, body, _, err := c.Client().SendGet(fmt.Sprintf("%sapi/securityconfig", serviceDetails.GetUrl()), false, &httpClientDetails)
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from <base_url>/artifactory/api/securityconfig during Read")
		}

		generalSettings := GeneralSettings{}
		err = json.Unmarshal(body, &generalSettings)
		if err != nil {
			return fmt.Errorf("error: failed to unmarshal general security settings")
		} else if generalSettings.AnonAccessEnabled != false {
			return fmt.Errorf("error: general security setting to allow anonymous access is still enabled")
		}

		return nil
	}
}
