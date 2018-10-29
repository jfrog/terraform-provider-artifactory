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

const remoteRepository_basic = `
resource "artifactory_remote_repository" "basic" {
	key             = "tf-virtual-basic"
    package_type    = "npm"
	url             = "https://registry.npmjs.org/"
	repo_layout_ref = "npm-default"
}`

func TestAccRemoteRepository_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceRemoteRepositoryCheckDestroy("artifactory_remote_repository.basic"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: remoteRepository_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_remote_repository.basic", "key", "tf-virtual-basic"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.basic", "package_type", "npm"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.basic", "url", "https://registry.npmjs.org/"),
				),
			},
		},
	})
}

const remoteRepositoryConfig_full = `
resource "artifactory_remote_repository" "full" {
    key                             	  = "tf-virtual-full"
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
		CheckDestroy: resourceRemoteRepositoryCheckDestroy("artifactory_remote_repository.full"),
		Steps: []resource.TestStep{
			{
				Config: remoteRepositoryConfig_full,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "key", "tf-virtual-full"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "package_type", "npm"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "url", "https://registry.npmjs.org/"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "username", "user"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "password", GetMD5Hash("pass")),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "proxy", ""),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "description", "desc (local file cache)"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "notes", "notes"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "includes_pattern", "**/*.js"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "excludes_pattern", "**/*.jsx"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "repo_layout_ref", "npm-default"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "handle_releases", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "handle_snapshots", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "max_unique_snapshots", "15"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "suppress_pom_consistency_checks", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "hard_fail", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "offline", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "blacked_out", "false"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "store_artifacts_locally", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "socket_timeout_millis", "25000"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "local_address", ""),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "retrieval_cache_period_seconds", "15"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "missed_cache_period_seconds", "2500"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "unused_artifacts_cleanup_period_hours", "96"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "fetch_jars_eagerly", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "fetch_sources_eagerly", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "share_configuration", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "synchronize_properties", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "block_mismatching_mime_types", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "property_sets.#", "1"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "property_sets.214975871", "artifactory"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "allow_any_host_auth", "false"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "enable_cookie_management", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.full", "client_tls_certificate", ""),
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
			return fmt.Errorf("Not found %s", id)
		}

		// It seems artifactory just can't keep up with high requests
		time.Sleep(time.Duration(1 * time.Second))
		_, resp, err := client.Repositories.GetRemote(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("Repository %s still exists", rs.Primary.ID)
		}
	}
}
