package artifactory

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

const remoteRepoNuget = `
resource "artifactory_remote_repository" "terraform-remote-test-repo-nuget" {
	key               		   = "terraform-remote-test-repo-nuget"
	url               		   = "https://www.nuget.org/"
	repo_layout_ref   		   = "nuget-default"
    package_type      		   = "nuget"
	feed_context_path 		   = "/api/notdefault"
	force_nuget_authentication = true
}`

func TestAccRemoteRepository_nugetNew(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceRemoteRepositoryCheckDestroy("artifactory_remote_repository.terraform-remote-test-repo-nuget"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: remoteRepoNuget,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-nuget", "key", "terraform-remote-test-repo-nuget"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-nuget", "v3_feed_url", ""),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-nuget", "feed_context_path", "/api/notdefault"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-nuget", "force_nuget_authentication", "true"),
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
	remote_repo_checksum_policy_type      = "ignore-and-generate"
	client_tls_certificate				  = ""
	force_nuget_authentication 			  = true
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
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "password", getMD5Hash("pass")),
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
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "remote_repo_checksum_policy_type", "ignore-and-generate"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "force_nuget_authentication", "true"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-full", "propagate_query_params", "true"),
				),
			},
		},
	})
}

func resourceRemoteRepositoryCheckDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.ArtOld

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("not found %s", id)
		}

		_, resp, err := client.V1.Repositories.GetRemote(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("repository %s still exists", rs.Primary.ID)
		}
	}
}

const remoteNpmRepoBasicWithPropagate = `
resource "artifactory_remote_repository" "terraform-remote-test-repo-basic" {
	key                     = "terraform-remote-test-repo-basic"
        package_type            = "npm"
	url                     = "https://registry.npmjs.org/"
	repo_layout_ref         = "npm-default"
	propagate_query_params  = true
}`

func TestAccRemoteRepository_npm_with_propagate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      remoteNpmRepoBasicWithPropagate,
				ExpectError: regexp.MustCompile(`Cannot use propagate_query_params with repository type npm. This parameter can be used only with generic repositories.`),
			},
		},
	})
}

const remoteGenericRepoBasicWithPropagate = `
resource "artifactory_remote_repository" "terraform-remote-test-repo-basic" {
	key                     = "terraform-remote-test-repo-basic"
        package_type            = "generic"
	url                     = "https://registry.npmjs.org/"
	repo_layout_ref         = "simple-default"
	propagate_query_params  = true
}`

func TestAccRemoteRepository_generic_with_propagate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceRemoteRepositoryCheckDestroy("artifactory_remote_repository.terraform-remote-test-repo-basic"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: remoteGenericRepoBasicWithPropagate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-basic", "key", "terraform-remote-test-repo-basic"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-basic", "package_type", "generic"),
					resource.TestCheckResourceAttr("artifactory_remote_repository.terraform-remote-test-repo-basic", "propagate_query_params", "true"),
				),
			},
		},
	})
}
