package remote_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccRemoteAllowDotsUnderscorersAndDashesInKeyGH129(t *testing.T) {
	_, fqrn, name := test.MkNames("terraform-local-test-repo-basic", "artifactory_remote_debian_repository")

	key := fmt.Sprintf("debian-remote.teleport_%d", test.RandomInt())
	remoteRepositoryBasic := fmt.Sprintf(`
		resource "artifactory_remote_debian_repository" "%s" {
			key              = "%s"
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: remoteRepositoryBasic,
				Check:  resource.TestCheckResourceAttr(fqrn, "key", key),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(key, "key"),
			},
		},
	})
}

func TestAccRemoteKeyHasSpecialCharsFails(t *testing.T) {
	const failKey = `
		resource "artifactory_remote_npm_repository" "terraform-remote-test-repo-basic" {
			key                     		= "IHave++special,Chars"
			url                     		= "https://registry.npmjs.org/"
			repo_layout_ref         		= "npm-default"
			retrieval_cache_period_seconds  = 70
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      failKey,
				ExpectError: regexp.MustCompile(".*expected value of key to not contain any of.*"),
			},
		},
	})
}

func TestAccRemoteDockerRepositoryDepTrue(t *testing.T) {
	const packageType = "docker"
	_, testCase := mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"external_dependencies_enabled":  true,
		"enable_token_authentication":    true,
		"block_pushing_schema1":          true,
		"priority_resolution":            false,
		"external_dependencies_patterns": []interface{}{"**/hub.docker.io/**", "**/bintray.jfrog.io/**"},
		"missed_cache_period_seconds":    1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false,
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	})
	resource.Test(t, testCase)
}

func TestAccRemoteDockerRepositoryDepFalse(t *testing.T) {
	const packageType = "docker"
	_, testCase := mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"external_dependencies_enabled":  false,
		"enable_token_authentication":    true,
		"block_pushing_schema1":          true,
		"priority_resolution":            false,
		"external_dependencies_patterns": []interface{}{"**/hub.docker.io/**", "**/bintray.jfrog.io/**"},
		"missed_cache_period_seconds":    1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false,
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	})
	resource.Test(t, testCase)
}

func TestAccRemoteDockerRepositoryDependenciesTrueEmptyListFails(t *testing.T) {
	const failKey = `
		resource "artifactory_remote_docker_repository" "terraform-remote-docker-repo-basic" {
			key                     		= "remote-docker"
			url                     		= "https://registry.npmjs.org/"
			retrieval_cache_period_seconds 	= 70
			enable_token_authentication    	= true
			block_pushing_schema1          	= true
			priority_resolution            	= false
			external_dependencies_enabled   = true
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      failKey,
				ExpectError: regexp.MustCompile(".*if `external_dependencies_enabled` is set to `true`, `external_dependencies_patterns` list must be set.*"),
			},
		},
	})
}

func TestAccRemoteDockerRepositoryDepListEmptyStringFails(t *testing.T) {
	const failKey = `
		resource "artifactory_remote_docker_repository" "terraform-remote-docker-repo-basic" {
			key                     		= "remote-docker"
			url                     		= "https://registry.npmjs.org/"
			retrieval_cache_period_seconds 	= 70
			enable_token_authentication    	= true
			block_pushing_schema1          	= true
			priority_resolution            	= false
			external_dependencies_enabled   = true
			external_dependencies_patterns 	= ["**/hub.docker.io/**", ""]
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      failKey,
				ExpectError: regexp.MustCompile(".*`external_dependencies_patterns` can't have an item of \"\" inside a list.*"),
			},
		},
	})
}

