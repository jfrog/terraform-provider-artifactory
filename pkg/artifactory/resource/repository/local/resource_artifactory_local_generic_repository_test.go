package local_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalGenericRepository_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-generic-local", "artifactory_local_generic_repository")

	config := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		}
	`, map[string]interface{}{
		"name": name,
	})

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "12.6.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "id", name),
					resource.TestCheckResourceAttr(fqrn, "key", name),
				),
			},
			{
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

func TestAccLocalGenericRepository_withProjectAttributes(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	projectEnv := testutil.RandSelect("DEV", "PROD").(string)
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
		"projectEnv": projectEnv,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = "{{ .projectKey }}"
	 	  project_environments = ["{{ .projectEnv }}"]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
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

func TestAccLocalGenericRepository_WithInvalidProjectKey(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = "invalid-project-key-too-long-really-long"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		Steps: []resource.TestStep{
			{
				Config:      localRepositoryBasic,
				ExpectError: regexp.MustCompile(".*Attribute project_key must be 2 - 32 lowercase alphanumeric and hyphen.*"),
			},
		},
	})
}

func TestAccLocalGenericLikePackageTypes(t *testing.T) {
	for _, packageType := range local.PackageTypesLikeGeneric {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(localGenericLikeTestCase(packageType, t))
		})
	}
}

func localGenericLikeTestCase(packageType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("local-%s-%d-full", packageType, rand.Int())
	resourceName := fmt.Sprintf("artifactory_local_%s_repository.%s", packageType, name)
	xrayIndex := testutil.RandBool()
	fqrn := fmt.Sprintf("artifactory_local_%s_repository.%s", packageType, name)

	params := map[string]interface{}{
		"packageType":  packageType,
		"name":         name,
		"xrayIndex":    xrayIndex,
		"cdnRedirect":  false, // even when set to true, it comes back as false on the wire (presumably unless testing against a cloud platform)
		"property_set": "artifactory",
	}
	cfg := util.ExecuteTemplate("TestAccLocalRepository", `
		resource "artifactory_local_{{ .packageType }}_repository" "{{ .name }}" {
		  key           = "{{ .name }}"
		  description   = "Test repo for {{ .name }}"
		  notes         = "Test repo for {{ .name }}"
		  xray_index    = {{ .xrayIndex }}
		  cdn_redirect  = {{ .cdnRedirect }}
		  property_sets = ["{{ .property_set }}"]
		}
	`, params)

	updatedCfg := util.ExecuteTemplate("TestAccLocalRepository", `
		resource "artifactory_local_{{ .packageType }}_repository" "{{ .name }}" {
		  key           = "{{ .name }}"
		  description   = ""
		  notes         = ""
		  xray_index    = {{ .xrayIndex }}
		  cdn_redirect  = {{ .cdnRedirect }}
		  property_sets = ["{{ .property_set }}"]
		}
	`, params)

	return t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, resourceName, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "notes", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", packageType); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
					resource.TestCheckResourceAttr(resourceName, "xray_index", fmt.Sprintf("%t", xrayIndex)),
					resource.TestCheckResourceAttr(resourceName, "cdn_redirect", fmt.Sprintf("%t", params["cdnRedirect"])),
					resource.TestCheckResourceAttr(resourceName, "property_sets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "property_sets.0", params["property_set"].(string)),
				),
			},
			{
				Config: updatedCfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "notes", ""),
					resource.TestCheckResourceAttr(resourceName, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", packageType); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
					resource.TestCheckResourceAttr(resourceName, "xray_index", fmt.Sprintf("%t", xrayIndex)),
					resource.TestCheckResourceAttr(resourceName, "cdn_redirect", fmt.Sprintf("%t", params["cdnRedirect"])),
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
	}
}

func TestAccAllLocalLikeGenericPackageTypes_with_repo_layout_ref(t *testing.T) {
	for _, packageType := range local.PackageTypesLikeGeneric {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(makeLocalRepoTestCaseWithRepoLayoutRef(packageType, t))
		})
	}
}

func makeLocalRepoTestCaseWithRepoLayoutRef(packageType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("terraform-local-%s-%d-full", packageType, rand.Int())
	resourceName := fmt.Sprintf("artifactory_local_%s_repository.%s", packageType, name)
	repoLayoutRef := acctest.GetValidRandomDefaultRepoLayoutRef()
	fqrn := fmt.Sprintf("artifactory_local_%s_repository.%s", packageType, name)

	const config = `
		resource "artifactory_local_%[1]s_repository" "%[2]s" {
			key             = "%[2]s"
			description     = "Test repo for %[2]s"
			notes           = "Test repo for %[2]s"
			repo_layout_ref = "%[3]s"
		}
	`

	const updatedConfig = `
		resource "artifactory_local_%[1]s_repository" "%[2]s" {
			key             = "%[2]s"
			description     = ""
			notes           = ""
			repo_layout_ref = "%[3]s"
		}
	`

	return t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, resourceName, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config, packageType, name, repoLayoutRef),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "notes", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "repo_layout_ref", repoLayoutRef), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: fmt.Sprintf(updatedConfig, packageType, name, repoLayoutRef),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "notes", ""),
					resource.TestCheckResourceAttr(resourceName, "repo_layout_ref", repoLayoutRef),
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
	}
}
