package artifactory

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

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

func TestAccVirtualRepository_reset_default_deployment_repo(t *testing.T) {
	id := randomInt()
	name := fmt.Sprintf("foo%d", id)
	localRepoName := fmt.Sprintf("%s-local", name)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryWithDefaultDeploymentRepo = `
		resource "artifactory_local_repository" "%[1]s" {
			key = "%[1]s"
			package_type = "maven"
		}

		resource "artifactory_virtual_repository" "%[2]s" {
			key          = "%[2]s"
			package_type = "maven"
			repositories = ["%[1]s"]
			default_deployment_repo = "%[1]s"
			depends_on = [artifactory_local_repository.%[1]s]
		}
	`
	const virtualRepositoryWithoutDefaultDeploymentRepo = `
		resource "artifactory_local_repository" "%[1]s" {
			key = "%[1]s"
			package_type = "maven"
		}

		resource "artifactory_virtual_repository" "%[2]s" {
			key          = "%[2]s"
			package_type = "maven"
			repositories = ["%[1]s"]
			depends_on = [artifactory_local_repository.%[1]s]
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryWithoutDefaultDeploymentRepo, localRepoName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "default_deployment_repo", ""),
				),
			},
			{
				Config: fmt.Sprintf(virtualRepositoryWithDefaultDeploymentRepo, localRepoName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "default_deployment_repo", localRepoName),
				),
			},
			{
				Config: fmt.Sprintf(virtualRepositoryWithoutDefaultDeploymentRepo, localRepoName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "default_deployment_repo", ""),
				),
			},
		},
	})
}

func TestAccVirtualGoRepository_basic(t *testing.T) {
	_, fqrn, name := mkNames("foo", "artifactory_virtual_go_repository")
	var virtualRepositoryBasic = fmt.Sprintf(`
		resource "artifactory_virtual_go_repository" "%s" {
		  key          = "%s"
		  repo_layout_ref = "go-default"
		  repositories = []
		  description = "A test virtual repo"
		  notes = "Internal description"
		  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
		  excludes_pattern = "com/google/**"
		  external_dependencies_enabled = true
		  external_dependencies_patterns = [
			"**/github.com/**",
			"**/go.googlesource.com/**"
		  ]
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				// we check to make sure some of the base params are picked up
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "go"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "**/github.com/**"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.1", "**/go.googlesource.com/**"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "2"),
				),
			},
		},
	})
}

func TestAccVirtualMavenRepository_basic(t *testing.T) {
	id := randomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_maven_repository.%s", name)
	var virtualRepositoryBasic = fmt.Sprintf(`
		resource "artifactory_virtual_maven_repository" "%s" {
			key          = "%s"
			repo_layout_ref = "maven-2-default"
			repositories = []
			description = "A test virtual repo"
			notes = "Internal description"
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			force_maven_authentication = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				// we check to make sure some of the base params are picked up
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "force_maven_authentication", "true"),
					// to test key pair, we'd have to be able to create them on the fly and we currently can't.
					resource.TestCheckResourceAttr(fqrn, "key_pair", ""),
					resource.TestCheckResourceAttr(fqrn, "pom_repository_references_cleanup_policy", "discard_active_reference"),
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
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

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
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

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
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

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
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

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