func TestAccRemoteDockerRepoUpdate(t *testing.T) {
	id, fqrn, name := test.MkNames("docker-remote-", "artifactory_remote_docker_repository")
	var testData = map[string]string{
		"resource_name":                  name,
		"repo_name":                      fmt.Sprintf("docker-remote-%d", id),
		"url":                            "https://registry-1.docker.io/",
		"assumed_offline_period_secs":    "300",
		"retrieval_cache_period_seconds": "43200",
		"missed_cache_period_seconds":    "7200",
		"excludes_pattern":               "nopat3,nopat2,nopat1",
		"includes_pattern":               "pat3,pat2,pat1",
		"notes":                          "internal description",
		"proxy":                          "",
		"username":                       "admin",
		"password":                       "password1",
		"xray_index":                     "false",
	}
	var testDataUpdated = map[string]string{
		"resource_name":                  name,
		"repo_name":                      fmt.Sprintf("docker-remote-%d", id),
		"url":                            "https://registry-1.docker.io/",
		"assumed_offline_period_secs":    "301",
		"retrieval_cache_period_seconds": "43201",
		"missed_cache_period_seconds":    "7201",
		"excludes_pattern":               "nopat3,nopat2,nopat1",
		"includes_pattern":               "pat3,pat2,pat1",
		"notes":                          "internal description",
		"proxy":                          "",
		"username":                       "admin1",
		"password":                       "password",
		"xray_index":                     "true",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, repoTemplate, testData),
				Check:  resource.ComposeTestCheckFunc(verifyRepository(fqrn, testData)),
			},
			{
				Config: util.ExecuteTemplate(fqrn, repoTemplate, testDataUpdated),
				Check:  resource.ComposeTestCheckFunc(verifyRepository(fqrn, testDataUpdated)),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

const repoTemplate = `
resource "artifactory_remote_docker_repository" "{{ .resource_name }}" {
  key                            = "{{ .repo_name }}"
  url                            = "{{ .url }}"
  assumed_offline_period_secs    = {{ .assumed_offline_period_secs }}

  retrieval_cache_period_seconds = {{ .retrieval_cache_period_seconds }}
  missed_cache_period_seconds    = {{ .missed_cache_period_seconds }}
  excludes_pattern               = "{{ .excludes_pattern }}"
  includes_pattern               = "{{ .includes_pattern }}"
  notes                          = "{{ .notes }}"
  proxy                          = "{{ .proxy }}"
  property_sets                  = [
    "artifactory",
  ]
  username                       = "{{ .username }}"
  password                       = "{{ .password }}"
  xray_index 					 = {{ .xray_index }}
}
`

func verifyRepository(fqrn string, testData map[string]string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(fqrn, "key", testData["repo_name"]),
		resource.TestCheckResourceAttr(fqrn, "url", testData["url"]),
		resource.TestCheckResourceAttr(fqrn, "assumed_offline_period_secs", testData["assumed_offline_period_secs"]),
		resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", testData["retrieval_cache_period_seconds"]),
		resource.TestCheckResourceAttr(fqrn, "missed_cache_period_seconds", testData["missed_cache_period_seconds"]),
		resource.TestCheckResourceAttr(fqrn, "excludes_pattern", testData["excludes_pattern"]),
		resource.TestCheckResourceAttr(fqrn, "includes_pattern", testData["includes_pattern"]),
		resource.TestCheckResourceAttr(fqrn, "notes", testData["notes"]),
		resource.TestCheckResourceAttr(fqrn, "proxy", testData["proxy"]),
		resource.TestCheckResourceAttr(fqrn, "username", testData["username"]),
		resource.TestCheckResourceAttr(fqrn, "xray_index", testData["xray_index"]),
	)
}

