package artifactory_test

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory"
)

func TestAccApiKey(t *testing.T) {
	const apiKey = `
		resource "artifactory_api_key" "foobar" {}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApiKeyDestroy("artifactory_api_key.foobar"),
		Steps: []resource.TestStep{
			{
				Config: apiKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("artifactory_api_key.foobar", "api_key"),
				),
			},
		},
	})
}

func testAccCheckApiKeyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(*resty.Client)
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		data := make(map[string]string)

		_, err := client.R().SetResult(&data).Get(artifactory.ApiKeyEndpoint)
		if err != nil {
			return err
		}

		if _, ok = data["apiKey"]; ok {
			return fmt.Errorf("error: API key %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
