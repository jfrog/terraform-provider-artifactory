package artifactory

import (
	"testing"
	"time"

	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"net/http"
)

const localRepository_basic = `
resource "artifactory_local_repository" "basic" {
	key 	     = "tf-local-basic"
	package_type = "docker"
}`

func TestAccLocalRepository_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceLocalRepositoryCheckDestroy("artifactory_local_repository.basic"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepository_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_local_repository.basic", "key", "tf-local-basic"),
					resource.TestCheckResourceAttr("artifactory_local_repository.basic", "package_type", "docker"),
				),
			},
		},
	})
}

const localRepositoryConfig_full = `
resource "artifactory_local_repository" "full" {
    key                             = "tf-local-full"
    package_type                    = "npm"
	description                     = "Test repo for terraform-provider-artifactory"
	notes                           = "Test repo for terraform-provider-artifactory"
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
}`

func TestAccLocalRepository_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceLocalRepositoryCheckDestroy("artifactory_local_repository.full"),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryConfig_full,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "key", "tf-local-full"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "package_type", "npm"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "description", "Test repo for terraform-provider-artifactory"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "notes", "Test repo for terraform-provider-artifactory"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "excludes_pattern", "**/*.tgz"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "repo_layout_ref", "npm-default"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "handle_releases", "true"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "handle_snapshots", "true"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "max_unique_snapshots", "25"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "debian_trivial_layout", "false"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "checksum_policy_type", "client-checksums"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "max_unique_tags", "100"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "snapshot_version_behavior", "unique"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "suppress_pom_consistency_checks", "true"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "blacked_out", "false"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "property_sets.#", "1"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "property_sets.214975871", "artifactory"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "archive_browsing_enabled", "false"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "calculate_yum_metadata", "false"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "yum_root_depth", "0"),
					resource.TestCheckResourceAttr("artifactory_local_repository.full", "docker_api_version", "V2"),
				),
			},
		},
	})
}

func resourceLocalRepositoryCheckDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*artifactory.Client)
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		// It seems artifactory just can't keep up with high requests
		time.Sleep(time.Duration(1 * time.Second))
		_, resp, err := client.Repositories.GetLocal(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Local repo %s still exists", rs.Primary.ID)
		}
	}
}