func TestAccRemoteDockerRepositoryWithAdditionalCheckFunctions(t *testing.T) {
	const packageType = "docker"
	_, testCase := mkRemoteTestCaseWithAdditionalCheckFunctions(packageType, t, map[string]interface{}{
		"external_dependencies_enabled":  true,
		"enable_token_authentication":    true,
		"block_pushing_schema1":          true,
		"priority_resolution":            false,
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
	const packageType = "cargo"
	_, testCase := mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"git_registry_url":            "https://github.com/rust-lang/foo.index",
		"anonymous_access":            true,
		"enable_sparse_index":         true,
		"priority_resolution":         false,
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

func TestAccRemoteCargoRepositoryWithAdditionalCheckFunctions(t *testing.T) {
	const packageType = "cargo"
	_, testCase := mkRemoteTestCaseWithAdditionalCheckFunctions(packageType, t, map[string]interface{}{
		"git_registry_url":            "https://github.com/rust-lang/foo.index",
		"anonymous_access":            true,
		"enable_sparse_index":         true,
		"priority_resolution":         false,
		"missed_cache_period_seconds": 1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"list_remote_folder_items":    true,
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
	const packageType = "helm"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"helm_charts_base_url":           "https://github.com/rust-lang/foo.index",
		"missed_cache_period_seconds":    1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"external_dependencies_enabled":  true,
		"priority_resolution":            false,
		"external_dependencies_patterns": []interface{}{"**github.com**"},
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false, // even when set to true, it seems to come back as false on the wire
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	}))
}

func TestAccRemoteHelmRepositoryDepFalse(t *testing.T) {
	const packageType = "helm"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"helm_charts_base_url":           "https://github.com/rust-lang/foo.index",
		"missed_cache_period_seconds":    1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"external_dependencies_enabled":  false,
		"priority_resolution":            false,
		"external_dependencies_patterns": []interface{}{"**github.com**"},
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false, // even when set to true, it seems to come back as false on the wire
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	}))
}

func TestAccRemoteHelmRepositoryWithAdditionalCheckFunctions(t *testing.T) {
	const packageType = "helm"
	resource.Test(mkRemoteTestCaseWithAdditionalCheckFunctions(packageType, t, map[string]interface{}{
		"helm_charts_base_url":           "https://github.com/rust-lang/foo.index",
		"missed_cache_period_seconds":    1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"external_dependencies_enabled":  true,
		"priority_resolution":            false,
		"list_remote_folder_items":       true,
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
	const packageType = "npm"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"list_remote_folder_items":             true,
		"priority_resolution":                  true,
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

func TestAccRemoteNpmRepositoryWithAdditionalCheckFunctions(t *testing.T) {
	const packageType = "npm"
	resource.Test(mkRemoteTestCaseWithAdditionalCheckFunctions(packageType, t, map[string]interface{}{
		"list_remote_folder_items":             true,
		"priority_resolution":                  true,
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

func TestAccRemotePypiRepository(t *testing.T) {
	const packageType = "pypi"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"pypi_registry_url":           "https://pypi.org",
		"priority_resolution":         true,
		"missed_cache_period_seconds": 1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false, // even when set to true, it seems to come back as false on the wire
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	}))
}

func TestAccRemotePypiRepositoryWithAdditionalCheckFunctions(t *testing.T) {
	const packageType = "pypi"
	resource.Test(mkRemoteTestCaseWithAdditionalCheckFunctions(packageType, t, map[string]interface{}{
		"pypi_registry_url":           "https://pypi.org",
		"priority_resolution":         true,
		"missed_cache_period_seconds": 1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"list_remote_folder_items":    true,
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false, // even when set to true, it seems to come back as false on the wire
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	}))
}

func TestAccRemoteMavenRepository(t *testing.T) {
	resource.Test(mkNewRemoteTestCase("maven", t, map[string]interface{}{
		"missed_cache_period_seconds":     1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
		"metadata_retrieval_timeout_secs": 30,   // https://github.com/jfrog/terraform-provider-artifactory/issues/509
		"list_remote_folder_items":        true,
		"content_synchronisation": map[string]interface{}{
			"enabled":                         false, // even when set to true, it seems to come back as false on the wire
			"statistics_enabled":              true,
			"properties_enabled":              true,
			"source_origin_absence_detection": true,
		},
	}))
}

func TestAccRemoteAllRepository(t *testing.T) {
	for _, repoType := range remote.PackageTypesLikeBasic {
		t.Run(repoType, func(t *testing.T) {
			resource.Test(mkNewRemoteTestCase(repoType, t, map[string]interface{}{
				"missed_cache_period_seconds": 1800,
			}))
		})
	}
}

func TestAccRemoteGoRepository(t *testing.T) {
	const packageType = "go"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"url":                         "https://proxy.golang.org/",
		"vcs_git_provider":            "ARTIFACTORY",
		"missed_cache_period_seconds": 1800,
	}))
}

func TestAccRemoteVcsRepository(t *testing.T) {
	const packageType = "vcs"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"url":                  "https://github.com/",
		"vcs_git_provider":     "CUSTOM",
		"vcs_git_download_url": "https://www.customrepo.com",
		// "max_unique_snapshots": 5, // commented out due to API bug in 7.49.3
	}))
}

