package repository_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRepository_assign_project_key_gh_329(t *testing.T) {

	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := test.MkNames(repoName, "artifactory_local_generic_repository")

	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		}
	`, map[string]interface{}{
		"name": name,
	})

	localRepositoryWithProjectKey := util.ExecuteTemplate("TestAccLocalGenericRepository", `
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
	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := test.MkNames(repoName, "artifactory_local_generic_repository")

	localRepositoryWithProjectKey := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key         = "{{ .name }}"
	 	  project_key = "{{ .projectKey }}"
		  project_environments = ["DEV"]
		}
	`, map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	})

	localRepositoryNoProjectKey := util.ExecuteTemplate("TestAccLocalGenericRepository", `
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
					resource.TestCheckResourceAttr(fqrn, "project_key", "default"),
				),
			},
		},
	})
}

func TestAccRepository_can_set_two_project_environments_before_7_53_1(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := test.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
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
					meta := acctest.Provider.Meta().(util.ProvderMetadata)
					return util.CheckVersion(meta.ArtifactoryVersion, repository.CustomProjectEnvironmentSupportedVersion)
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
	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := test.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
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
					meta := acctest.Provider.Meta().(util.ProvderMetadata)
					return util.CheckVersion(meta.ArtifactoryVersion, repository.CustomProjectEnvironmentSupportedVersion)
				},
				Config:      localRepositoryBasic,
				ExpectError: regexp.MustCompile(".*project_environment Foo not allowed.*"),
			},
		},
	})
}

func TestAccRepository_invalid_project_environments_after_7_53_1(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	repoName := fmt.Sprintf("%s-generic-local", projectKey)

	_, fqrn, name := test.MkNames(repoName, "artifactory_local_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
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
					meta := acctest.Provider.Meta().(util.ProvderMetadata)
					isSupported, err := util.CheckVersion(meta.ArtifactoryVersion, repository.CustomProjectEnvironmentSupportedVersion)
					return !isSupported, err
				},
				Config:      localRepositoryBasic,
				ExpectError: regexp.MustCompile(fmt.Sprintf(".*For Artifactory %s or later, only one environment can be assigned to a repository..*", repository.CustomProjectEnvironmentSupportedVersion)),
			},
		},
	})
}
