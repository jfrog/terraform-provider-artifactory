package artifactory

import (
	"testing"

	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"net/http"
)

const remoteRepoBasic = `
resource "artifactory_remote_repository" "terraform-remote-test-repo-basic" {
	key = "terraform-remote-test-repo-basic"
    package_type                          = "npm"
	url                                   = "https://registry.npmjs.org/"
	repo_layout_ref                       = "npm-default"
}`

func TestAccRemoteRepository_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceRemoteRepositoryCheckDestroy("artifactory_remote_repository.terraform-remote-test-repo-basic"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: remoteRepoBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-basic", "key", "terraform-remote-test-repo-basic"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-basic", "package_type", "npm"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-basic", "url", "https://registry.npmjs.org/"),
				),
			},
		},
	})
}

const remoteRepoFull = `
resource "artifactory_remote_repository" "terraform-remote-test-repo-full" {
    key                             	  = "terraform-remote-test-repo-full"
	package_type                          = "npm"
	url                                   = "https://registry.npmjs.org/"
	username                              = "user"
	password                              = "pass"
    proxy                                 = ""
	description                           = "desc"
	notes                                 = "notes"
	includes_pattern                      = "**/*.js"
	excludes_pattern                      = "**/*.jsx"
	repo_layout_ref                       = "npm-default"
	handle_releases                       = true
	handle_snapshots                      = true
	max_unique_snapshots                  = 15
	suppress_pom_consistency_checks       = true
	hard_fail                             = true
	offline                               = true
	blacked_out                           = false
	store_artifacts_locally               = true
	socket_timeout_millis                 = 25000
	local_address                         = ""
	retrieval_cache_period_seconds        = 15
	missed_cache_period_seconds           = 2500
	unused_artifacts_cleanup_period_hours = 96
	fetch_jars_eagerly                    = true
    fetch_sources_eagerly                 = true
	share_configuration                   = true
	synchronize_properties                = true
	block_mismatching_mime_types		  = true
	property_sets                         = ["artifactory"]
	allow_any_host_auth                   = false
	enable_cookie_management              = true

	client_tls_certificate				  = ""
}`

func TestAccRemoteRepository_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceRemoteRepositoryCheckDestroy("artifactory_remote_repository.terraform-remote-test-repo-full"),
		Steps: []resource.TestStep{
			{
				Config: remoteRepoFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "key", "terraform-remote-test-repo-full"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "package_type", "npm"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "url", "https://registry.npmjs.org/"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "username", "user"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "password", "pass"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "proxy", ""),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "description", "desc (local file cache)"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "notes", "notes"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "includes_pattern", "**/*.js"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "excludes_pattern", "**/*.jsx"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "repo_layout_ref", "npm-default"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "handle_releases", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "handle_snapshots", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "max_unique_snapshots", "15"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "suppress_pom_consistency_checks", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "hard_fail", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "offline", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "blacked_out", "false"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "store_artifacts_locally", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "socket_timeout_millis", "25000"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "local_address", ""),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "retrieval_cache_period_seconds", "15"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "missed_cache_period_seconds", "2500"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "unused_artifacts_cleanup_period_hours", "96"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "fetch_jars_eagerly", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "fetch_sources_eagerly", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "share_configuration", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "synchronize_properties", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "block_mismatching_mime_types", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "property_sets.#", "1"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "property_sets.214975871", "artifactory"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "allow_any_host_auth", "false"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "enable_cookie_management", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "client_tls_certificate", ""),
				),
			},
		},
	})
}

func resourceRemoteRepositoryCheckDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*artifactory.Client)
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("not found %s", id)
		}

		_, resp, err := client.Repositories.GetRemote(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("repository %s still exists", rs.Primary.ID)
		}
	}
}