func TestAccRemoteCocoapodsRepository(t *testing.T) {
	const packageType = "cocoapods"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"url":                         "https://github.com/",
		"vcs_git_provider":            "GITHUB",
		"pods_specs_repo_url":         "https://github.com/CocoaPods/Specs1",
		"missed_cache_period_seconds": 1800,
	}))
}

func TestAccRemoteComposerRepository(t *testing.T) {
	const packageType = "composer"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"url":                         "https://github.com/",
		"vcs_git_provider":            "GITHUB",
		"composer_registry_url":       "https://packagist1.org",
		"missed_cache_period_seconds": 1800,
	}))
}

func TestAccRemoteBowerRepository(t *testing.T) {
	const packageType = "bower"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"url":                         "https://github.com/",
		"vcs_git_provider":            "ARTIFACTORY",
		"bower_registry_url":          "https://registry1.bower.io",
		"missed_cache_period_seconds": 1800,
	}))
}

func TestAccRemoteConanRepository(t *testing.T) {
	const packageType = "conan"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"force_conan_authentication": true,
	}))
}

func TestAccRemoteNugetRepository(t *testing.T) {
	const packageType = "nuget"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"url":                         "https://www.nuget.org/",
		"download_context_path":       "api/v2/package",
		"force_nuget_authentication":  true,
		"missed_cache_period_seconds": 1800,
		"symbol_server_url":           "https://symbols.nuget.org/download/symbols",
	}))
}

func TestAccRemoteTerraformRepository(t *testing.T) {
	const packageType = "terraform"
	resource.Test(mkNewRemoteTestCase(packageType, t, map[string]interface{}{
		"url":                     "https://github.com/",
		"terraform_registry_url":  "https://registry.terraform.io",
		"terraform_providers_url": "https://releases.hashicorp.com",
		"repo_layout_ref":         "simple-default",
	}))
}

func TestAccRemoteAllGradleLikeRepository(t *testing.T) {
	for _, repoType := range repository.GradleLikePackageTypes {
		t.Run(repoType, func(t *testing.T) {
			resource.Test(mkNewRemoteTestCase(repoType, t, map[string]interface{}{
				"missed_cache_period_seconds": 1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
				"list_remote_folder_items":    true,
				"content_synchronisation": map[string]interface{}{
					"enabled":                         false, // even when set to true, it seems to come back as false on the wire
					"statistics_enabled":              true,
					"properties_enabled":              true,
					"source_origin_absence_detection": true,
				},
			}))
		})
	}
}

func TestAccRemotePypiRepositoryWithCustomRegistryUrl(t *testing.T) {
	const packageType = "pypi"
	extraFields := map[string]interface{}{
		"pypi_registry_url": "https://custom.PYPI.registry.url",
	}
	resource.Test(mkNewRemoteTestCase(packageType, t, extraFields))
}

func TestAccRemoteDockerRepositoryWithListRemoteFolderItems(t *testing.T) {
	extraFields := map[string]interface{}{
		"list_remote_folder_items": true,
	}
	resource.Test(mkNewRemoteTestCase("docker", t, extraFields))
}

