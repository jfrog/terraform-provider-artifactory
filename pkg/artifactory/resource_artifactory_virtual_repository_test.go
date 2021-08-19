package artifactory

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const virtualRepositoryBasic = `
resource "artifactory_virtual_repository" "foo" {
	key          = "foo"
	package_type = "maven"
	repositories = []
}
`

func TestAccVirtualRepository_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVirtualRepositoryDestroy("artifactory_virtual_repository.foo"),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "key", "foo"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "package_type", "maven"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "repositories.#", "0"),
				),
			},
		},
	})
}

const virtualRepositoryUpdateBefore = `
resource "artifactory_virtual_repository" "foo" {
	key          = "foo"
	description  = "Before"
	package_type = "maven"
	repositories = []
}
`

const virtualRepositoryUpdateAfter = `
resource "artifactory_virtual_repository" "foo" {
	key          = "foo"
	description  = "After"
	package_type = "maven"
	repositories = []
}
`

func TestAccVirtualRepository_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVirtualRepositoryDestroy("artifactory_virtual_repository.foo"),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryUpdateBefore,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "key", "foo"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "description", "Before"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "package_type", "maven"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "repositories.#", "0"),
				),
			},
			{
				Config: virtualRepositoryUpdateAfter,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "key", "foo"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "description", "After"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "package_type", "maven"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "repositories.#", "0"),
				),
			},
		},
	})
}

const virtualRepositoryFull = `
resource "artifactory_virtual_repository" "foo" {
	key = "foo"
	package_type = "maven"
	repo_layout_ref = "maven-1-default"
	repositories = []
	description = "A test virtual repo"
	notes = "Internal description"
	includes_pattern = "com/atlassian/**,cloud/atlassian/**"
    excludes_pattern = "com/google/**"
	artifactory_requests_can_retrieve_remote_artifacts = true
	pom_repository_references_cleanup_policy = "discard_active_reference"
	force_nuget_authentication	= true
}
`

func TestAccVirtualRepository_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVirtualRepositoryDestroy("artifactory_virtual_repository.foo"),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "key", "foo"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "package_type", "maven"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "repo_layout_ref", "maven-1-default"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "repositories.#", "0"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "description", "A test virtual repo"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "notes", "Internal description"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "includes_pattern", "com/atlassian/**,cloud/atlassian/**"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "excludes_pattern", "com/google/**"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "pom_repository_references_cleanup_policy", "discard_active_reference"),
					resource.TestCheckResourceAttr("artifactory_virtual_repository.foo", "force_nuget_authentication", "true"),
				),
			},
		},
	})
}

func testAccCheckVirtualRepositoryDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.ArtOld

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("error: Resource id [%s] not found", id)
		}

		repo, resp, err := client.V1.Repositories.GetVirtual(context.Background(), rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
			return nil
		}
		return fmt.Errorf("error: Repository %s still exists %s", rs.Primary.ID, repo)
	}
}
