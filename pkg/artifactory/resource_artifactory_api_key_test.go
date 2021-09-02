package artifactory

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)



func TestAccApiKey(t *testing.T) {
	const apiKey = `
		resource "artifactory_api_key" "foobar" {}
	`

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
		client := apis.Resty
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}
		data := make(map[string]string)

		_, err := client.R().SetResult(&data).Get(apiKeyEndpoint)

		if err != nil {
			return err
		}
		if _,ok = data["apiKey"]; ok {
			return fmt.Errorf("error: API key %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
