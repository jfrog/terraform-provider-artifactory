package artifactory

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccLocalAllowDotsUnderscorersAndDashesInKeyGH129(t *testing.T) {
	_, fqrn, name := mkNames("terraform-local-test-repo-basic", "artifactory_remote_repository")

	key := fmt.Sprintf("debian-remote.teleport_%d", randomInt())
	localRepositoryBasic := fmt.Sprintf(`
		resource "artifactory_remote_repository" "%s" {
			key              = "%s"
			package_type     = "debian"
			repo_layout_ref  = "simple-default"
			url              = "https://deb.releases.teleport.dev/"
			notes            = "managed by terraform"
			property_sets    = ["artifactory"]
			includes_pattern = "**/*"
			content_synchronisation {
				enabled = false
			}
		}
	`, name, key)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check:  resource.TestCheckResourceAttr(fqrn, "key", key),
			},
		},
	})
}

func TestKeyHasSpecialCharsFails(t *testing.T) {
	const failKey = `
		resource "artifactory_remote_repository" "terraform-remote-test-repo-basic" {
			key                     = "IHave++special,Chars"
			package_type            = "npm"
			url                     = "https://registry.npmjs.org/"
			repo_layout_ref         = "npm-default"
			propagate_query_params  = true
			retrieval_cache_period_seconds        = 70
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      failKey,
				ExpectError: regexp.MustCompile(".*expected value of key to not contain any of.*"),
			},
		},
	})
}

func TestAccRemoteDockerRepository(t *testing.T) {
	_, testCase := mkNewRemoteTestCase("docker", t, map[string]interface{}{
		"external_dependencies_enabled":  true,
		"enable_token_authentication":    true,
		"block_pushing_schema1":          true,
		"external_dependencies_patterns": []interface{}{"**/hub.docker.io/**", "**/bintray.jfrog.io/**"},
		"missed_cache_period_seconds":    1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false, // even when set to true, it seems to come back as false on the wire
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	})
	resource.Test(t, testCase)
}

func TestAccRemoteCargoRepository(t *testing.T) {
	_, testCase := mkNewRemoteTestCase("cargo", t, map[string]interface{}{
		"git_registry_url":            "https://github.com/rust-lang/foo.index",
		"anonymous_access":            true,
		"missed_cache_period_seconds": 1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false, // even when set to true, it seems to come back as false on the wire
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	})
	resource.Test(t, testCase)
}

func TestAccRemoteHelmRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase("helm", t, map[string]interface{}{
		"helm_charts_base_url":           "https://github.com/rust-lang/foo.index",
		"missed_cache_period_seconds":    1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"external_dependencies_enabled":  true,
		"external_dependencies_patterns": []interface{}{"**github.com**"},
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false, // even when set to true, it seems to come back as false on the wire
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	}))
}

func TestAccRemoteNpmRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase("npm", t, map[string]interface{}{
		"list_remote_folder_items":             true,
		"mismatching_mime_types_override_list": "application/json,application/xml",
		"missed_cache_period_seconds":          1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false, // even when set to true, it seems to come back as false on the wire
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	}))
}

func TestAccRemoteRepositoryChangeConfigGH148(t *testing.T) {
	_, fqrn, name := mkNames("github-remote", "artifactory_remote_repository")
	const step1 = `
		locals {
		  allowed_github_repos = [
			"quixoten/gotee/releases/download/v*/gotee-*",
			"nats-io/gnatsd/releases/download/v*/gnatsd-*"
		  ]
		}
		resource "artifactory_remote_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		  package_type = "generic"
		  url = "https://github.com"
		  repo_layout_ref = "simple-default"
		  notes = "managed by terraform"
		  content_synchronisation {
			enabled = false
		  }
		  bypass_head_requests = true
		  property_sets = [
			"artifactory"
		  ]
		  includes_pattern = join(", ", local.allowed_github_repos)
		}
	`
	const step2 = `
		locals {
		  allowed_github_repos = [
			"quixoten/gotee/releases/download/v*/gotee-*"
		  ]
		}
		resource "artifactory_remote_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		  package_type = "generic"
		  url = "https://github.com"
		  repo_layout_ref = "simple-default"
		  notes = "managed by terraform"
		  content_synchronisation {
			enabled = false
		  }
		  bypass_head_requests = true
		  property_sets = [
			"artifactory"
		  ]
		  includes_pattern = join(", ", local.allowed_github_repos)
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: executeTemplate("one", step1, map[string]interface{}{
					"name": name,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "generic"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com"),
				),
			},
			{
				Config: executeTemplate("two", step2, map[string]interface{}{
					"name": name,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "generic"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com"),
				),
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
			content_synchronisation {
				enabled = false
			}
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(remoteRepoBasic, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "npm"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://registry.npmjs.org/"),
					resource.TestCheckResourceAttr(fqrn, "content_synchronisation.0.enabled", "false"),
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
	id := randomInt()
	name := fmt.Sprintf("terraform-remote-test-repo-nuget%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_repository.%s", name)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
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

func TestAllLegacyRemoteRepoTypes(t *testing.T) {
	//
	for _, repo := range repoTypesSupported {
		if repo != "nuget" { // this requires special testing
			t.Run(fmt.Sprintf("TestLegacyRemote%sRepo", strings.Title(strings.ToLower(repo))), func(t *testing.T) {
				// NuGet Repository configuration is missing mandatory field downloadContextPath
				resource.Test(mkLegacyRemoteTestCase(repo, t))
			})
		}
	}
}

// if you wish to override any of the default fields, just pass it as "extrFields" as these will overwrite
func mkNewRemoteTestCase(repoType string, t *testing.T, extraFields map[string]interface{}) (*testing.T, resource.TestCase) {
	_, fqrn, name := mkNames("terraform-remote-test-repo-full", fmt.Sprintf("artifactory_remote_%s_repository", repoType))

	defaultFields := map[string]interface{}{
		"key":      name,
		"url":      "https://registry.npmjs.org/",
		"username": "user",
		// This returns encrypted. Can't be tested
		//"password":                           "foo",
		"proxy": "",

		//"description":                        "foo", // the server returns this suffixed. Test seperate
		"notes":                          "notes",
		"includes_pattern":               "**/*.js",
		"excludes_pattern":               "**/*.jsx",
		"repo_layout_ref":                "npm-default",
		"hard_fail":                      true,
		"offline":                        true,
		"blacked_out":                    true,
		"xray_index":                     true,
		"store_artifacts_locally":        true,
		"socket_timeout_millis":          25000,
		"local_address":                  "",
		"retrieval_cache_period_seconds": 70,
		// this doesn't get returned on a GET
		//"failed_retrieval_cache_period_secs": 140,
		"missed_cache_period_seconds":             2500,
		"unused_artifacts_cleanup_period_enabled": true,
		"unused_artifacts_cleanup_period_hours":   96,
		"assumed_offline_period_secs":             96,
		"share_configuration":                     true,
		"synchronize_properties":                  true,
		"block_mismatching_mime_types":            true,
		"property_sets":                           []interface{}{"artifactory"},
		"allow_any_host_auth":                     true,
		"enable_cookie_management":                true,
		"bypass_head_requests":                    true,
		"client_tls_certificate":                  "",
		"content_synchronisation": map[string]interface{}{
			"enabled": false, // even when set to true, it seems to come back as false on the wire
		},
	}
	allFields := mergeMaps(defaultFields, extraFields)
	allFieldsHcl := fmtMapToHcl(allFields)
	const remoteRepoFull = `
		resource "artifactory_remote_%s_repository" "%s" {
%s
		}
	`
	extraChecks := mapToTestChecks(fqrn, extraFields)
	defaultChecks := mapToTestChecks(fqrn, allFields)

	checks := append(defaultChecks, extraChecks...)
	config := fmt.Sprintf(remoteRepoFull, repoType, name, allFieldsHcl)

	return t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	}
}

func mkLegacyRemoteTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
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

	_, fqrn, name := mkNames("terraform-remote-test-repo-full", "artifactory_remote_repository")
	return t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
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
					resource.TestCheckResourceAttr(fqrn, "property_sets.0", "artifactory"),
					resource.TestCheckResourceAttr(fqrn, "allow_any_host_auth", "false"),
					resource.TestCheckResourceAttr(fqrn, "enable_cookie_management", "true"),
					resource.TestCheckResourceAttr(fqrn, "client_tls_certificate", ""),
					resource.TestCheckResourceAttr(fqrn, "remote_repo_checksum_policy_type", "ignore-and-generate"),
				),
			},
		},
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
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
	id := randomInt()
	name := fmt.Sprintf("terraform-remote-test-repo-basic%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_repository.%s", name)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
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

