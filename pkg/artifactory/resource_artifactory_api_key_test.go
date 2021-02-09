package artifactory

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const apiKey = `
resource "artifactory_api_key" "foobar" {}
`

func TestAccApiKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckApiKeyDestroy("artifactory_api_key.foobar"),
		Providers:    testAccProviders,
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
		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.ArtOld
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		key, _, err := client.V1.Security.GetApiKey(context.Background())

		if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else if key.ApiKey != nil {
			return fmt.Errorf("error: API key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
