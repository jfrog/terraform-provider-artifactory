package virtual_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccVirtualRepository_basic(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_maven_repository.%s", name)
	const virtualRepositoryBasic = `
		resource "artifactory_virtual_maven_repository" "%s" {
			key          = "%s"
			repositories = []
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryBasic, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
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

func TestAccVirtualRepository_reset_default_deployment_repo(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("foo%d", id)
	localRepoName := fmt.Sprintf("%s-local", name)
	fqrn := fmt.Sprintf("artifactory_virtual_maven_repository.%s", name)
	const virtualRepositoryWithDefaultDeploymentRepo = `
		resource "artifactory_local_maven_repository" "%[1]s" {
			key = "%[1]s"
		}

		resource "artifactory_virtual_maven_repository" "%[2]s" {
			key          = "%[2]s"
			repositories = ["%[1]s"]
			default_deployment_repo = "%[1]s"
			depends_on = [artifactory_local_maven_repository.%[1]s]
		}
	`
	const virtualRepositoryWithoutDefaultDeploymentRepo = `
		resource "artifactory_local_maven_repository" "%[1]s" {
			key = "%[1]s"
		}

		resource "artifactory_virtual_maven_repository" "%[2]s" {
			key          = "%[2]s"
			repositories = ["%[1]s"]
			depends_on = [artifactory_local_maven_repository.%[1]s]
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryWithoutDefaultDeploymentRepo, localRepoName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "default_deployment_repo", ""),
				),
			},
			{
				Config: fmt.Sprintf(virtualRepositoryWithDefaultDeploymentRepo, localRepoName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "default_deployment_repo", localRepoName),
				),
			},
			{
				Config: fmt.Sprintf(virtualRepositoryWithoutDefaultDeploymentRepo, localRepoName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "default_deployment_repo", ""),
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

func TestAccVirtualGoRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("foo", "artifactory_virtual_go_repository")
	const packageType = "go"
	var virtualRepositoryBasic = fmt.Sprintf(`
		resource "artifactory_virtual_go_repository" "%s" {
		  key          = "%s"
		  repositories = []
		  description = "A test virtual repo"
		  notes = "Internal description"
		  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
		  excludes_pattern = "com/google/**"
		  external_dependencies_enabled = true
		  external_dependencies_patterns = [
			"**/github.com/**",
			"**/go.googlesource.com/**"
		  ]
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				// we check to make sure some of the base params are picked up
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", packageType),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "**/github.com/**"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.1", "**/go.googlesource.com/**"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef(virtual.Rclass, packageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
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

func TestAccVirtualConanRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("foo", "artifactory_virtual_conan_repository")
	var virtualRepositoryBasic = fmt.Sprintf(`
		resource "artifactory_virtual_conan_repository" "%s" {
		  key          = "%s"
		  repo_layout_ref = "conan-default"
		  repositories = []
		  description = "A test virtual repo"
		  notes = "Internal description"
		  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
		  excludes_pattern = "com/google/**"
 		  retrieval_cache_period_seconds = 7100
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				// we check to make sure some of the base params are picked up
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "conan"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "7100"),
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

func TestAccVirtualGenericRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("foo", "artifactory_virtual_generic_repository")
	const packageType = "generic"
	var virtualRepositoryBasic = fmt.Sprintf(`
		resource "artifactory_virtual_generic_repository" "%s" {
		  key          = "%s"
		  repositories = []
		  description = "A test virtual repo"
		  notes = "Internal description"
		  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
		  excludes_pattern = "com/google/**"
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				// we check to make sure some of the base params are picked up
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", packageType),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef(virtual.Rclass, packageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
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

func TestAccVirtualMavenRepository_basic(t *testing.T) {
	const packageType = "maven"

	id := testutil.RandomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_maven_repository.%s", name)
	var virtualRepositoryBasic = fmt.Sprintf(`
		resource "artifactory_virtual_maven_repository" "%s" {
			key          = "%s"
			repositories = []
			description = "A test virtual repo"
			notes = "Internal description"
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			force_maven_authentication = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				// we check to make sure some of the base params are picked up
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", packageType),
					resource.TestCheckResourceAttr(fqrn, "force_maven_authentication", "true"),
					// to test key pair, we'd have to be able to create them on the fly and we currently can't.
					resource.TestCheckResourceAttr(fqrn, "key_pair", ""),
					resource.TestCheckResourceAttr(fqrn, "pom_repository_references_cleanup_policy", "discard_active_reference"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef(virtual.Rclass, packageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
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

func TestAccVirtualHelmRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("virtual-helm-repo", "artifactory_virtual_helm_repository")
	useNamespaces := testutil.RandBool()

	params := map[string]interface{}{
		"name":          name,
		"useNamespaces": useNamespaces,
	}
	virtualRepositoryBasic := util.ExecuteTemplate("TestAccVirtualHelmRepository", `
		resource "artifactory_virtual_helm_repository" "{{ .name }}" {
		  key            				 = "{{ .name }}"
	 	  use_namespaces 				 = {{ .useNamespaces }}
		  retrieval_cache_period_seconds = 650
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "helm"),
					resource.TestCheckResourceAttr(fqrn, "use_namespaces", fmt.Sprintf("%t", useNamespaces)),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "650"),
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

func TestAccVirtualHelmOciRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase("helmoci", t, map[string]interface{}{
		"description":                   "helmoci virtual repository public description testing.",
		"resolve_oci_tags_by_timestamp": true,
	}))
}

func TestAccVirtualRpmRepository(t *testing.T) {
	const packageType = "rpm"
	_, fqrn, name := testutil.MkNames("virtual-rpm-repo", "artifactory_virtual_rpm_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1-", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2-", "artifactory_keypair")
	virtualRepositoryBasic := util.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
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
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_virtual_rpm_repository" "{{ .repo_name }}" {
			key 	              = "{{ .repo_name }}"
			primary_keypair_ref   = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name

			depends_on = [
				artifactory_keypair.{{ .kp_name }},
				artifactory_keypair.{{ .kp_name2 }},
			]
		}
	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"kp_id2":      kpId2,
		"kp_name2":    kpName2,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_PGP_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_PGP_PUBLIC_KEY"),
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "", security.VerifyKeyPair),
			acctest.VerifyDeleted(t, kpFqrn2, "", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", packageType),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef(virtual.Rclass, packageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
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

func TestAccVirtualRepository_update(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_maven_repository.%s", name)
	const virtualRepositoryUpdateBefore = `
		resource "artifactory_virtual_maven_repository" "%s" {
			key          = "%s"
			description  = "Before"
			repositories = []
			artifactory_requests_can_retrieve_remote_artifacts = true
		}
	`
	const virtualRepositoryUpdateAfter = `
		resource "artifactory_virtual_maven_repository" "%s" {
			key          = "%s"
			description  = "After"
			repositories = []
			artifactory_requests_can_retrieve_remote_artifacts = false
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryUpdateBefore, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "description", "Before"),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "artifactory_requests_can_retrieve_remote_artifacts", "true"),
				),
			},
			{
				Config: fmt.Sprintf(virtualRepositoryUpdateAfter, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "description", "After"),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "artifactory_requests_can_retrieve_remote_artifacts", "false"),
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

func TestAccVirtualNugetRepository_PackageCreationFull(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_nuget_repository.%s", name)
	const virtualRepositoryFull = `
		resource "artifactory_virtual_nuget_repository" "%s" {
			key = "%s"
			repo_layout_ref = "nuget-default"
			repositories = []
			description = "A test virtual repo"
			notes = "Internal description"
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			artifactory_requests_can_retrieve_remote_artifacts = true
			force_nuget_authentication	= true
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryFull, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "nuget"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "nuget-default"),
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

func TestAccVirtualRepository_full(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_maven_repository.%s", name)
	const virtualRepositoryFull = `
		resource "artifactory_virtual_maven_repository" "%s" {
			key = "%s"
			repo_layout_ref = "maven-1-default"
			repositories = []
			description = "A test virtual repo"
			notes = "Internal description"
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			artifactory_requests_can_retrieve_remote_artifacts = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryFull, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "maven-1-default"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "description", "A test virtual repo"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal description"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "com/jfrog/**,cloud/jfrog/**"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "com/google/**"),
					resource.TestCheckResourceAttr(fqrn, "pom_repository_references_cleanup_policy", "discard_active_reference"),
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

func TestAccVirtualGenericRepositoryWithProjectAttributesGH318(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	projectEnv := testutil.RandSelect("DEV", "PROD").(string)
	repoName := fmt.Sprintf("%s-generic-virtual", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_virtual_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
		"projectEnv": projectEnv,
	}
	virtualRepositoryBasic := util.ExecuteTemplate("TestAccVirtualGenericRepository", `
		resource "artifactory_virtual_generic_repository" "{{ .name }}" {
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
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "", func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
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

func TestAccVirtualRepositoryWithInvalidProjectKeyGH318(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-virtual", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_virtual_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	virualRepositoryBasic := util.ExecuteTemplate("TestAccVirtualGenericRepository", `
		resource "artifactory_virtual_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = "invalid-project-key-too-long-really-long"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "", func(id string, request *resty.Request) (*resty.Response, error) {
			acctest.DeleteProject(t, projectKey)
			return acctest.CheckRepo(id, request)
		}),
		Steps: []resource.TestStep{
			{
				Config:      virualRepositoryBasic,
				ExpectError: regexp.MustCompile(".*project_key must be 2 - 32 lowercase alphanumeric and hyphen characters"),
			},
		},
	})
}

func TestAccVirtualRepository(t *testing.T) {
	for _, repoType := range virtual.PackageTypesLikeGeneric {
		// TODO: workaround due to bug in Artifactory 7.55.2, 'bypass_head_requests' inconsistency for terraform repo type.
		if repoType != "terraform" {
			t.Run(repoType, func(t *testing.T) {
				resource.Test(mkNewVirtualTestCase(repoType, t, map[string]interface{}{
					"description": fmt.Sprintf("%s virtual repository public description testing.", repoType),
				}))
			})
		}
	}
	for _, repoType := range virtual.PackageTypesLikeGenericWithRetrievalCachePeriodSecs {
		t.Run(repoType, func(t *testing.T) {
			resource.Test(mkNewVirtualTestCase(repoType, t, map[string]interface{}{
				"description":                    fmt.Sprintf("%s virtual repository public description testing.", repoType),
				"retrieval_cache_period_seconds": 650,
			}))
		})
	}
}

func TestAccAllVirtualGradleLikeRepository(t *testing.T) {
	for _, packageType := range repository.PackageTypesLikeGradle {
		t.Run(packageType, func(t *testing.T) {
			resource.Test(mkNewVirtualTestCase(packageType, t, map[string]interface{}{
				"description": fmt.Sprintf("%s virtual repository public description testing.", packageType),
				"pom_repository_references_cleanup_policy": "discard_active_reference",
			}))
		})
	}
}

// if you wish to override any of the default fields, just pass it as "extraFields" as these will overwrite
func mkNewVirtualTestCase(packageType string, t *testing.T, extraFields map[string]interface{}) (*testing.T, resource.TestCase) {
	_, fqrn, name := testutil.MkNames("terraform-virtual-test-repo-full-", fmt.Sprintf("artifactory_virtual_%s_repository", packageType))
	remoteRepoName := fmt.Sprintf("%s-remote", name)
	defaultFields := map[string]interface{}{
		"key":         name,
		"description": "A test virtual repo",
		"notes":       "Internal description",
	}
	allFields := utilsdk.MergeMaps(defaultFields, extraFields)
	allFieldsHcl := utilsdk.FmtMapToHcl(allFields)
	const virtualRepoFull = `
        resource "artifactory_remote_%[1]s_repository" "%[3]s" {
			key = "%[3]s"
            url = "http://tempurl.org"
		}

		resource "artifactory_virtual_%[1]s_repository" "%[2]s" {
%[4]s
            repositories = ["%[3]s"]
            depends_on = [artifactory_remote_%[1]s_repository.%[3]s]
		}
	`
	extraChecks := testutil.MapToTestChecks(fqrn, extraFields)
	defaultChecks := testutil.MapToTestChecks(fqrn, allFields)
	defaultChecks = append(defaultChecks, resource.TestCheckResourceAttr(fqrn, "package_type", packageType))

	checks := append(defaultChecks, extraChecks...)
	config := fmt.Sprintf(virtualRepoFull, packageType, name, remoteRepoName, allFieldsHcl)

	updatedFields := utilsdk.MergeMaps(defaultFields, extraFields, map[string]any{
		"description": "",
		"notes":       "",
	})
	updatedFieldsHcl := utilsdk.FmtMapToHcl(updatedFields)
	updatedConfig := fmt.Sprintf(virtualRepoFull, packageType, name, remoteRepoName, updatedFieldsHcl)
	updatedChecks := testutil.MapToTestChecks(fqrn, updatedFields)
	updatedChecks = append(updatedChecks, resource.TestCheckResourceAttr(fqrn, "package_type", packageType))

	return t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
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
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	}
}

func TestAccVirtualAlpineRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase("alpine", t, map[string]interface{}{
		"description":                    "alpine virtual repository public description testing.",
		"retrieval_cache_period_seconds": 650,
	}))
}

func TestAccVirtualAnsibleRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase("ansible", t, map[string]interface{}{
		"description":                    "ansible virtual repository public description testing.",
		"retrieval_cache_period_seconds": 650,
	}))
}

func TestAccVirtualAlpineRepositoryZeroRetrievalPeriod(t *testing.T) {
	resource.Test(mkNewVirtualTestCase("alpine", t, map[string]interface{}{
		"description":                    "alpine virtual repository public description testing.",
		"retrieval_cache_period_seconds": 0,
	}))
}

func TestAccVirtualConanRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase("conan", t, map[string]interface{}{
		"description":                    "conan virtual repository public description testing.",
		"retrieval_cache_period_seconds": 650,
		"force_conan_authentication":     true,
	}))
}

func TestAccVirtualNugetRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase("nuget", t, map[string]interface{}{
		"description":                "nuget virtual repository public description testing.",
		"force_nuget_authentication": true,
	}))
}

func TestAccVirtualDockerRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase("docker", t, map[string]interface{}{
		"description":                      "docker virtual repository public description testing.",
		"resolve_docker_tags_by_timestamp": true,
	}))
}

func TestAccVirtualBowerExternalDependenciesRepository(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("bower-virtual-%d", id)
	remoteRepoName := fmt.Sprintf("bower-remote-%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_bower_repository.%s", name)

	params := map[string]interface{}{
		"name":           name,
		"remoteRepoName": remoteRepoName,
	}
	config := util.ExecuteTemplate("TestAccVirtualBower", `
		resource "artifactory_remote_bower_repository" "bower-remote" {
			key = "{{ .remoteRepoName }}"
			url = "https://registry.npmjs.org"
		}

		resource "artifactory_virtual_bower_repository" "{{ .name }}" {
			key                               = "{{ .name }}"
			repositories                      = ["{{ .remoteRepoName }}"]
			external_dependencies_enabled     = true
			external_dependencies_patterns    = ["**/github.com/**", "**/go.googlesource.com/**"]
			external_dependencies_remote_repo = "{{ .remoteRepoName }}"

			depends_on = ["artifactory_remote_bower_repository.bower-remote"]
		}
	`, params)

	updatedConfig := util.ExecuteTemplate("TestAccVirtualBower", `
		resource "artifactory_remote_bower_repository" "bower-remote" {
			key = "{{ .remoteRepoName }}"
			url = "https://registry.npmjs.org"
		}

		resource "artifactory_virtual_bower_repository" "{{ .name }}" {
			key                               = "{{ .name }}"
			repositories                      = ["{{ .remoteRepoName }}"]
			external_dependencies_enabled     = true
			external_dependencies_patterns    = ["**/go.googlesource.com/**"]
			external_dependencies_remote_repo = "{{ .remoteRepoName }}"

			depends_on = ["artifactory_remote_bower_repository.bower-remote"]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "bower"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repositories.0", remoteRepoName),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/github.com/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/go.googlesource.com/**"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "bower"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repositories.0", remoteRepoName),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/go.googlesource.com/**"),
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

func TestAccVirtualGoExternalDependenciesRepository(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("go-virtual-%d", id)
	remoteRepoName := fmt.Sprintf("go-remote-%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_go_repository.%s", name)

	params := map[string]interface{}{
		"name":           name,
		"remoteRepoName": remoteRepoName,
	}
	config := util.ExecuteTemplate("TestAccVirtualGo", `
	resource "artifactory_remote_go_repository" "go-remote" {
		key = "{{ .remoteRepoName }}"
		url = "https://proxy.golang.org/"
	}

	resource "artifactory_virtual_go_repository" "{{ .name }}" {
			key                               = "{{ .name }}"
			repositories                      = [artifactory_remote_go_repository.go-remote.key]
			external_dependencies_enabled     = true
			external_dependencies_patterns    = [
				"**/github.com/**",
				"**/bitbucket.org/**",
				"**/gopkg.in/**",
				"**/golang.org/**",
				"**/k8s.io/**",
			]
		}
	`, params)

	updatedConfig := util.ExecuteTemplate("TestAccVirtualGo", `
		resource "artifactory_remote_go_repository" "go-remote" {
			key = "{{ .remoteRepoName }}"
			url = "https://proxy.golang.org/"
		}

		resource "artifactory_virtual_go_repository" "{{ .name }}" {
			key                               = "{{ .name }}"
			repositories                      = [artifactory_remote_go_repository.go-remote.key]
			external_dependencies_enabled     = true
			external_dependencies_patterns    = [
				"**/github.com/**",
				"**/gopkg.in/**",
				"**/golang.org/**",
				"**/k8s.io/**",
			]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "go"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repositories.0", remoteRepoName),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "5"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/github.com/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/bitbucket.org/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/gopkg.in/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/golang.org/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/k8s.io/**"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "go"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repositories.0", remoteRepoName),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "4"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/github.com/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/gopkg.in/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/golang.org/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/k8s.io/**"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fqrn, plancheck.ResourceActionUpdate),
					},
				},
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

func TestAccVirtualNpmExternalDependenciesRepository(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("npm-virtual-%d", id)
	remoteRepoName := fmt.Sprintf("npm-remote-%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_npm_repository.%s", name)

	params := map[string]interface{}{
		"name":           name,
		"remoteRepoName": remoteRepoName,
	}
	config := util.ExecuteTemplate("TestAccVirtualNpm", `
		resource "artifactory_remote_npm_repository" "npm-remote" {
			key = "{{ .remoteRepoName }}"
			url = "https://registry.npmjs.org"
		}

		resource "artifactory_virtual_npm_repository" "{{ .name }}" {
			key                               = "{{ .name }}"
			repositories                      = ["{{ .remoteRepoName }}"]
			external_dependencies_enabled     = true
			retrieval_cache_period_seconds    = 650
			external_dependencies_patterns    = ["**/github.com/**", "**/go.googlesource.com/**"]
			external_dependencies_remote_repo = "{{ .remoteRepoName }}"

			depends_on = ["artifactory_remote_npm_repository.npm-remote"]
		}
	`, params)

	updatedConfig := util.ExecuteTemplate("TestAccVirtualNpm", `
		resource "artifactory_remote_npm_repository" "npm-remote" {
			key = "{{ .remoteRepoName }}"
			url = "https://registry.npmjs.org"
		}

		resource "artifactory_virtual_npm_repository" "{{ .name }}" {
			key                               = "{{ .name }}"
			repositories                      = ["{{ .remoteRepoName }}"]
			external_dependencies_enabled     = true
			retrieval_cache_period_seconds    = 650
			external_dependencies_patterns    = ["**/go.googlesource.com/**"]
			external_dependencies_remote_repo = "{{ .remoteRepoName }}"

			depends_on = ["artifactory_remote_npm_repository.npm-remote"]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "npm"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repositories.0", remoteRepoName),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "650"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/github.com/**"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/go.googlesource.com/**"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "npm"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repositories.0", remoteRepoName),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "650"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "external_dependencies_patterns.*", "**/go.googlesource.com/**"),
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

func TestAccVirtualOciRepository(t *testing.T) {
	resource.Test(mkNewVirtualTestCase("oci", t, map[string]interface{}{
		"description":                   "oci virtual repository public description testing.",
		"resolve_oci_tags_by_timestamp": true,
	}))
}

func TestAccVirtualDebianRepository_full(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_debian_repository.%s", name)
	const virtualRepositoryBasic = `
		resource "artifactory_virtual_debian_repository" "%s" {
			key          = "%s"
			repositories = []
            debian_default_architectures = "i386,amd64"
			retrieval_cache_period_seconds = 650
            optional_index_compression_formats = [
                "bz2",
                "xz",
            ]
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryBasic, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "debian"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "650"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
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
