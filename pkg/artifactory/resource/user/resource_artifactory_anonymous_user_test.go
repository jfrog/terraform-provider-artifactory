package user_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v8/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v8/pkg/artifactory/provider"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccAnonymousUser_Importable(t *testing.T) {
	const anonymousUserConfig = `
		resource "artifactory_anonymous_user" "anonymous" {
		}
	`

	fqrn := "artifactory_anonymous_user.anonymous"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		Steps: []resource.TestStep{
			{
				Config:           anonymousUserConfig,
				ResourceName:     fqrn,
				ImportState:      true,
				ImportStateId:    "anonymous",
				ImportStateCheck: validator.CheckImportState("anonymous", "id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", "anonymous"),
				),
			},
		},
	})
}

func TestAccAnonymousUser_NotCreatable(t *testing.T) {

	const anonymousUserConfig = `
		resource "artifactory_anonymous_user" "anonymous" {
			name = "anonymous"
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
		},
		Steps: []resource.TestStep{
			{
				Config:      anonymousUserConfig,
				ExpectError: regexp.MustCompile(".*Anonymous Artifactory user cannot be created.*"),
			},
		},
	})
}
