package artifactory

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
					resource.TestCheckResourceAttrSet("artifactory_api_key", "api_key"),
				),
			},
		},
	})
}

func testAccCheckApiKeyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*artifactory.Artifactory)
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		_, resp, err := client.V1.Security.GetApiKey(context.Background())

		if resp.StatusCode == http.StatusNotFound {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: API key %s still exists", rs.Primary.ID)
		}
	}
}
