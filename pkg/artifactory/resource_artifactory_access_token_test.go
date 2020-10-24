package artifactory

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const accessToken = `
resource "artifactory_access_token" "test_access_token" {
	username = "test"
	scope    = "api:*"
  }
`

func TestAccAccessToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckAccessTokenDestroy("artifactory_access_token.test_access_token"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: accessToken,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("artifactory_access_token.test_access_token", "access_token"),
					resource.TestCheckResourceAttrSet("artifactory_access_token.test_access_token", "refresh_token"),
				),
			},
		},
	})
}

func testAccCheckAccessTokenDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		return nil
	}
}
