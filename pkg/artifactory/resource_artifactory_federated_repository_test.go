package artifactory

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const federatedRepositoryBasic = `
resource "artifactory_federated_repository" "terraform-federated-test-repo-basic" {
	key 	     = "terraform-federated-test-repo-basic"
	package_type = "docker"
}`

func TestAccFederatedRepository_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceFederatedRepositoryCheckDestroy("artifactory_federated_repository.terraform-federated-test-repo-basic"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_federated_repository.terraform-federated-test-repo-basic", "key", "terraform-federated-test-repo-basic"),
					resource.TestCheckResourceAttr("artifactory_federated_repository.terraform-federated-test-repo-basic", "package_type", "docker"),
				),
			},
		},
	})
}

func resourceFederatedRepositoryCheckDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.ArtOld
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		_, resp, err := client.V1.Repositories.GetFederated(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Federated repo %s still exists", rs.Primary.ID)
		}
	}
}
