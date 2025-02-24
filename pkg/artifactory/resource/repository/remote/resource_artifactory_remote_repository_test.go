package remote_test

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccRemoteUpgradeFromVersionWithNoDisableProxyAttr(t *testing.T) {
	providerHost := os.Getenv("TF_ACC_PROVIDER_HOST")
	if providerHost == "registry.opentofu.org" {
		t.Skipf("provider host is registry.opentofu.org. Previous version of Artifactory provider is unknown to OpenTofu.")
	}

	_, fqrn, name := testutil.MkNames("tf-go-remote-", "artifactory_remote_go_repository")

	params := map[string]string{
		"name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteGoRepository", `
		resource "artifactory_remote_go_repository" "{{ .name }}" {
			key             = "{{ .name }}"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
			list_remote_folder_items = true
		}

	`, params)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "8.1.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config:             config,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "go-default"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://gocenter.io"),
				),
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "12.8.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				// Plugin Framework UpgradeState is not implemented in remoteResource for schema v2 to v3
				// due to Golang/Terraform type complexity.
				// Instead user/practitioner should upgrade from v8 to v12.8.0 (last version in SDKv2 with state upgrade) first
				// then upgrade to >12.8.1
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccRemoteAllowDotsUnderscorersAndDashesInKeyGH129(t *testing.T) {
	_, fqrn, name := testutil.MkNames("remote-test-repo-basic", "artifactory_remote_debian_repository")

	key := fmt.Sprintf("debian-remote.teleport_%d", testutil.RandomInt())
	remoteRepositoryBasic := fmt.Sprintf(`
		resource "artifactory_remote_debian_repository" "%s" {
			key              = "%s"
			repo_layout_ref  = "simple-default"
			url              = "https://deb.releases.teleport.dev/"
			notes            = "managed by terraform"
			property_sets    = ["artifactory"]
			includes_pattern = "**/*"
		}
	`, name, key)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
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
		resource "artifactory_remote_npm_repository" "remote-test-repo-basic" {
			key                     		= "IHave++special,Chars"
			url                     		= "https://registry.npmjs.org/"
			repo_layout_ref         		= "npm-default"
			retrieval_cache_period_seconds  = 70
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      failKey,
				ExpectError: regexp.MustCompile(".*Attribute key cannot contain spaces or special characters.*"),
			},
		},
	})
}

func verifyRepository(fqrn string, testData map[string]string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(fqrn, "key", testData["repo_name"]),
		resource.TestCheckResourceAttr(fqrn, "url", testData["url"]),
		resource.TestCheckResourceAttr(fqrn, "assumed_offline_period_secs", testData["assumed_offline_period_secs"]),
		resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", testData["retrieval_cache_period_seconds"]),
		resource.TestCheckResourceAttr(fqrn, "missed_cache_period_seconds", testData["missed_cache_period_seconds"]),
		resource.TestCheckResourceAttr(fqrn, "excludes_pattern", testData["excludes_pattern"]),
		resource.TestCheckResourceAttr(fqrn, "includes_pattern", testData["includes_pattern"]),
		resource.TestCheckResourceAttr(fqrn, "project_id", testData["project_id"]),
		resource.TestCheckResourceAttr(fqrn, "notes", testData["notes"]),
		resource.TestCheckResourceAttr(fqrn, "proxy", testData["proxy"]),
		resource.TestCheckResourceAttr(fqrn, "username", testData["username"]),
		resource.TestCheckResourceAttr(fqrn, "xray_index", testData["xray_index"]),
		resource.TestCheckResourceAttr(fqrn, "property_sets.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "property_sets.0", "artifactory"),
	)
}

