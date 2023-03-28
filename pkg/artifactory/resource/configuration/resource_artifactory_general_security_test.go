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

const GeneralSecurityTemplateFull = `
resource "artifactory_general_security" "security" {
	enable_anonymous_access = true
}`

func TestAccGeneralSecurity_full(t *testing.T) {
	fqrn := "artifactory_general_security.security"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccGeneralSecurityDestroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: GeneralSecurityTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enable_anonymous_access", "true"),
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

func testAccGeneralSecurityDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProvderMetadata).Client

		_, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		generalSettings := configuration.GeneralSettings{}
		_, err := client.R().SetResult(&generalSettings).Get("artifactory/api/securityconfig")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from <base_url>/artifactory/api/securityconfig during Read")
		}
		if generalSettings.AnonAccessEnabled != false {
			return fmt.Errorf("error: general security setting to allow anonymous access is still enabled")
		}

		return nil
	}
}
