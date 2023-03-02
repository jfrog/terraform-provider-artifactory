package user_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/validator"
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
				Config:           anonymousUserConfig,
				ResourceName:     fqrn,
				ImportState:      true,
				ImportStateId:    "anonymous",
				ImportStateCheck: validator.CheckImportState("anonymous", "name"),
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
