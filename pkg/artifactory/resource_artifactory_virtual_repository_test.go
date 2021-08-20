package artifactory

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccVirtualRepository_basic(t *testing.T) {
	id := rand.Int()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryBasic = `
		resource "artifactory_virtual_repository" "%s" {
			key          = "%s"
			package_type = "maven"
			repositories = []
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVirtualRepositoryDestroy(fqrn),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryBasic, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
				),
			},
		},
	})
}

func TestAccVirtualRepository_update(t *testing.T) {
	id := rand.Int()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryUpdateBefore = `
		resource "artifactory_virtual_repository" "%s" {
			key          = "%s"
			description  = "Before"
			package_type = "maven"
			repositories = []
		}
	`
	const virtualRepositoryUpdateAfter = `
		resource "artifactory_virtual_repository" "%s" {
			key          = "%s"
			description  = "After"
			package_type = "maven"
			repositories = []
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVirtualRepositoryDestroy(fqrn),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryUpdateBefore, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "description", "Before"),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(virtualRepositoryUpdateAfter, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "description", "After"),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
				),
			},
		},
	})
}

func TestNugetPackageCreationFull(t *testing.T) {
	id := rand.Int()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryFull = `
		resource "artifactory_virtual_repository" "%s" {
			key = "%s"
			package_type = "nuget"
			repo_layout_ref = "nuget-default"
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVirtualRepositoryDestroy(fqrn),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryFull, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "nuget"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "nuget-default"),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", "true"),
				),
			},
		},
	})

}
func TestAccVirtualRepository_full(t *testing.T) {
	id := rand.Int()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryFull = `
		resource "artifactory_virtual_repository" "%s" {
			key = "%s"
			package_type = "maven"
			repo_layout_ref = "maven-1-default"
			repositories = []
			description = "A test virtual repo"
			notes = "Internal description"
			includes_pattern = "com/atlassian/**,cloud/atlassian/**"
			excludes_pattern = "com/google/**"
			artifactory_requests_can_retrieve_remote_artifacts = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVirtualRepositoryDestroy(fqrn),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryFull, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "maven-1-default"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "description", "A test virtual repo"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal description"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "com/atlassian/**,cloud/atlassian/**"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "com/google/**"),
					resource.TestCheckResourceAttr(fqrn, "pom_repository_references_cleanup_policy", "discard_active_reference"),
				),
			},
		},
	})
}

func testAccCheckVirtualRepositoryDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ArtClient).Resty

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("error: Resource id [%s] not found", id)
		}

		resp, err := client.R().Head("artifactory/api/repositories/" + rs.Primary.ID)

		if err != nil {

			if resp != nil && (resp.StatusCode() == http.StatusNotFound || resp.StatusCode() == http.StatusBadRequest) {
				return nil
			}
			return err
		}

		return fmt.Errorf("error: Repository %s still exists", rs.Primary.ID)
	}
}
