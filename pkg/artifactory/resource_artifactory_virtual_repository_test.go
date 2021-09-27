package artifactory

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVirtualRepository_basic(t *testing.T) {
	id := randomInt()
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
		CheckDestroy: testAccCheckRepositoryDestroy(fqrn),
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
	id := randomInt()
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
		CheckDestroy: testAccCheckRepositoryDestroy(fqrn),
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
func TestAllPackageTypes(t *testing.T) {
	for _, repo := range repoTypesSupported {
		if repo != "nuget" { // this requires special testing
			t.Run(fmt.Sprintf("TestVirtual%sRepo", strings.Title(strings.ToLower(repo))), func(t *testing.T) {
				// NuGet Repository configuration is missing mandatory field downloadContextPath
				resource.Test(mkVirtualTestCase(repo, t))
			})
		}
	}
}

func mkVirtualTestCase(repo string, t *testing.T) (*testing.T, resource.TestCase) {
	id := randomInt()
	name := fmt.Sprintf("%s%d", repo, id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryFull = `
		resource "artifactory_virtual_repository" "%s" {
			key = "%s"
			package_type = "%s"
			repo_layout_ref = "maven-1-default"
			repositories = []
			description = "A test virtual repo"
			notes = "Internal description"
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			artifactory_requests_can_retrieve_remote_artifacts = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
		}
	`
	return t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckRepositoryDestroy(fqrn),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryFull, name, name, repo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", repo),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "maven-1-default"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "description", "A test virtual repo"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal description"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "com/jfrog/**,cloud/jfrog/**"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "com/google/**"),
					resource.TestCheckResourceAttr(fqrn, "pom_repository_references_cleanup_policy", "discard_active_reference"),
				),
			},
		},
	}
}

func TestNugetPackageCreationFull(t *testing.T) {
	id := randomInt()
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
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			artifactory_requests_can_retrieve_remote_artifacts = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
			force_nuget_authentication	= true
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckRepositoryDestroy(fqrn),
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
	id := randomInt()
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
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			artifactory_requests_can_retrieve_remote_artifacts = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckRepositoryDestroy(fqrn),
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
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "com/jfrog/**,cloud/jfrog/**"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "com/google/**"),
					resource.TestCheckResourceAttr(fqrn, "pom_repository_references_cleanup_policy", "discard_active_reference"),
				),
			},
		},
	})
}

func testAccCheckRepositoryDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("error: Resource id [%s] not found", id)
		}
		exists, _ := repoExists(rs.Primary.ID, testAccProvider.Meta())
		if exists {
			return fmt.Errorf("error: Repository %s still exists", rs.Primary.ID)
		}
		return nil
	}
}


func TestAccVirtualRepository_basic_with_project_key(t *testing.T) {
	id := randomInt()
	name := fmt.Sprintf("foo-with-project-key%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryBasic = `
		resource "artifactory_virtual_repository" "%s" {
			key          = "%s"
			package_type = "maven"
			repositories = []
			project_key = "frog-proj"
			environments = [ "DEV", "PROD" ]
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckRepositoryDestroy(fqrn),
		Providers:    testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryBasic, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "project_key", "frog-proj"),
					resource.TestCheckResourceAttr(fqrn, "environments.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "environments.0", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "environments.1", "PROD"),
				),
			},
		},
	})
}