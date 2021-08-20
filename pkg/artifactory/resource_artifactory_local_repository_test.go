package artifactory

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccLocalRepository_basic(t *testing.T) {
	name := fmt.Sprintf("terraform-local-test-repo-basic%d", rand.Int())
	resourceName := fmt.Sprintf("artifactory_local_repository.%s", name)
	localRepositoryBasic := fmt.Sprintf(`
		resource "artifactory_local_repository" "%s" {
			key 	     = "%s"
			package_type = "docker"
		}
	`, name, name) // we use randomness so that, in the case of failure and dangle, the next test can run without collision
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceLocalRepositoryCheckDestroy(resourceName),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", "docker"),
				),
			},
		},
	})
}

func mkTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("terraform-local-test-%d-full", rand.Int())
	resourceName := fmt.Sprintf("artifactory_local_repository.%s", name)
	const localRepositoryConfigFull = `
		resource "artifactory_local_repository" "%s" {
			key                             = "%s"
			package_type                    = "%s"
			description                     = "Test repo for %s"
			notes                           = "Test repo for %s"
			includes_pattern                = "**/*"
			excludes_pattern                = "**/*.tgz"
			repo_layout_ref                 = "npm-default"
			handle_releases                 = true
			handle_snapshots                = true
			max_unique_snapshots            = 25
			debian_trivial_layout           = false
			checksum_policy_type            = "client-checksums"
			max_unique_tags                 = 100
			snapshot_version_behavior       = "unique"
			suppress_pom_consistency_checks = true
			blacked_out                     = false
			property_sets                   = [ "artifactory" ]
			archive_browsing_enabled        = false
			calculate_yum_metadata          = false
			yum_root_depth                  = 0
			docker_api_version              = "V2"
		}
	`

	cfg := fmt.Sprintf(localRepositoryConfigFull, name, name, repoType, name, name)
	return t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceLocalRepositoryCheckDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", repoType),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "notes", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(resourceName, "excludes_pattern", "**/*.tgz"),
					resource.TestCheckResourceAttr(resourceName, "repo_layout_ref", "npm-default"),
					resource.TestCheckResourceAttr(resourceName, "handle_releases", "true"),
					resource.TestCheckResourceAttr(resourceName, "handle_snapshots", "true"),
					resource.TestCheckResourceAttr(resourceName, "max_unique_snapshots", "25"),
					resource.TestCheckResourceAttr(resourceName, "debian_trivial_layout", "false"),
					resource.TestCheckResourceAttr(resourceName, "checksum_policy_type", "client-checksums"),
					resource.TestCheckResourceAttr(resourceName, "max_unique_tags", "100"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_version_behavior", "unique"),
					resource.TestCheckResourceAttr(resourceName, "suppress_pom_consistency_checks", "true"),
					resource.TestCheckResourceAttr(resourceName, "blacked_out", "false"),
					resource.TestCheckResourceAttr(resourceName, "property_sets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "property_sets.0", "artifactory"),
					resource.TestCheckResourceAttr(resourceName, "archive_browsing_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "calculate_yum_metadata", "false"),
					resource.TestCheckResourceAttr(resourceName, "yum_root_depth", "0"),
					resource.TestCheckResourceAttr(resourceName, "docker_api_version", "V2"),
				),
			},
		},
	}
}

func TestAccAllRepoTypesLocal(t *testing.T) {

	for _, repo := range repoTypesSupported {
		resource.Test(mkTestCase(repo, t))
	}
}

func resourceLocalRepositoryCheckDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ArtClient).Resty
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}
		resp, err := client.R().SetHeader("accept", "*/*").
			Delete("/artifactory/api/repositories/" + rs.Primary.ID)
		if err != nil && resp != nil && (resp.StatusCode() == http.StatusNotFound || resp.StatusCode() == http.StatusBadRequest) {
			return nil
		}
		return fmt.Errorf("local should not exist repo err %s: %d", rs.Primary.ID, resp.StatusCode())
	}
}
