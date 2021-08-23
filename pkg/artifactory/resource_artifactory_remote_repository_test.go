package artifactory

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestKeyHasSpecialCharsFails(t *testing.T) {
	const failKey = `
		resource "artifactory_remote_repository" "terraform-remote-test-repo-basic" {
			key                     = "IHave++special_Chars"
			package_type            = "npm"
			url                     = "https://registry.npmjs.org/"
			repo_layout_ref         = "npm-default"
			propagate_query_params  = true
			retrieval_cache_period_seconds        = 70
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      failKey,
				ExpectError: regexp.MustCompile(".*expected value of key to not contain any of.*"),
			},
		},
	})
}

func TestAccRemoteRepository_basic(t *testing.T) {
	id := rand.Int()
	name := fmt.Sprintf("terraform-remote-test-repo-basic%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_repository.%s", name)
	const remoteRepoBasic = `
		resource "artifactory_remote_repository" "%s" {
			key 				  = "%s"
			package_type          = "npm"
			url                   = "https://registry.npmjs.org/"
			repo_layout_ref       = "npm-default"
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceRemoteRepositoryCheckDestroy(fqrn),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(remoteRepoBasic, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "npm"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://registry.npmjs.org/"),
				),
			},
		},
	})
}

func TestAccRemoteRepository_nugetNew(t *testing.T) {
	const remoteRepoNuget = `
		resource "artifactory_remote_repository" "%s" {
			key               		   = "%s"
			url               		   = "https://www.nuget.org/"
			repo_layout_ref   		   = "nuget-default"
			package_type      		   = "nuget"
			download_context_path	   = "Download"
			feed_context_path 		   = "/api/notdefault"
			force_nuget_authentication = true
		}
	`
	id := rand.Int()
	name := fmt.Sprintf("terraform-remote-test-repo-nuget%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_repository.%s", name)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceRemoteRepositoryCheckDestroy(fqrn),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(remoteRepoNuget, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "v3_feed_url", ""),
					resource.TestCheckResourceAttr(fqrn, "feed_context_path", "/api/notdefault"),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", "true"),
				),
			},
		},
	})
}

func TestAllRemoteRepoTypes(t *testing.T) {
	//
	for _, repo := range repoTypesSupported {
		if repo != "nuget" { // this requires special testing
			t.Run(fmt.Sprintf("TestVirtual%sRepo", strings.Title(strings.ToLower(repo))), func(t *testing.T) {
				// NuGet Repository configuration is missing mandatory field downloadContextPath
				resource.Test(mkRemoteRepoTestCase(repo, t))
			})
		}
	}
}

func mkRemoteRepoTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	const remoteRepoFull = `
		resource "artifactory_remote_repository" "%s" {
			key                             	  = "%s"
			package_type                          = "%s"
			url                                   = "https://registry.npmjs.org/"
			username                              = "user"
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
			retrieval_cache_period_seconds        = 70
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
		}
	`
	id := rand.Int()
	name := fmt.Sprintf("terraform-remote-test-repo-full%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_repository.%s", name)
	return t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceRemoteRepositoryCheckDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(remoteRepoFull, name, name, repoType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", repoType),
					resource.TestCheckResourceAttr(fqrn, "url", "https://registry.npmjs.org/"),
					resource.TestCheckResourceAttr(fqrn, "username", "user"),
					resource.TestCheckResourceAttr(fqrn, "proxy", ""),
					resource.TestCheckResourceAttr(fqrn, "description", "desc (local file cache)"),
					resource.TestCheckResourceAttr(fqrn, "notes", "notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*.js"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "**/*.jsx"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "npm-default"),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", "true"),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", "true"),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", "15"),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", "true"),
					resource.TestCheckResourceAttr(fqrn, "hard_fail", "true"),
					resource.TestCheckResourceAttr(fqrn, "offline", "true"),
					resource.TestCheckResourceAttr(fqrn, "blacked_out", "false"),
					resource.TestCheckResourceAttr(fqrn, "store_artifacts_locally", "true"),
					resource.TestCheckResourceAttr(fqrn, "socket_timeout_millis", "25000"),
					resource.TestCheckResourceAttr(fqrn, "local_address", ""),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "70"),
					resource.TestCheckResourceAttr(fqrn, "missed_cache_period_seconds", "2500"),
					resource.TestCheckResourceAttr(fqrn, "unused_artifacts_cleanup_period_hours", "96"),
					resource.TestCheckResourceAttr(fqrn, "fetch_jars_eagerly", "true"),
					resource.TestCheckResourceAttr(fqrn, "fetch_sources_eagerly", "true"),
					resource.TestCheckResourceAttr(fqrn, "share_configuration", "true"),
					resource.TestCheckResourceAttr(fqrn, "synchronize_properties", "true"),
					resource.TestCheckResourceAttr(fqrn, "block_mismatching_mime_types", "true"),
					resource.TestCheckResourceAttr(fqrn, "property_sets.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "property_sets.214975871", "artifactory"),
					resource.TestCheckResourceAttr(fqrn, "allow_any_host_auth", "false"),
					resource.TestCheckResourceAttr(fqrn, "enable_cookie_management", "true"),
					resource.TestCheckResourceAttr(fqrn, "client_tls_certificate", ""),
					resource.TestCheckResourceAttr(fqrn, "remote_repo_checksum_policy_type", "ignore-and-generate"),
				),
			},
		},
	}
}

func resourceRemoteRepositoryCheckDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ArtClient).Resty
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("not found %s", id)
		}
		resp, err := client.R().Head(repositoriesEndpoint+ rs.Primary.ID)

		if err != nil {
			if resp != nil && (resp.StatusCode() == http.StatusNotFound || resp.StatusCode() == http.StatusBadRequest) {
				return nil
			}
			return err
		}
		return nil
	}
}

func TestAccRemoteRepository_npm_with_propagate(t *testing.T) {
	const remoteNpmRepoBasicWithPropagate = `
		resource "artifactory_remote_repository" "terraform-remote-test-repo-basic" {
			key                     = "terraform-remote-test-repo-basic"
			package_type            = "npm"
			url                     = "https://registry.npmjs.org/"
			repo_layout_ref         = "npm-default"
			propagate_query_params  = true
			retrieval_cache_period_seconds        = 70
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      remoteNpmRepoBasicWithPropagate,
				ExpectError: regexp.MustCompile(".*cannot use propagate_query_params with repository type npm.*"),
			},
		},
	})
}

func TestAccRemoteRepository_generic_with_propagate(t *testing.T) {

	const remoteGenericRepoBasicWithPropagate = `
		resource "artifactory_remote_repository" "%s" {
			key                     = "%s"
			description = "This is a test"
			package_type            = "generic"
			url                     = "https://registry.npmjs.org/"
			repo_layout_ref         = "simple-default"
			propagate_query_params  = true
			retrieval_cache_period_seconds        = 70

		}
	`
	id := rand.Int()
	name := fmt.Sprintf("terraform-remote-test-repo-basic%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_repository.%s", name)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: resourceRemoteRepositoryCheckDestroy(fqrn),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(remoteGenericRepoBasicWithPropagate, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "generic"),
					resource.TestCheckResourceAttr(fqrn, "propagate_query_params", "true"),
				),
			},
		},
	})
}
