package artifactory_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
)

func TestAccAnonymousUserImportable(t *testing.T) {
	const anonymousUserConfig = `
		resource "artifactory_anonymous_user" "anonymous" {
		}
	`

	fqrn := "artifactory_anonymous_user.anonymous"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:            anonymousUserConfig,
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateId:     "anonymous",
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) == 0 {
						return fmt.Errorf("No import state")
					}

					instanceState := states[0]
					if instanceState.ID != "anonymous" {
						return fmt.Errorf("Incorrect state ID: %s", instanceState.ID)
					}

					if instanceState.Attributes["name"] != "anonymous" {
						return fmt.Errorf("Incorrect state attribute 'name': %s", instanceState.Attributes["name"])
					}

					return nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", "anonymous"),
				),
			},
		},
	})
}

func TestAccAnonymousUserNotCreateable(t *testing.T) {

	const anonymousUserConfig = `
		resource "artifactory_anonymous_user" "anonymous" {
			name = "anonymous"
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      anonymousUserConfig,
				ExpectError: regexp.MustCompile(".*Anonymous Artifactory user cannot be created.*"),
			},
		},
	})
}