// https://github.com/jfrog/terraform-provider-artifactory/issues/225
func TestAccRemoteLegacyRepository_MissedRetrievalCachePeriodSecs_retained_between_updates_GH225(t *testing.T) {
	_, fqrn, name := mkNames("terraform-remote-test-repo-basic", "artifactory_remote_repository")

	key := fmt.Sprintf("cran-remote-%d", randomInt())
	remoteRepositoryInit := fmt.Sprintf(`
		resource "artifactory_remote_repository" "%s" {
			key              = "%s"
			package_type     = "cran"
			repo_layout_ref  = "bower-default"
			url              = "https://cran.r-project.org/"
			notes            = "managed by terraform"
			property_sets    = ["artifactory"]
			unused_artifacts_cleanup_period_hours = 10100
			retrieval_cache_period_seconds        = 600
			missed_cache_period_seconds           = 1800
		}
	`, name, key)

	remoteRepositoryUpdate := fmt.Sprintf(`
		resource "artifactory_remote_repository" "%s" {
			key              = "%s"
			package_type     = "cran"
			repo_layout_ref  = "simple-default"
			url              = "https://cran.r-project.org/"
			notes            = "managed by terraform"
			property_sets    = ["artifactory"]
			unused_artifacts_cleanup_period_hours = 10100
			retrieval_cache_period_seconds        = 600
			missed_cache_period_seconds           = 1800
		}
	`, name, key)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: remoteRepositoryInit,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", key),
					resource.TestCheckResourceAttr(fqrn, "missed_cache_period_seconds", "1800"),
				),
			},
			{
				Config: remoteRepositoryUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", key),
					resource.TestCheckResourceAttr(fqrn, "missed_cache_period_seconds", "1800"),
				),
			},
		},
	})
}