func TestAccRemoteRepositoryChangeConfigGH148(t *testing.T) {
	_, fqrn, name := test.MkNames("github-remote", "artifactory_remote_generic_repository")
	const step1 = `
		locals {
		  allowed_github_repos = [
			"quixoten/gotee/releases/download/v*/gotee-*",
			"nats-io/gnatsd/releases/download/v*/gnatsd-*"
		  ]
		}
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
		  key = "{{ .name }}"
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
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
		  key = "{{ .name }}"
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate("one", step1, map[string]interface{}{
					"name": name,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "generic"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com"),
				),
			},
			{
				Config: util.ExecuteTemplate("two", step2, map[string]interface{}{
					"name": name,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "generic"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

func TestAccRemoteRepository_basic(t *testing.T) {
	id := rand.Int()
	name := fmt.Sprintf("terraform-remote-test-repo-basic%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_npm_repository.%s", name)
	const remoteRepoBasic = `
		resource "artifactory_remote_npm_repository" "%s" {
			key 				  = "%s"
			url                   = "https://registry.npmjs.org/"
			repo_layout_ref       = "npm-default"
			content_synchronisation {
				enabled = false
			}
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
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
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

func TestAccRemoteRepository_nugetNew(t *testing.T) {
	const remoteRepoNuget = `
		resource "artifactory_remote_nuget_repository" "%s" {
			key               		   = "%s"
			url               		   = "https://www.nuget.org/"
			repo_layout_ref   		   = "nuget-default"
			download_context_path	   = "Download"
			feed_context_path 		   = "/api/notdefault"
			force_nuget_authentication = true
		}
	`
	id := test.RandomInt()
	name := fmt.Sprintf("terraform-remote-test-repo-nuget%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_nuget_repository.%s", name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(remoteRepoNuget, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "v3_feed_url", "https://api.nuget.org/v3/index.json"),
					resource.TestCheckResourceAttr(fqrn, "feed_context_path", "/api/notdefault"),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", "true"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

// if you wish to override any of the default fields, just pass it as "extrFields" as these will overwrite
func mkNewRemoteTestCase(repoType string, t *testing.T, extraFields map[string]interface{}) (*testing.T, resource.TestCase) {
	_, fqrn, name := test.MkNames("terraform-remote-test-repo-full", fmt.Sprintf("artifactory_remote_%s_repository", repoType))
	certificateAlias := fmt.Sprintf("certificate-%d", test.RandomInt())

	defaultFields := map[string]interface{}{
		"key":      name,
		"url":      "https://registry.npmjs.org/",
		"username": "user",
		"password": "Passw0rd!",
		"proxy":    "",

		//"description":                        "foo", // the server returns this suffixed. Test separate
		"notes":                          "notes",
		"includes_pattern":               "**/*.js",
		"excludes_pattern":               "**/*.jsx",
		"repo_layout_ref":                "npm-default",
		"hard_fail":                      true,
		"offline":                        true,
		"blacked_out":                    true,
		"xray_index":                     test.RandBool(),
		"store_artifacts_locally":        true,
		"socket_timeout_millis":          25000,
		"local_address":                  "",
		"retrieval_cache_period_seconds": 70,
		// this doesn't get returned on a GET
		//"failed_retrieval_cache_period_secs": 140,
		"missed_cache_period_seconds":           2500,
		"unused_artifacts_cleanup_period_hours": 96,
		"assumed_offline_period_secs":           96,
		"share_configuration":                   true,
		"synchronize_properties":                true,
		"block_mismatching_mime_types":          true,
		"property_sets":                         []interface{}{"artifactory"},
		"allow_any_host_auth":                   true,
		"enable_cookie_management":              true,
		"bypass_head_requests":                  true,
		"client_tls_certificate":                certificateAlias,
		"content_synchronisation": map[string]interface{}{
			"enabled": false, // even when set to true, it seems to come back as false on the wire
		},
		"download_direct": true,
		"cdn_redirect":    false, // even when set to true, it comes back as false on the wire (presumably unless testing against a cloud platform)
	}
	allFields := util.MergeMaps(defaultFields, extraFields)
	allFieldsHcl := util.FmtMapToHcl(allFields)
	const remoteRepoFull = `
		resource "artifactory_remote_%s_repository" "%s" {
%s
		}
	`
	extraChecks := test.MapToTestChecks(fqrn, extraFields)
	defaultChecks := test.MapToTestChecks(fqrn, allFields)

	checks := append(defaultChecks, extraChecks...)
	config := fmt.Sprintf(remoteRepoFull, repoType, name, allFieldsHcl)

	var delCertTestCheckRepo = func(id string, request *resty.Request) (*resty.Response, error) {
		deleteTestCertificate(t, certificateAlias, security.CertificateEndpoint)
		return acctest.CheckRepo(id, request.AddRetryCondition(client.NeverRetry))
	}

	return t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			addTestCertificate(t, certificateAlias, security.CertificateEndpoint)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, delCertTestCheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	}
}

func addTestCertificate(t *testing.T, certificateAlias string, certificateEndpoint string) {
	restyClient := acctest.GetTestResty(t)

	certFileBytes, err := ioutil.ReadFile("../../../../../samples/cert.pem")
	if err != nil {
		t.Fatal(err)
	}

	_, err = restyClient.R().
		SetBody(string(certFileBytes)).
		SetContentLength(true).
		Post(fmt.Sprintf("%s%s", certificateEndpoint, certificateAlias))
	if err != nil {
		t.Fatal(err)
	}
}

func deleteTestCertificate(t *testing.T, certificateAlias string, certificateEndpoint string) {
	restyClient := acctest.GetTestResty(t)

	_, err := restyClient.R().
		Delete(fmt.Sprintf("%s%s", certificateEndpoint, certificateAlias))
	if err != nil {
		t.Fatal(err)
	}
}

func mkRemoteTestCaseWithAdditionalCheckFunctions(repoType string, t *testing.T, extraFields map[string]interface{}) (*testing.T, resource.TestCase) {
	_, fqrn, name := test.MkNames("terraform-remote-test-repo-full", fmt.Sprintf("artifactory_remote_%s_repository", repoType))

	defaultFields := map[string]interface{}{
		"key":      name,
		"url":      "https://registry.npmjs.org/",
		"username": "user",
		"password": "Passw0rd!",
		"proxy":    "",

		//"description":                        "foo", // the server returns this suffixed. Test separate
		"notes":                          "notes",
		"includes_pattern":               "**/*.js",
		"excludes_pattern":               "**/*.jsx",
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
		"missed_cache_period_seconds":           2500,
		"unused_artifacts_cleanup_period_hours": 96,
		"assumed_offline_period_secs":           96,
		"share_configuration":                   true,
		"synchronize_properties":                true,
		"block_mismatching_mime_types":          true,
		"property_sets":                         []interface{}{"artifactory"},
		"allow_any_host_auth":                   true,
		"enable_cookie_management":              true,
		"bypass_head_requests":                  true,
		"client_tls_certificate":                "",
		"content_synchronisation": map[string]interface{}{
			"enabled": false, // even when set to true, it seems to come back as false on the wire
		},
	}
	allFields := util.MergeMaps(defaultFields, extraFields)
	allFieldsHcl := util.FmtMapToHcl(allFields)
	const remoteRepoFull = `
		resource "artifactory_remote_%s_repository" "%s" {
%s
		}
	`
	extraChecks := test.MapToTestChecks(fqrn, extraFields)
	defaultChecks := test.MapToTestChecks(fqrn, allFields)

	var addCheckFunctions = []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("remote", repoType)(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
	}

	checks := append(defaultChecks, append(extraChecks, addCheckFunctions...)...)
	config := fmt.Sprintf(remoteRepoFull, repoType, name, allFieldsHcl)

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	}
}

func TestAccRemoteRepository_generic_with_propagate(t *testing.T) {

	const remoteGenericRepoBasicWithPropagate = `
		resource "artifactory_remote_generic_repository" "%s" {
			key                     		= "%s"
			description 					= "This is a test"
			url                     		= "https://registry.npmjs.org/"
			repo_layout_ref         		= "simple-default"
			propagate_query_params  		= true
			retrieval_cache_period_seconds  = 70
		}
	`
	id := test.RandomInt()
	name := fmt.Sprintf("terraform-remote-test-repo-basic%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_generic_repository.%s", name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(remoteGenericRepoBasicWithPropagate, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "generic"),
					resource.TestCheckResourceAttr(fqrn, "propagate_query_params", "true"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

func TestAccRemoteRepository_gems_with_propagate_fails(t *testing.T) {
	for _, repoType := range remote.PackageTypesLikeBasic {
		const remoteGemsRepoBasicWithPropagate = `
		resource "artifactory_remote_%s_repository" "%s" {
			key                     		= "%s"
			description 					= "This is a test"
			url                     		= "https://rubygems.org/"
			repo_layout_ref         		= "simple-default"
			propagate_query_params  		= true
		}
	`
		id := test.RandomInt()
		name := fmt.Sprintf("terraform-remote-test-repo-basic%d", id)
		fqrn := fmt.Sprintf("artifactory_remote_gems_repository.%s", name)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { acctest.PreCheck(t) },
			ProviderFactories: acctest.ProviderFactories,
			CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
			Steps: []resource.TestStep{
				{
					Config:      fmt.Sprintf(remoteGemsRepoBasicWithPropagate, repoType, name, name),
					ExpectError: regexp.MustCompile(".*Unsupported argument.*"),
				},
			},
		})
	}
}

func TestRemoteRepoResourceStateUpgradeV1(t *testing.T) {
	v1Data := map[string]interface{}{
		"description":            "This is a test",
		"propagate_query_params": "true",
		"repo_layout_ref":        "simple-default",
	}
	v2Data := map[string]interface{}{
		"description":     "This is a test",
		"repo_layout_ref": "simple-default",
	}

	actual, err := remote.ResourceStateUpgradeV1(context.Background(), v1Data, nil)

	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(v2Data, actual) {
		t.Fatalf("expected: %v\n\ngot: %v", v2Data, actual)
	}
}

func TestRemoteMavenRepoResourceStateUpgradeV1(t *testing.T) {
	v1Data := map[string]interface{}{
		"description":                        "This is a test",
		"url":                                "https://repo1.maven.org/maven2/",
		"metadata_retrieval_timeout_seconds": 120,
	}
	v2Data := map[string]interface{}{
		"description":                     "This is a test",
		"url":                             "https://repo1.maven.org/maven2/",
		"metadata_retrieval_timeout_secs": 120,
	}

	actual, err := remote.ResourceMavenStateUpgradeV1(context.Background(), v1Data, nil)

	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(v2Data, actual) {
		t.Fatalf("expected: %v\n\ngot: %v", v2Data, actual)
	}
}

// https://github.com/jfrog/terraform-provider-artifactory/issues/225
func TestAccRemoteRepository_MissedRetrievalCachePeriodSecs_retained_between_updates_GH225(t *testing.T) {
	_, fqrn, name := test.MkNames("terraform-remote-test-cran-remote-", "artifactory_remote_cran_repository")

	remoteRepositoryInit := fmt.Sprintf(`
		resource "artifactory_remote_cran_repository" "%s" {
			key              = "%s"
			repo_layout_ref  = "bower-default"
			url              = "https://cran.r-project.org/"
			notes            = "managed by terraform"
			property_sets    = ["artifactory"]
			unused_artifacts_cleanup_period_hours = 10100
			retrieval_cache_period_seconds        = 600
			missed_cache_period_seconds           = 1800
		}
	`, name, name)

	remoteRepositoryUpdate := fmt.Sprintf(`
		resource "artifactory_remote_cran_repository" "%s" {
			key              = "%s"
			repo_layout_ref  = "simple-default"
			url              = "https://cran.r-project.org/"
			notes            = "managed by terraform"
			property_sets    = ["artifactory"]
			unused_artifacts_cleanup_period_hours = 10100
			retrieval_cache_period_seconds        = 600
			missed_cache_period_seconds           = 1800
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: remoteRepositoryInit,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "missed_cache_period_seconds", "1800"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "600"),
				),
			},
			{
				Config: remoteRepositoryUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "missed_cache_period_seconds", "1800"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "600"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

// https://github.com/jfrog/terraform-provider-artifactory/issues/241
func TestAccRemoteRepository_assumed_offline_period_secs_has_default_value_GH241(t *testing.T) {
	_, fqrn, name := test.MkNames("terraform-remote-test-repo-docker", "artifactory_remote_docker_repository")

	remoteRepositoryInit := fmt.Sprintf(`
		resource "artifactory_remote_docker_repository" "%s" {
			key                                   = "%s"
			description                           = "DockerHub mirror"
			url                                   = "https://registry-1.docker.io/"
			external_dependencies_enabled         = true
			external_dependencies_patterns		  = ["**"]	
			enable_token_authentication           = true
			block_pushing_schema1                 = false
			unused_artifacts_cleanup_period_hours = 2 * 7 * 24
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: remoteRepositoryInit,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "assumed_offline_period_secs", "300"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

func TestAccRemoteProxyUpdateGH2(t *testing.T) {
	_, fqrn, name := test.MkNames("terraform-remote-test-go-remote-proxy-", "artifactory_remote_go_repository")

	fakeProxy := "test-proxy"

	remoteRepositoryWithProxy := fmt.Sprintf(`
		resource "artifactory_remote_go_repository" "%s" {
			key             = "%s"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
			proxy           = "%s"
		}
	`, name, name, fakeProxy)

	remoteRepositoryResetProxyWithEmptyString := fmt.Sprintf(`
		resource "artifactory_remote_go_repository" "%s" {
			key             = "%s"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
			proxy           = ""
		}
	`, name, name)

	remoteRepositoryResetProxyWithNoAttr := fmt.Sprintf(`
		resource "artifactory_remote_go_repository" "%s" {
			key             = "%s"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
		}
	`, name, name)

	testProxyKey := "test-proxy"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProxy(t, testProxyKey)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: acctest.VerifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProxy(t, testProxyKey)
			return acctest.CheckRepo(id, request)
		}),
		Steps: []resource.TestStep{
			{
				Config: remoteRepositoryWithProxy,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "proxy", fakeProxy),
				),
			},
			{
				Config: remoteRepositoryResetProxyWithEmptyString,
				Check:  resource.TestCheckResourceAttr(fqrn, "proxy", ""),
			},
			{
				Config: remoteRepositoryWithProxy,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "proxy", fakeProxy),
				),
			},
			{
				Config: remoteRepositoryResetProxyWithNoAttr,
				Check:  resource.TestCheckResourceAttr(fqrn, "proxy", ""),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

func TestAccRemoteRepositoryWithProjectAttributesGH318(t *testing.T) {

	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	projectEnv := test.RandSelect("DEV", "PROD").(string)
	repoName := fmt.Sprintf("%s-pypi-remote", projectKey)

	_, fqrn, name := test.MkNames(repoName, "artifactory_remote_pypi_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
		"projectEnv": projectEnv,
	}
	remoteRepositoryBasic := util.ExecuteTemplate("TestAccRemotePyPiRepository", `
		resource "artifactory_remote_pypi_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = "{{ .projectKey }}"
	 	  project_environments = ["{{ .projectEnv }}"]
		  url                  = "http://tempurl.org"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		CheckDestroy: acctest.VerifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: remoteRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.0", projectEnv),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

func TestAccRemoteRepositoryWithInvalidProjectKeyGH318(t *testing.T) {

	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	repoName := fmt.Sprintf("%s-pypi-remote", projectKey)

	_, fqrn, name := test.MkNames(repoName, "artifactory_remote_pypi_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	remoteRepositoryBasic := util.ExecuteTemplate("TestAccRemotePyPiRepository", `
		resource "artifactory_remote_pypi_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = "invalid-project-key-too-long-really-long"
		  url                  = "http://tempurl.org"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		CheckDestroy: acctest.VerifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      remoteRepositoryBasic,
				ExpectError: regexp.MustCompile(".*project_key must be 2 - 32 lowercase alphanumeric and hyphen characters"),
			},
		},
	})
}

func TestAccRemoteRepository_excludes_pattern_reset(t *testing.T) {
	_, fqrn, name := test.MkNames("generic-remote", "artifactory_remote_generic_repository")
	const step1 = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
		  key              = "{{ .name }}"
		  url              = "https://github.com"
		  repo_layout_ref  = "simple-default"
		  excludes_pattern = "fake-pattern"
		}
	`
	const step2 = `
		resource "artifactory_remote_generic_repository" "{{ .name }}" {
		  key              = "{{ .name }}"
		  url              = "https://github.com"
		  repo_layout_ref  = "simple-default"
		  excludes_pattern = ""
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate("one", step1, map[string]interface{}{
					"name": name,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "generic"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "fake-pattern"),
				),
			},
			{
				Config: util.ExecuteTemplate("two", step2, map[string]interface{}{
					"name": name,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "generic"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", ""),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}
