package configuration_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
)

const BaseUrlTemplate = `
resource "artifactory_configuration_base_url" "baseUrl" {
  base_url        = "http://fake-url"
}`

func TestAccBase_Url_Configuration(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	fqrn := "artifactory_configuration_base_url.baseUrl"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: BaseUrlTemplate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "base_url", jfrogURL),
				),
			},
		},
	})
}
