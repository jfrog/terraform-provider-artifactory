package repository_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func TestAccRepository_assign_project_key_gh_329(t *testing.T) {

	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	localRepositoryBasic := utilsdk.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		}
	`, map[string]interface{}{
		"name": name,
	})

	localRepositoryWithProjectKey := utilsdk.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key         = "{{ .name }}"
	 	  project_key = "{{ .projectKey }}"
		}
	`, map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	})

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
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
				),
			},
			{
				Config: localRepositoryWithProjectKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
				),
			},
		},
	})
}

func TestAccRepository_unassign_project_key_gh_329(t *testing.T) {

	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	localRepositoryWithProjectKey := utilsdk.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key         = "{{ .name }}"
	 	  project_key = "{{ .projectKey }}"
		  project_environments = ["DEV"]
		}
	`, map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	})

	localRepositoryNoProjectKey := utilsdk.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		}
	`, map[string]interface{}{
		"name": name,
	})

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
				Config: localRepositoryWithProjectKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
				),
			},
			{
				Config: localRepositoryNoProjectKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_key", ""),
				),
			},
		},
	})
}

func TestAccRepository_can_set_two_project_environments_before_7_53_1(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := utilsdk.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = "{{ .projectKey }}"
	 	  project_environments = ["DEV", "PROD"]
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
				SkipFunc: func() (bool, error) {
					meta := acctest.Provider.Meta().(utilsdk.ProvderMetadata)
					return utilsdk.CheckVersion(meta.ArtifactoryVersion, repository.CustomProjectEnvironmentSupportedVersion)
				},
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.0", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.1", "PROD"),
				),
			},
		},
	})
}

func TestAccRepository_invalid_project_environments_before_7_53_1(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := utilsdk.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = "{{ .projectKey }}"
	 	  project_environments = ["Foo"]
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
				SkipFunc: func() (bool, error) {
					meta := acctest.Provider.Meta().(utilsdk.ProvderMetadata)
					return utilsdk.CheckVersion(meta.ArtifactoryVersion, repository.CustomProjectEnvironmentSupportedVersion)
				},
				Config:      localRepositoryBasic,
				ExpectError: regexp.MustCompile(".*project_environment Foo not allowed.*"),
			},
		},
	})
}

func TestAccRepository_invalid_project_environments_after_7_53_1(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := utilsdk.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = "{{ .projectKey }}"
	 	  project_environments = ["DEV", "PROD"]
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
				SkipFunc: func() (bool, error) {
					meta := acctest.Provider.Meta().(utilsdk.ProvderMetadata)
					isSupported, err := utilsdk.CheckVersion(meta.ArtifactoryVersion, repository.CustomProjectEnvironmentSupportedVersion)
					return !isSupported, err
				},
				Config:      localRepositoryBasic,
				ExpectError: regexp.MustCompile(fmt.Sprintf(".*for Artifactory %s or later, only one environment can be assigned to a repository.*", repository.CustomProjectEnvironmentSupportedVersion)),
			},
		},
	})
}