func TestAccRemoteRepositoryChangeConfigGH148(t *testing.T) {
	_, fqrn, name := testutil.MkNames("github-remote", "artifactory_remote_generic_repository")
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
		  bypass_head_requests = true
		  property_sets = [
			"artifactory"
		  ]
		  includes_pattern = join(", ", local.allowed_github_repos)
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate("one", step1, map[string]interface{}{
					"name": name,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com"),
				),
			},
			{
				Config: util.ExecuteTemplate("two", step2, map[string]interface{}{
					"name": name,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccRemoteRepository_basic(t *testing.T) {
	id := rand.Int()
	name := fmt.Sprintf("remote-test-repo-basic%d", id)
	fqrn := fmt.Sprintf("artifactory_remote_npm_repository.%s", name)
	const remoteRepoBasic = `
		resource "artifactory_remote_npm_repository" "%s" {
			key 				  = "%s"
			url                   = "https://registry.npmjs.org/"
			repo_layout_ref       = "npm-default"
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(remoteRepoBasic, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "url", "https://registry.npmjs.org/"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

// if you wish to override any of the default fields, just pass it as "extrFields" as these will overwrite
func mkNewRemoteTestCase(packageType string, t *testing.T, extraFields map[string]interface{}) (*testing.T, resource.TestCase) {
	_, fqrn, name := testutil.MkNames("remote-test-repo-full", fmt.Sprintf("artifactory_remote_%s_repository", packageType))
	certificateAlias := fmt.Sprintf("certificate-%d", testutil.RandomInt())

	defaultFields := map[string]interface{}{
		"key":                            name,
		"url":                            "https://registry.npmjs.org/",
		"username":                       "user",
		"password":                       "Passw0rd!",
		"proxy":                          "",
		"description":                    "description",
		"notes":                          "notes",
		"includes_pattern":               "**/*.js",
		"excludes_pattern":               "**/*.jsx",
		"repo_layout_ref":                "npm-default",
		"hard_fail":                      true,
		"offline":                        true,
		"blacked_out":                    true,
		"xray_index":                     testutil.RandBool(),
		"store_artifacts_locally":        true,
		"socket_timeout_millis":          25000,
		"local_address":                  "",
		"retrieval_cache_period_seconds": 70,
		// this doesn't get returned on a GET
		//"failed_retrieval_cache_period_secs": 140,
		"missed_cache_period_seconds":           2500,
		"unused_artifacts_cleanup_period_hours": 96,
		"assumed_offline_period_secs":           96,
		"synchronize_properties":                true,
		"block_mismatching_mime_types":          true,
		"property_sets":                         []interface{}{"artifactory"},
		"allow_any_host_auth":                   true,
		"enable_cookie_management":              true,
		"bypass_head_requests":                  true,
		"client_tls_certificate":                certificateAlias,
		"download_direct":                       true,
		"cdn_redirect":                          false, // even when set to true, it comes back as false on the wire (presumably unless testing against a cloud platform)
		"disable_url_normalization":             true,
	}
	allFields := utilsdk.MergeMaps(defaultFields, extraFields)
	allFieldsHcl := utilsdk.FmtMapToHcl(allFields)
	const remoteRepoFull = `
		resource "artifactory_remote_%s_repository" "%s" {
%s
		}
	`
	extraChecks := testutil.MapToTestChecks(fqrn, extraFields)
	defaultChecks := testutil.MapToTestChecks(fqrn, allFields)

	checks := append(defaultChecks, extraChecks...)
	config := fmt.Sprintf(remoteRepoFull, packageType, name, allFieldsHcl)

	updatedFields := utilsdk.MergeMaps(defaultFields, extraFields, map[string]any{
		"description": "",
		"notes":       "",
	})
	updatedFieldsHcl := utilsdk.FmtMapToHcl(updatedFields)
	updatedConfig := fmt.Sprintf(remoteRepoFull, packageType, name, updatedFieldsHcl)
	updatedChecks := testutil.MapToTestChecks(fqrn, updatedFields)
	updatedChecks = append(updatedChecks, extraChecks...)

	var delCertTestCheckRepo = func(id string, request *resty.Request) (*resty.Response, error) {
		deleteTestCertificate(t, certificateAlias, security.CertificateEndpoint)
		return acctest.CheckRepo(id, request.AddRetryCondition(client.NeverRetry))
	}

	return t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			addTestCertificate(t, certificateAlias, security.CertificateEndpoint)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", delCertTestCheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			{
				Config: updatedConfig,
				Check:  resource.ComposeTestCheckFunc(updatedChecks...),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"password"},
			},
		},
	}
}

func addTestCertificate(t *testing.T, certificateAlias string, certificateEndpoint string) {
	restyClient := acctest.GetTestResty(t)

	certFileBytes, err := os.ReadFile("../../../../../samples/cert.pem")
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

// https://github.com/jfrog/terraform-provider-artifactory/issues/225
func TestAccRemoteRepository_MissedRetrievalCachePeriodSecs_retained_between_updates_GH225(t *testing.T) {
	_, fqrn, name := testutil.MkNames("remote-test-cran-remote-", "artifactory_remote_cran_repository")

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
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
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
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

// https://github.com/jfrog/terraform-provider-artifactory/issues/241
func TestAccRemoteRepository_assumed_offline_period_secs_has_default_value_GH241(t *testing.T) {
	_, fqrn, name := testutil.MkNames("remote-test-repo-docker", "artifactory_remote_docker_repository")

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
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: remoteRepositoryInit,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "assumed_offline_period_secs", "300"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccRemoteProxyUpdateGH2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("remote-test-go-remote-proxy-", "artifactory_remote_go_repository")

	fakeProxy := "test-proxy"

	remoteRepositoryWithProxy := fmt.Sprintf(`
		resource "artifactory_proxy" "%s" {
			key  = "%s"
			host = "http://tempurl.org"
			port = 8080
		}
		
		resource "artifactory_remote_go_repository" "%s" {
			key             = "%s"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
			proxy           = artifactory_proxy.%s.key
		}
	`, fakeProxy, fakeProxy, name, name, fakeProxy)

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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
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
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccRemoteDisableDefaultProxyGH739(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	_, fqrn, name := testutil.MkNames("tf-go-remote-", "artifactory_remote_go_repository")

	params := map[string]string{
		"name": name,
	}
	// Default proxy will be assigned to the repository no matter what, and it's impossible to remove it by submitting an empty string or
	// removing the attribute. If `disable_proxy` is set to true, then both repo and default proxies are removed and not returned in the
	// GET body.
	config := util.ExecuteTemplate("TestAccRemoteGoRepository", `
		resource "artifactory_proxy" "my-proxy" {
		  	key               = "my-proxy"
		  	host              = "my-proxy.mycompany.com"
		  	port              = 8888
		  	username          = "user1"
		  	password          = "password"
		  	nt_host           = "MYCOMPANY.COM"
		  	nt_domain         = "MYCOMPANY"
		  	platform_default  = false
		  	redirect_to_hosts = ["redirec-host.mycompany.com"]
		  	services          = ["jfrt"]
		}		

		resource "artifactory_remote_go_repository" "{{ .name }}" {
			key             = "{{ .name }}"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
			disable_proxy 	= true
			depends_on 		= [artifactory_proxy.my-proxy]
		}

	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "proxy", ""),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccRemoteDisableProxyGH739(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	_, fqrn, name := testutil.MkNames("tf-go-remote-", "artifactory_remote_go_repository")

	params := map[string]string{
		"name": name,
	}
	config := util.ExecuteTemplate("TestAccRemoteGoRepository", `
		resource "artifactory_proxy" "my-proxy" {
		  	key               = "my-proxy"
		  	host              = "my-proxy.mycompany.com"
		  	port              = 8888
		  	username          = "user1"
		  	password          = "password"
		  	nt_host           = "MYCOMPANY.COM"
		  	nt_domain         = "MYCOMPANY"
		  	platform_default  = false
		  	redirect_to_hosts = ["redirec-host.mycompany.com"]
		}		

		resource "artifactory_remote_go_repository" "{{ .name }}" {
			key             = "{{ .name }}"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
			proxy 			= artifactory_proxy.my-proxy.key
		}

	`, params)

	configUpdate := util.ExecuteTemplate("TestAccRemoteGoRepository", `
		resource "artifactory_remote_go_repository" "{{ .name }}" {
			key             = "{{ .name }}"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
			disable_proxy 	= true
		}

	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "proxy", "my-proxy"),
					resource.TestCheckResourceAttr(fqrn, "disable_proxy", "false"),
				),
			},
			{
				Config: configUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "proxy", ""),
					resource.TestCheckResourceAttr(fqrn, "disable_proxy", "true"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccRemoteDisableDefaultProxyConflictAttrGH739(t *testing.T) {
	_, fqrn, name := testutil.MkNames("tf-go-remote-", "artifactory_remote_go_repository")

	params := map[string]string{
		"name": name,
	}
	config := util.ExecuteTemplate("TestAccRemoteGoRepository", `
		resource "artifactory_remote_go_repository" "{{ .name }}" {
			key             = "{{ .name }}"
			repo_layout_ref = "go-default"
			url             = "https://gocenter.io"
			disable_proxy 	= true
			proxy 			= "my-proxy"
		}

	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*proxy cannot be set to 'when disable_proxy is set to 'true'.*"),
			},
		},
	})
}

func TestAccRemoteRepositoryWithProjectAttributesGH318(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	projectEnv := testutil.RandSelect("DEV", "PROD").(string)
	repoName := fmt.Sprintf("%s-pypi-remote", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_remote_pypi_repository")

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
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "", func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
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
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccRemoteRepositoryWithInvalidProjectKeyGH318(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-pypi-remote", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_remote_pypi_repository")

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
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "", func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      remoteRepositoryBasic,
				ExpectError: regexp.MustCompile(".*Attribute project_key must be 2 - 32 lowercase alphanumeric and hyphen.*"),
			},
		},
	})
}

func TestAccRemoteRepository_excludes_pattern_reset(t *testing.T) {
	_, fqrn, name := testutil.MkNames("generic-remote", "artifactory_remote_generic_repository")
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
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate("one", step1, map[string]interface{}{
					"name": name,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
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
					resource.TestCheckResourceAttr(fqrn, "url", "https://github.com"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", ""),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}