// https://github.com/jfrog/terraform-provider-artifactory/issues/241
func TestAccRemoteRepository_assumed_offline_period_secs_has_default_value_GH241(t *testing.T) {
	_, fqrn, name := mkNames("terraform-remote-test-repo-docker", "artifactory_remote_docker_repository")

	key := fmt.Sprintf("docker-remote-%d", randomInt())
	remoteRepositoryInit := fmt.Sprintf(`
		resource "artifactory_remote_docker_repository" "%s" {
			key                                   = "%s"
			description                           = "DockerHub mirror"
			url                                   = "https://registry-1.docker.io/"
			external_dependencies_enabled         = true
			enable_token_authentication           = true
			block_pushing_schema1                 = false
			unused_artifacts_cleanup_period_hours = 2 * 7 * 24
		}
	`, name, key)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: remoteRepositoryInit,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", key),
					resource.TestCheckResourceAttr(fqrn, "assumed_offline_period_secs", "300"),
				),
			},
		},
	})
}

func TestAccRemoteProxyUpdateGH2(t *testing.T) {
	_, fqrn, name := mkNames("terraform-remote-test-repo-proxy", "artifactory_remote_repository")

	key := fmt.Sprintf("go-remote.proxy_%d", randomInt())
	fakeProxy := "test-proxy"

	remoteRepositoryWithProxy := fmt.Sprintf(`
		resource "artifactory_remote_repository" "%s" {
			key             = "%s"
			package_type    = "go"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
			proxy           = "%s"
		}
	`, name, key, fakeProxy)

	remoteRepositoryResetProxyWithEmptyString := fmt.Sprintf(`
		resource "artifactory_remote_repository" "%s" {
			key             = "%s"
			package_type    = "go"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
			proxy           = ""
		}
	`, name, key)

	remoteRepositoryResetProxyWithNoAttr := fmt.Sprintf(`
		resource "artifactory_remote_repository" "%s" {
			key             = "%s"
			package_type    = "go"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
		}
	`, name, key)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: remoteRepositoryWithProxy,
				Check:  resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", key),
					resource.TestCheckResourceAttr(fqrn, "proxy", fakeProxy),
				),
			},
			{
				Config: remoteRepositoryResetProxyWithEmptyString,
				Check: resource.TestCheckResourceAttr(fqrn, "proxy", ""),
			},
			{
				Config: remoteRepositoryWithProxy,
				Check:  resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", key),
					resource.TestCheckResourceAttr(fqrn, "proxy", fakeProxy),
				),
			},
			{
				Config: remoteRepositoryResetProxyWithNoAttr,
				Check: resource.TestCheckResourceAttr(fqrn, "proxy", ""),
			},
		},
	})
}
