package federated_test

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

// To make tests work add `ARTIFACTORY_URL_2=http://artifactory-2:8081` or `ARTIFACTORY_URL_2=http://host.docker.internal:9081`
func skipFederatedRepo() (bool, string) {
	if len(os.Getenv("ARTIFACTORY_URL_2")) > 0 {
		return false, "Env var `ARTIFACTORY_URL_2` is set. Executing test."
	}

	return true, "Env var `ARTIFACTORY_URL_2` is not set. Skipping test."
}

// In order to run this test, make sure your environment variables are set properly:
// https://github.com/jfrog/terraform-provider-artifactory/wiki/Testing#enable-acceptance-tests
func TestAccFederatedRepoWithMembers(t *testing.T) {
	if skip, reason := skipFederatedRepo(); skip {
		t.Skipf(reason)
	}

	name := fmt.Sprintf("federated-generic-%d-full", rand.Int())
	resourceType := "artifactory_federated_generic_repository"
	fqrn := fmt.Sprintf("%s.%s", resourceType, name)
	federatedMember1Url := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)
	// For local testing using Charles Proxy with cleanup_on_delete = true, instead of os.Getenv("ARTIFACTORY_URL_2") use "http://localhost.charlesproxy.com:9082"
	// internal hostname of RT2 is http://host.docker.internal:9082, works to create a member, but delete doesn't work with this host name, it needs "http://localhost.charlesproxy.com:9082"
	federatedMember2Url := fmt.Sprintf("%s/artifactory/%s", os.Getenv("ARTIFACTORY_URL_2"), name)

	params := map[string]interface{}{
		"resourceType": resourceType,
		"name":         name,
		"member1Url":   federatedMember1Url,
		"member2Url":   federatedMember2Url,
	}
	config := utilsdk.ExecuteTemplate("TestAccFederatedRepositoryConfigWithMembers", `
		resource "{{ .resourceType }}" "{{ .name }}" {
			key         = "{{ .name }}"
			description = "Test federated repo for {{ .name }}"
			notes       = "Test federated repo for {{ .name }}"

			member {
				url     = "{{ .member1Url }}"
				enabled = true
			}
		}
	`, params)
	updatedConfig := utilsdk.ExecuteTemplate("TestAccFederatedRepositoryConfigWithMembers", `
		resource "{{ .resourceType }}" "{{ .name }}" {
			key         = "{{ .name }}"
			description = "Test federated repo for {{ .name }}"
			notes       = "Test federated repo for {{ .name }}"

			member {
				url     = "{{ .member1Url }}"
				enabled = true
			}

			member {
				url     = "{{ .member2Url }}"
				enabled = true
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "member.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "member.0.url", federatedMember1Url),
					resource.TestCheckResourceAttr(fqrn, "member.0.enabled", "true"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "member.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "member.0.url", federatedMember2Url),
					resource.TestCheckResourceAttr(fqrn, "member.0.enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "member.1.url", federatedMember1Url),
					resource.TestCheckResourceAttr(fqrn, "member.1.enabled", "true"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

func genericTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	if skip, reason := skipFederatedRepo(); skip {
		t.Skipf(reason)
	}

	name := fmt.Sprintf("federated-%s-%d", repoType, rand.Int())
	resourceType := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
	fqrn := fmt.Sprintf("%s.%s", resourceType, name)
	xrayIndex := testutil.RandBool()
	proxyKey := fmt.Sprintf("test-proxy-%d", testutil.RandomInt())
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	params := map[string]interface{}{
		"resourceType": resourceType,
		"name":         name,
		"xrayIndex":    xrayIndex,
		"memberUrl":    federatedMemberUrl,
		"proxyKey":     proxyKey,
		"disableProxy": false,
	}

	repoTypeAdjusted := local.GetPackageType(repoType)

	// Default proxy will be assigned to the repository no matter what, and it's impossible to remove it by submitting an empty string or
	// removing the attribute. If `disable_proxy` is set to true, then both repo and default proxies are removed and not returned in the
	// GET body.
	config := utilsdk.ExecuteTemplate("TestAccFederatedRepositoryConfig", `
		resource "artifactory_proxy" "{{ .proxyKey }}" {
			key  = "{{ .proxyKey }}"
			host = "http://tempurl.org"
			port = 8000
		}

		resource "{{ .resourceType }}" "{{ .name }}" {
			key           = "{{ .name }}"
			description   = "Test federated repo for {{ .name }}"
			notes         = "Test federated repo for {{ .name }}"
			xray_index    = {{ .xrayIndex }}
			proxy         = artifactory_proxy.{{ .proxyKey }}.key
			disable_proxy = {{ .disableProxy }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`, params)

	updatedParams := map[string]interface{}{
		"resourceType": resourceType,
		"name":         name,
		"xrayIndex":    !xrayIndex,
		"memberUrl":    federatedMemberUrl,
		"proxyKey":     proxyKey,
		"disableProxy": true,
	}

	updatedConfig := utilsdk.ExecuteTemplate("TestAccFederatedRepositoryConfig", `
		resource "artifactory_proxy" "{{ .proxyKey }}" {
			key  = "{{ .proxyKey }}"
			host = "http://tempurl.org"
			port = 8000
		}

		resource "{{ .resourceType }}" "{{ .name }}" {
			key           = "{{ .name }}"
			description   = "Test federated repo for {{ .name }}"
			notes         = "Test federated repo for {{ .name }}"
			xray_index    = {{ .xrayIndex }}
			disable_proxy = {{ .disableProxy }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`, updatedParams)

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", repoTypeAdjusted),
					resource.TestCheckResourceAttr(fqrn, "description", fmt.Sprintf("Test federated repo for %s", name)),
					resource.TestCheckResourceAttr(fqrn, "notes", fmt.Sprintf("Test federated repo for %s", name)),
					resource.TestCheckResourceAttr(fqrn, "xray_index", fmt.Sprintf("%t", xrayIndex)),
					resource.TestCheckResourceAttr(fqrn, "proxy", proxyKey),
					resource.TestCheckResourceAttr(fqrn, "disable_proxy", fmt.Sprintf("%t", params["disableProxy"].(bool))),

					resource.TestCheckResourceAttr(fqrn, "member.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "member.0.url", federatedMemberUrl),
					resource.TestCheckResourceAttr(fqrn, "member.0.enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", repoType)(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", repoTypeAdjusted),
					resource.TestCheckResourceAttr(fqrn, "description", fmt.Sprintf("Test federated repo for %s", name)),
					resource.TestCheckResourceAttr(fqrn, "notes", fmt.Sprintf("Test federated repo for %s", name)),
					resource.TestCheckResourceAttr(fqrn, "xray_index", fmt.Sprintf("%t", !xrayIndex)),
					resource.TestCheckResourceAttr(fqrn, "proxy", ""),
					resource.TestCheckResourceAttr(fqrn, "disable_proxy", fmt.Sprintf("%t", updatedParams["disableProxy"].(bool))),

					resource.TestCheckResourceAttr(fqrn, "member.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "member.0.url", federatedMemberUrl),
					resource.TestCheckResourceAttr(fqrn, "member.0.enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", repoType)(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	}
}

func TestAccFederatedRepoGenericTypes(t *testing.T) {
	for _, packageType := range federated.PackageTypesLikeGeneric {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(genericTestCase(packageType, t))
		})
	}
}

func TestAccFederatedRepo_DisableDefaultProxyConflictAttr(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-go-remote-", "artifactory_federated_go_repository")
	memberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	params := map[string]string{
		"name":      name,
		"memberUrl": memberUrl,
	}
	config := utilsdk.ExecuteTemplate("TestAccFederatedGoRepository", `
		resource "artifactory_federated_go_repository" "{{ .name }}" {
			key             = "{{ .name }}"
			repo_layout_ref = "go-default"
			proxy 			= "my-proxy"
			disable_proxy 	= true

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}

	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*if `disable_proxy` is set to `true`, `proxy` can't be set"),
			},
		},
	})
}

