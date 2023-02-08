package local_test

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func createLocalGenericRepository(name string, repoType string, t *testing.T) {
	localGenericRepository := local.RepositoryBaseParams{
		Key:         name,
		PackageType: repoType,
		Rclass:      "local",
		Description: "Test repo for " + name,
	}

	restyClient := acctest.GetTestResty(t)
	_, err := restyClient.R().
		SetBody(localGenericRepository).
		SetPathParam("key", localGenericRepository.Key).
		Put(repository.RepositoriesEndpoint)

	if err != nil {
		t.Fatal(err)
	}
}

func deleteLocalGenericRepository(name string, t *testing.T) error {
	restyClient := acctest.GetTestResty(t)
	_, err := restyClient.R().
		SetPathParam("key", name).
		Delete(repository.RepositoriesEndpoint)

	return err
}

func TestAccLocalAllRepoTypes(t *testing.T) {
	for _, repoType := range local.RepoTypesLikeGeneric {
		title := fmt.Sprintf("TestLocal%sRepo", cases.Title(language.AmericanEnglish).String(strings.ToLower(repoType)))
		t.Run(title, func(t *testing.T) {
			resource.Test(mkTestCase(repoType, t))
		})
	}
}

func mkTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("terraform-local-%s-%d-full", repoType, rand.Int())
	resourceName := fmt.Sprintf("data.artifactory_local_%s_repository.%s", repoType, name)

	params := map[string]interface{}{
		"repoType": repoType,
		"name":     name,
	}

	cfg := util.ExecuteTemplate("TestAccLocalRepository", `
		data "artifactory_local_{{ .repoType }}_repository" "{{ .name }}" {
		  key                 = "{{ .name }}"
		}
	`, params)

	return t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			createLocalGenericRepository(name, repoType, t)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			return deleteLocalGenericRepository(name, t)
		},
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", repoType),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test repo for %s", name)),
				),
			},
		},
	}
}

func TestAccLocalGenericRepository(t *testing.T) {
	_, fqrn, name := test.MkNames("generic-local", "data.artifactory_local_generic_repository")
	params := map[string]interface{}{
		"name": name,
	}

	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalGenericRepository", `
		data "artifactory_local_generic_repository" "{{ .name }}" {
		  key = "{{ .name }}"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			createLocalGenericRepository(name, "generic", t)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			return deleteLocalGenericRepository(name, t)
		},
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
				),
			},
		},
	})
}