func TestAccFederatedRepoWithProjectAttributesGH318(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	projectEnv := testutil.RandSelect("DEV", "PROD").(string)
	repoName := fmt.Sprintf("%s-generic-federated", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_federated_generic_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
		"projectEnv": projectEnv,
		"memberUrl":  federatedMemberUrl,
	}
	federatedRepositoryConfig := utilsdk.ExecuteTemplate("TestAccFederatedRepositoryConfig", `
		resource "artifactory_federated_generic_repository" "{{ .name }}" {
			key                  = "{{ .name }}"
			project_key          = "{{ .projectKey }}"
	 		project_environments = ["{{ .projectEnv }}"]

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: acctest.VerifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "member.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "member.0.url", federatedMemberUrl),
					resource.TestCheckResourceAttr(fqrn, "member.0.enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.0", projectEnv),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

func TestAccFederatedRepositoryWithInvalidProjectKeyGH318(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-federated", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_federated_generic_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
		"memberUrl":  federatedMemberUrl,
	}
	federatedRepositoryConfig := utilsdk.ExecuteTemplate("TestAccFederatedRepositoryConfig", `
		resource "artifactory_federated_generic_repository" "{{ .name }}" {
			key         = "{{ .name }}"
		 	project_key = "invalid-project-key-too-long-really-long"

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: acctest.VerifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		Steps: []resource.TestStep{
			{
				Config:      federatedRepositoryConfig,
				ExpectError: regexp.MustCompile(".*project_key must be 2 - 32 lowercase alphanumeric and hyphen characters"),
			},
		},
	})
}

func TestAccFederatedAlpineRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("alpine-federated", "artifactory_federated_alpine_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	federatedRepositoryBasic := utilsdk.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA2ymVc24BoaZb0ElXoI3X4zUKJGZEetR6F4yT1tJhkPDg7UTm
iNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbqFUaXPgud8VabfHl0imXvN746zmpj
YEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEppj5yN0tVWDnqjOJjR7EpxMSdP3TSd
6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQ
FmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDMX8KGs7e9ZgjANkT5PnipLOaeLJU4
H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJDQIDAQABAoIBAFb7wyhEIfuhhlE9
uryrb2LrGzJlMIq7qBWOouKhLz4SjIM/VGw+c76VkjZGoSU+LeLj+D0W1ie0u2Cw
gJM8aW22TbK/c5lksWOO5PVFDdPG+ZoRWY3MLGlhlL5E4UhMPgJyy/eeiRjZ3CZM
pja+UfVAwn1KVNR8UinVZYPt680AvEd1FGxoNLxemIPNug46nNqp6Al86Bn+BnkN
GXpwyooXVSfo4k+pnFBFIXUdA1dUEXQSVb1AxlTo6g/Ok/+8Gfq8idCdu+5fcZI2
1eLeC+FAa92rr1SFX/UWeB4cMyuAqwuxbFFIplT22SaUSsNuOUSH4E00nbP/AuCb
1BqrLmECgYEA7tQKfyF9aiXTsOMdOnSAa5OyEaCfsFtcmLd4ykVrwN8O36NoX005
VbGuqo87fwIXQIKHU+kOEs/TmaQ8bNcbCD/SfWGTtOnHG4qfIsepJuoMwbQHRhGF
JeoXh5yEUKg5pcDBY8PENEtEVKmFuL4bPOdn+9FNLGsjftvXpmGWbGUCgYEA6uuQ
7kzO6WD88OsxdJzlJM11hg2SaSBCh3+5tnOhF1ULOUt4tdYXzh3QI6BPX7tkArYf
XteVfWoWqn6T7LtCjFm350BqVpPhqfLKnt6fYf1yotsj/cyZXlXquRbxbgakB0n0
4PrsZaube0TPPVeirzNyOVHyFc+iW+F+IUYD+4kCgYEApDEjBkP/9PoMj4+UiJuP
rmXcBkJnhtdI0bVRVb5kVjUEBLxTBTISONfvPVM7lBXb5n3Wi9mt00EOOJKw+CLq
csFt9MUgxz/xov2qaj7aC+bc3k7msUVaRLardpAkZ09AUrQyQGRWf50/XPUu+dO4
5iYxVu6OH/uIa664k6qDwAECgYEAslS8oomgEL3VhbWkx1dLA5MMggTPfgpFNsMY
4Y4JXcLrUEUgjzjEvW0YUdMiLhP8qapDSiXxj1D3f9myxWSp8g0xc9UMZEjCZ9at
RcjNyP8zBLnCKqokSt6B3puyDsnvvrC/ugIBbnTFBOCJSZG7J7CwJx8z3KbQI1ub
+fpCj7ECgYAd69soLEybUGMjsdI+OijIGoUTUoZGXJm+0VpBt4QJCe7AMnYPfYzA
JnEmN4D7HLTKUBklQnb/FhP/RuiT2bSAd1l+PNeuU7gYROCBBonzxXQ1wEbNrSYA
iyoc9g/kvV8HTW8361xEhu7wmuSEEx1gQ/7sdhTkgrTncc8hxVRxuA==
-----END RSA PRIVATE KEY-----
EOF
			public_key = <<EOF
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2ymVc24BoaZb0ElXoI3X
4zUKJGZEetR6F4yT1tJhkPDg7UTmiNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbq
FUaXPgud8VabfHl0imXvN746zmpjYEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEpp
j5yN0tVWDnqjOJjR7EpxMSdP3TSd6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof
+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQFmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDM
X8KGs7e9ZgjANkT5PnipLOaeLJU4H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJ
DQIDAQAB
-----END PUBLIC KEY-----
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_federated_alpine_repository" "{{ .repo_name }}" {
			key 	            = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}

			depends_on = [artifactory_keypair.{{ .kp_name }}]
		}
	`, map[string]interface{}{
		"kp_id":     kpId,
		"kp_name":   kpName,
		"repo_name": name,
		"memberUrl": federatedMemberUrl,
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
			acctest.VerifyDeleted(kpFqrn, security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "alpine"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "alpine")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

func TestAccFederatedCargoRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("cargo-federated", "artifactory_federated_cargo_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)
	anonAccess := testutil.RandBool()
	enabledSparseIndex := testutil.RandBool()

	params := map[string]interface{}{
		"anonymous_access":    anonAccess,
		"enable_sparse_index": enabledSparseIndex,
		"name":                name,
		"memberUrl":           federatedMemberUrl,
	}

	template := `
		resource "artifactory_federated_cargo_repository" "{{ .name }}" {
			key                 = "{{ .name }}"
			anonymous_access    = {{ .anonymous_access }}
			enable_sparse_index = {{ .enable_sparse_index }}
			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`
	federatedRepositoryBasic := utilsdk.ExecuteTemplate("TestAccFederatedCargoRepository", template, params)
	federatedRepositoryUpdated := utilsdk.ExecuteTemplate(
		"TestAccFederatedCargoRepository",
		template,
		map[string]interface{}{
			"anonymous_access":    !anonAccess,
			"enable_sparse_index": !enabledSparseIndex,
			"name":                name,
			"memberUrl":           federatedMemberUrl,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "anonymous_access", fmt.Sprintf("%t", anonAccess)),
					resource.TestCheckResourceAttr(fqrn, "enable_sparse_index", fmt.Sprintf("%t", enabledSparseIndex)),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "cargo")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "anonymous_access", fmt.Sprintf("%t", !anonAccess)),
					resource.TestCheckResourceAttr(fqrn, "enable_sparse_index", fmt.Sprintf("%t", !enabledSparseIndex)),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "cargo")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

func TestAccFederatedConanRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("conan-federated", "artifactory_federated_conan_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)
	forceConanAuthentication := testutil.RandBool()

	params := map[string]interface{}{
		"force_conan_authentication": forceConanAuthentication,
		"name":                       name,
		"memberUrl":                  federatedMemberUrl,
	}

	template := `
		resource "artifactory_federated_conan_repository" "{{ .name }}" {
			key                        = "{{ .name }}"
			force_conan_authentication = {{ .force_conan_authentication }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`
	federatedRepositoryBasic := utilsdk.ExecuteTemplate("TestAccFederatedConanRepository", template, params)

	federatedRepositoryUpdated := utilsdk.ExecuteTemplate(
		"TestAccFederatedCargoRepository",
		template,
		map[string]interface{}{
			"force_conan_authentication": !forceConanAuthentication,
			"name":                       name,
			"memberUrl":                  federatedMemberUrl,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "force_conan_authentication", fmt.Sprintf("%t", forceConanAuthentication)),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "conan")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "force_conan_authentication", fmt.Sprintf("%t", !forceConanAuthentication)),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "conan")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

func TestAccFederatedDebianRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("debian-federated", "artifactory_federated_debian_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")

	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
-----BEGIN PGP PRIVATE KEY BLOCK-----
lIYEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib/+BwMCFjb4odY28+n0NWj7KZ53BkA0qzzqT9IpIfsW/tLNPTxYEFrDVbcF
1CuiAgAhyUfBEr9HQaMJBLfIIvo/B3nlWvwWHkiQFuWpsnJ2pj8F8LQqQ2hyaXN0
aWFuIEJvbmdpb3JubyA8Y2hyaXN0aWFuYkBqZnJvZy5jb20+iJoEExYKAEIWIQSS
w8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbAwUJA8JnAAULCQgHAgMiAgEGFQoJ
CAsCBBYCAwECHgcCF4AACgkQwL80hJIR2yRQDgD/X1t/hW9+uXdSY59FOClhQw/t
AzTYjDW+KLKadYJ3RAIBALD53rj7EnrXsSqv9Vqj3mJ7O38eXu50P57tD8ErpHMD
nIsEYYU7tRIKKwYBBAGXVQEFAQEHQCfT+jXHVkslGAJqVafoeWO8Nwz/oPPzNDJb
EOASsMRcAwEIB/4HAwK+Wi8OaidLuvQ6yknLUspoRL8KJlQu0JkfLxj6Wl6GrRtf
MdUBxaGUQX5UzMIqyYstgHKz2kBYvrJijWdOkkRuL82FySSh4yi/97FBikOBiHgE
GBYKACAWIQSSw8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbDAAKCRDAvzSEkhHb
JNR/AQCQjGWljmP8pYj6ohP8bOwVB4VE5qxjdfWQvBCUA0LFwgEAxLGVeT88pw3+
x7Cwd7SsuxlIOOCIJssFnUhA9Qsq2wE=
=qCzy
-----END PGP PRIVATE KEY BLOCK-----
EOF
			public_key = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----
mDMEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib+0KkNocmlzdGlhbiBCb25naW9ybm8gPGNocmlzdGlhbmJAamZyb2cuY29t
PoiaBBMWCgBCFiEEksPI7fvaXVQtxrbOwL80hJIR2yQFAmGFO7UCGwMFCQPCZwAF
CwkIBwIDIgIBBhUKCQgLAgQWAgMBAh4HAheAAAoJEMC/NISSEdskUA4A/19bf4Vv
frl3UmOfRTgpYUMP7QM02Iw1viiymnWCd0QCAQCw+d64+xJ617Eqr/Vao95iezt/
Hl7udD+e7Q/BK6RzA7g4BGGFO7USCisGAQQBl1UBBQEBB0An0/o1x1ZLJRgCalWn
6HljvDcM/6Dz8zQyWxDgErDEXAMBCAeIeAQYFgoAIBYhBJLDyO372l1ULca2zsC/
NISSEdskBQJhhTu1AhsMAAoJEMC/NISSEdsk1H8BAJCMZaWOY/yliPqiE/xs7BUH
hUTmrGN19ZC8EJQDQsXCAQDEsZV5PzynDf7HsLB3tKy7GUg44IgmywWdSED1Cyrb
AQ==
=2kMe
-----END PGP PUBLIC KEY BLOCK-----
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_keypair" "{{ .kp_name2 }}" {
			pair_name  = "{{ .kp_name2 }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id2 }}"
			private_key = <<EOF
-----BEGIN PGP PRIVATE KEY BLOCK-----
lIYEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib/+BwMCFjb4odY28+n0NWj7KZ53BkA0qzzqT9IpIfsW/tLNPTxYEFrDVbcF
1CuiAgAhyUfBEr9HQaMJBLfIIvo/B3nlWvwWHkiQFuWpsnJ2pj8F8LQqQ2hyaXN0
aWFuIEJvbmdpb3JubyA8Y2hyaXN0aWFuYkBqZnJvZy5jb20+iJoEExYKAEIWIQSS
w8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbAwUJA8JnAAULCQgHAgMiAgEGFQoJ
CAsCBBYCAwECHgcCF4AACgkQwL80hJIR2yRQDgD/X1t/hW9+uXdSY59FOClhQw/t
AzTYjDW+KLKadYJ3RAIBALD53rj7EnrXsSqv9Vqj3mJ7O38eXu50P57tD8ErpHMD
nIsEYYU7tRIKKwYBBAGXVQEFAQEHQCfT+jXHVkslGAJqVafoeWO8Nwz/oPPzNDJb
EOASsMRcAwEIB/4HAwK+Wi8OaidLuvQ6yknLUspoRL8KJlQu0JkfLxj6Wl6GrRtf
MdUBxaGUQX5UzMIqyYstgHKz2kBYvrJijWdOkkRuL82FySSh4yi/97FBikOBiHgE
GBYKACAWIQSSw8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbDAAKCRDAvzSEkhHb
JNR/AQCQjGWljmP8pYj6ohP8bOwVB4VE5qxjdfWQvBCUA0LFwgEAxLGVeT88pw3+
x7Cwd7SsuxlIOOCIJssFnUhA9Qsq2wE=
=qCzy
-----END PGP PRIVATE KEY BLOCK-----
EOF
			public_key = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----
mDMEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib+0KkNocmlzdGlhbiBCb25naW9ybm8gPGNocmlzdGlhbmJAamZyb2cuY29t
PoiaBBMWCgBCFiEEksPI7fvaXVQtxrbOwL80hJIR2yQFAmGFO7UCGwMFCQPCZwAF
CwkIBwIDIgIBBhUKCQgLAgQWAgMBAh4HAheAAAoJEMC/NISSEdskUA4A/19bf4Vv
frl3UmOfRTgpYUMP7QM02Iw1viiymnWCd0QCAQCw+d64+xJ617Eqr/Vao95iezt/
Hl7udD+e7Q/BK6RzA7g4BGGFO7USCisGAQQBl1UBBQEBB0An0/o1x1ZLJRgCalWn
6HljvDcM/6Dz8zQyWxDgErDEXAMBCAeIeAQYFgoAIBYhBJLDyO372l1ULca2zsC/
NISSEdskBQJhhTu1AhsMAAoJEMC/NISSEdsk1H8BAJCMZaWOY/yliPqiE/xs7BUH
hUTmrGN19ZC8EJQDQsXCAQDEsZV5PzynDf7HsLB3tKy7GUg44IgmywWdSED1Cyrb
AQ==
=2kMe
-----END PGP PUBLIC KEY BLOCK-----
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_federated_debian_repository" "{{ .repo_name }}" {
			key 	                  = "{{ .repo_name }}"
			primary_keypair_ref       = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref     = artifactory_keypair.{{ .kp_name2 }}.pair_name
			index_compression_formats = ["bz2","lzma","xz"]
			trivial_layout            = {{ .trivialLayout }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}

			depends_on = [
				artifactory_keypair.{{ .kp_name }},
				artifactory_keypair.{{ .kp_name2 }},
			]
		}
	`

	federatedRepositoryBasic := utilsdk.ExecuteTemplate("keypair", template, map[string]interface{}{
		"kp_id":         kpId,
		"kp_name":       kpName,
		"kp_id2":        kpId2,
		"kp_name2":      kpName2,
		"repo_name":     name,
		"trivialLayout": true,
		"memberUrl":     federatedMemberUrl,
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	federatedRepositoryUpdated := utilsdk.ExecuteTemplate("keypair", template, map[string]interface{}{
		"kp_id":         kpId,
		"kp_name":       kpName,
		"kp_id2":        kpId2,
		"kp_name2":      kpName2,
		"repo_name":     name,
		"trivialLayout": false,
		"memberUrl":     federatedMemberUrl,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
			acctest.VerifyDeleted(kpFqrn, security.VerifyKeyPair),
			acctest.VerifyDeleted(kpFqrn2, security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "debian"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "trivial_layout", "true"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.0", "bz2"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.1", "lzma"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.2", "xz"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "debian")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "debian"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "trivial_layout", "false"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.0", "bz2"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.1", "lzma"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.2", "xz"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "debian")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

func TestAccFederatedDockerV2Repository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("docker-federated", "artifactory_federated_docker_v2_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_federated_docker_v2_repository" "{{ .name }}" {
			key 	              = "{{ .name }}"
			tag_retention         = {{ .retention }}
			max_unique_tags       = {{ .max_tags }}
			block_pushing_schema1 = {{ .block }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`

	params := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryBasic := utilsdk.ExecuteTemplate("TestAccFederatedDockerRepository", template, params)

	updated := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryUpdated := utilsdk.ExecuteTemplate("TestAccFederatedDockerRepository", template, updated)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", params["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "docker")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", updated["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", updated["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", updated["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "docker")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

// TestAccFederatedDockerRepository tests for backward compatibility
func TestAccFederatedDockerRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("docker-federated", "artifactory_federated_docker_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_federated_docker_repository" "{{ .name }}" {
			key 	              = "{{ .name }}"
			tag_retention         = {{ .retention }}
			max_unique_tags       = {{ .max_tags }}
			block_pushing_schema1 = {{ .block }}

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`

	params := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryBasic := utilsdk.ExecuteTemplate("TestAccFederatedDockerRepository", template, params)

	updated := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryUpdated := utilsdk.ExecuteTemplate("TestAccFederatedDockerRepository", template, updated)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", params["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "docker")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", updated["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", updated["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", updated["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "docker")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

func TestAccFederatedDockerV1Repository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("docker-federated", "artifactory_federated_docker_v1_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_federated_docker_v1_repository" "{{ .name }}" {
			key = "{{ .name }}"

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`

	params := map[string]interface{}{
		"name":      name,
		"memberUrl": federatedMemberUrl,
	}
	federatedRepositoryBasic := utilsdk.ExecuteTemplate("TestAccFederatedDockerRepository", template, params)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", "false"),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", "1"),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", "0"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "docker")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

func TestAccFederatedNugetRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("nuget-federated", "artifactory_federated_nuget_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_federated_nuget_repository" "{{ .name }}" {
			key                        = "{{ .name }}"
			max_unique_snapshots       = {{ .max_unique_snapshots }}
			force_nuget_authentication = {{ .force_nuget_authentication }}
			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`

	params := map[string]interface{}{
		"force_nuget_authentication": testutil.RandBool(),
		"max_unique_snapshots":       testutil.RandSelect(0, 5, 10),
		"name":                       name,
		"memberUrl":                  federatedMemberUrl,
	}
	federatedRepositoryBasic := utilsdk.ExecuteTemplate("TestAccLocalNugetRepository", template, params)

	updates := map[string]interface{}{
		"force_nuget_authentication": testutil.RandBool(),
		"max_unique_snapshots":       testutil.RandSelect(0, 5, 10),
		"name":                       name,
		"memberUrl":                  federatedMemberUrl,
	}
	federatedRepositoryUpdated := utilsdk.ExecuteTemplate("TestAccLocalNugetRepository", template, updates)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", params["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", fmt.Sprintf("%t", params["force_nuget_authentication"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "nuget")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", updates["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", fmt.Sprintf("%t", updates["force_nuget_authentication"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "nuget")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

var commonJavaParams = map[string]interface{}{
	"name":                            "",
	"checksum_policy_type":            testutil.RandSelect("client-checksums", "server-generated-checksums"),
	"snapshot_version_behavior":       testutil.RandSelect("unique", "non-unique", "deployer"),
	"max_unique_snapshots":            testutil.RandSelect(0, 5, 10),
	"handle_releases":                 true,
	"handle_snapshots":                true,
	"suppress_pom_consistency_checks": false,
}

const federatedJavaRepositoryBasic = `
	resource "{{ .resource_name }}" "{{ .name }}" {
		key                             = "{{ .name }}"
		checksum_policy_type            = "{{ .checksum_policy_type }}"
		snapshot_version_behavior       = "{{ .snapshot_version_behavior }}"
		max_unique_snapshots            = {{ .max_unique_snapshots }}
		handle_releases                 = {{ .handle_releases }}
		handle_snapshots                = {{ .handle_snapshots }}
		suppress_pom_consistency_checks = {{ .suppress_pom_consistency_checks }}
		member {
			url     = "{{ .memberUrl }}"
			enabled = true
		}
	}
`

func TestAccFederatedMavenRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("maven-federated", "artifactory_federated_maven_repository")

	repoLayoutRef := func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "maven")(); return r.(string) }()
	tempStruct := utilsdk.MergeMaps(commonJavaParams)
	tempStruct["name"] = name
	tempStruct["resource_name"] = strings.Split(fqrn, ".")[0]
	tempStruct["suppress_pom_consistency_checks"] = false
	tempStruct["memberUrl"] = fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	updatedStruct := tempStruct
	updatedStruct["snapshot_version_behavior"] = "non-unique"
	updatedStruct["handle_releases"] = false
	updatedStruct["handle_snapshots"] = false
	updatedStruct["suppress_pom_consistency_checks"] = true

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: utilsdk.ExecuteTemplate(fqrn, federatedJavaRepositoryBasic, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "checksum_policy_type", fmt.Sprintf("%s", tempStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(fqrn, "snapshot_version_behavior", fmt.Sprintf("%s", tempStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", tempStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%v", tempStruct["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%v", tempStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", tempStruct["suppress_pom_consistency_checks"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", repoLayoutRef), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: utilsdk.ExecuteTemplate(fqrn, federatedJavaRepositoryBasic, updatedStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "checksum_policy_type", fmt.Sprintf("%s", updatedStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(fqrn, "snapshot_version_behavior", fmt.Sprintf("%s", updatedStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", updatedStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%v", updatedStruct["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%v", updatedStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", updatedStruct["suppress_pom_consistency_checks"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", repoLayoutRef), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

func makeFederatedGradleLikeRepoTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("%s-federated", repoType)
	resourceName := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
	_, fqrn, name := testutil.MkNames(name, resourceName)
	tempStruct := utilsdk.MergeMaps(commonJavaParams)

	tempStruct["name"] = name
	tempStruct["resource_name"] = strings.Split(fqrn, ".")[0]
	tempStruct["suppress_pom_consistency_checks"] = true
	tempStruct["memberUrl"] = fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	updatedStruct := tempStruct
	updatedStruct["snapshot_version_behavior"] = "non-unique"
	updatedStruct["handle_releases"] = false
	updatedStruct["handle_snapshots"] = false
	updatedStruct["suppress_pom_consistency_checks"] = true

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: utilsdk.ExecuteTemplate(fqrn, federatedJavaRepositoryBasic, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "checksum_policy_type", fmt.Sprintf("%s", tempStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(fqrn, "snapshot_version_behavior", fmt.Sprintf("%s", tempStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", tempStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%v", tempStruct["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%v", tempStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", tempStruct["suppress_pom_consistency_checks"])),
				),
			},
			{
				Config: utilsdk.ExecuteTemplate(fqrn, federatedJavaRepositoryBasic, updatedStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "checksum_policy_type", fmt.Sprintf("%s", updatedStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(fqrn, "snapshot_version_behavior", fmt.Sprintf("%s", updatedStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", updatedStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%v", updatedStruct["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%v", updatedStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", updatedStruct["suppress_pom_consistency_checks"])),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	}
}

func TestAccFederatedAllGradleLikePackageTypes(t *testing.T) {
	for _, packageType := range repository.GradleLikePackageTypes {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(makeFederatedGradleLikeRepoTestCase(packageType, t))
		})
	}
}

func TestAccFederatedRpmRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("rpm-federated", "artifactory_federated_rpm_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")

	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	template := `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
-----BEGIN PGP PRIVATE KEY BLOCK-----
lIYEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib/+BwMCFjb4odY28+n0NWj7KZ53BkA0qzzqT9IpIfsW/tLNPTxYEFrDVbcF
1CuiAgAhyUfBEr9HQaMJBLfIIvo/B3nlWvwWHkiQFuWpsnJ2pj8F8LQqQ2hyaXN0
aWFuIEJvbmdpb3JubyA8Y2hyaXN0aWFuYkBqZnJvZy5jb20+iJoEExYKAEIWIQSS
w8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbAwUJA8JnAAULCQgHAgMiAgEGFQoJ
CAsCBBYCAwECHgcCF4AACgkQwL80hJIR2yRQDgD/X1t/hW9+uXdSY59FOClhQw/t
AzTYjDW+KLKadYJ3RAIBALD53rj7EnrXsSqv9Vqj3mJ7O38eXu50P57tD8ErpHMD
nIsEYYU7tRIKKwYBBAGXVQEFAQEHQCfT+jXHVkslGAJqVafoeWO8Nwz/oPPzNDJb
EOASsMRcAwEIB/4HAwK+Wi8OaidLuvQ6yknLUspoRL8KJlQu0JkfLxj6Wl6GrRtf
MdUBxaGUQX5UzMIqyYstgHKz2kBYvrJijWdOkkRuL82FySSh4yi/97FBikOBiHgE
GBYKACAWIQSSw8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbDAAKCRDAvzSEkhHb
JNR/AQCQjGWljmP8pYj6ohP8bOwVB4VE5qxjdfWQvBCUA0LFwgEAxLGVeT88pw3+
x7Cwd7SsuxlIOOCIJssFnUhA9Qsq2wE=
=qCzy
-----END PGP PRIVATE KEY BLOCK-----
EOF
			public_key = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----
mDMEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib+0KkNocmlzdGlhbiBCb25naW9ybm8gPGNocmlzdGlhbmJAamZyb2cuY29t
PoiaBBMWCgBCFiEEksPI7fvaXVQtxrbOwL80hJIR2yQFAmGFO7UCGwMFCQPCZwAF
CwkIBwIDIgIBBhUKCQgLAgQWAgMBAh4HAheAAAoJEMC/NISSEdskUA4A/19bf4Vv
frl3UmOfRTgpYUMP7QM02Iw1viiymnWCd0QCAQCw+d64+xJ617Eqr/Vao95iezt/
Hl7udD+e7Q/BK6RzA7g4BGGFO7USCisGAQQBl1UBBQEBB0An0/o1x1ZLJRgCalWn
6HljvDcM/6Dz8zQyWxDgErDEXAMBCAeIeAQYFgoAIBYhBJLDyO372l1ULca2zsC/
NISSEdskBQJhhTu1AhsMAAoJEMC/NISSEdsk1H8BAJCMZaWOY/yliPqiE/xs7BUH
hUTmrGN19ZC8EJQDQsXCAQDEsZV5PzynDf7HsLB3tKy7GUg44IgmywWdSED1Cyrb
AQ==
=2kMe
-----END PGP PUBLIC KEY BLOCK-----
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_keypair" "{{ .kp_name2 }}" {
			pair_name  = "{{ .kp_name2 }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id2 }}"
			private_key = <<EOF
-----BEGIN PGP PRIVATE KEY BLOCK-----
lIYEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib/+BwMCFjb4odY28+n0NWj7KZ53BkA0qzzqT9IpIfsW/tLNPTxYEFrDVbcF
1CuiAgAhyUfBEr9HQaMJBLfIIvo/B3nlWvwWHkiQFuWpsnJ2pj8F8LQqQ2hyaXN0
aWFuIEJvbmdpb3JubyA8Y2hyaXN0aWFuYkBqZnJvZy5jb20+iJoEExYKAEIWIQSS
w8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbAwUJA8JnAAULCQgHAgMiAgEGFQoJ
CAsCBBYCAwECHgcCF4AACgkQwL80hJIR2yRQDgD/X1t/hW9+uXdSY59FOClhQw/t
AzTYjDW+KLKadYJ3RAIBALD53rj7EnrXsSqv9Vqj3mJ7O38eXu50P57tD8ErpHMD
nIsEYYU7tRIKKwYBBAGXVQEFAQEHQCfT+jXHVkslGAJqVafoeWO8Nwz/oPPzNDJb
EOASsMRcAwEIB/4HAwK+Wi8OaidLuvQ6yknLUspoRL8KJlQu0JkfLxj6Wl6GrRtf
MdUBxaGUQX5UzMIqyYstgHKz2kBYvrJijWdOkkRuL82FySSh4yi/97FBikOBiHgE
GBYKACAWIQSSw8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbDAAKCRDAvzSEkhHb
JNR/AQCQjGWljmP8pYj6ohP8bOwVB4VE5qxjdfWQvBCUA0LFwgEAxLGVeT88pw3+
x7Cwd7SsuxlIOOCIJssFnUhA9Qsq2wE=
=qCzy
-----END PGP PRIVATE KEY BLOCK-----
EOF
			public_key = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----
mDMEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
jn7zib+0KkNocmlzdGlhbiBCb25naW9ybm8gPGNocmlzdGlhbmJAamZyb2cuY29t
PoiaBBMWCgBCFiEEksPI7fvaXVQtxrbOwL80hJIR2yQFAmGFO7UCGwMFCQPCZwAF
CwkIBwIDIgIBBhUKCQgLAgQWAgMBAh4HAheAAAoJEMC/NISSEdskUA4A/19bf4Vv
frl3UmOfRTgpYUMP7QM02Iw1viiymnWCd0QCAQCw+d64+xJ617Eqr/Vao95iezt/
Hl7udD+e7Q/BK6RzA7g4BGGFO7USCisGAQQBl1UBBQEBB0An0/o1x1ZLJRgCalWn
6HljvDcM/6Dz8zQyWxDgErDEXAMBCAeIeAQYFgoAIBYhBJLDyO372l1ULca2zsC/
NISSEdskBQJhhTu1AhsMAAoJEMC/NISSEdsk1H8BAJCMZaWOY/yliPqiE/xs7BUH
hUTmrGN19ZC8EJQDQsXCAQDEsZV5PzynDf7HsLB3tKy7GUg44IgmywWdSED1Cyrb
AQ==
=2kMe
-----END PGP PUBLIC KEY BLOCK-----
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_federated_rpm_repository" "{{ .repo_name }}" {
			key 	                   = "{{ .repo_name }}"
			primary_keypair_ref        = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref      = artifactory_keypair.{{ .kp_name2 }}.pair_name
			yum_root_depth             = {{ .yum_root_depth }}
			enable_file_lists_indexing = {{ .enable_file_lists_indexing }}
			calculate_yum_metadata     = true

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}

			depends_on = [
				artifactory_keypair.{{ .kp_name }},
				artifactory_keypair.{{ .kp_name2 }},
			]
		}
	`

	federatedRepositoryBasic := utilsdk.ExecuteTemplate("keypair", template, map[string]interface{}{
		"kp_id":                      kpId,
		"kp_name":                    kpName,
		"kp_id2":                     kpId2,
		"kp_name2":                   kpName2,
		"repo_name":                  name,
		"yum_root_depth":             1,
		"enable_file_lists_indexing": true,
		"memberUrl":                  federatedMemberUrl,
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	federatedRepositoryUpdated := utilsdk.ExecuteTemplate("keypair", template, map[string]interface{}{
		"kp_id":                      kpId,
		"kp_name":                    kpName,
		"kp_id2":                     kpId2,
		"kp_name2":                   kpName2,
		"repo_name":                  name,
		"yum_root_depth":             2,
		"enable_file_lists_indexing": false,
		"memberUrl":                  federatedMemberUrl,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
			acctest.VerifyDeleted(kpFqrn, security.VerifyKeyPair),
			acctest.VerifyDeleted(kpFqrn2, security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "rpm"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "enable_file_lists_indexing", "true"),
					resource.TestCheckResourceAttr(fqrn, "calculate_yum_metadata", "true"),
					resource.TestCheckResourceAttr(fqrn, "yum_root_depth", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "rpm")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: federatedRepositoryUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "rpm"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "enable_file_lists_indexing", "false"),
					resource.TestCheckResourceAttr(fqrn, "calculate_yum_metadata", "true"),
					resource.TestCheckResourceAttr(fqrn, "yum_root_depth", "2"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("federated", "rpm")(); return r.(string) }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	})
}

func makeFederatedTerraformRepoTestCase(registryType string, t *testing.T) (*testing.T, resource.TestCase) {
	resourceName := fmt.Sprintf("terraform-module-%s", registryType)
	resourceType := fmt.Sprintf("artifactory_federated_terraform_%s_repository", registryType)
	_, fqrn, name := testutil.MkNames(resourceName, resourceType)
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", acctest.GetArtifactoryUrl(t), name)

	params := map[string]interface{}{
		"registryType": registryType,
		"name":         name,
		"memberUrl":    federatedMemberUrl,
	}

	template := `
		resource "artifactory_federated_terraform_{{ .registryType }}_repository" "{{ .name }}" {
			key = "{{ .name }}"

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`
	federatedRepositoryBasic := utilsdk.ExecuteTemplate("TestAccFederatedTerraformRepository", template, params)

	return t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "terraform"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("federated", "terraform_"+registryType)()
						return r.(string)
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"cleanup_on_delete"},
			},
		},
	}
}

func TestAccFederatedTerraformRepositories(t *testing.T) {
	for _, registryType := range []string{"module", "provider"} {
		t.Run(registryType, func(t *testing.T) {
			resource.Test(makeFederatedTerraformRepoTestCase(registryType, t))
		})
	}
}
