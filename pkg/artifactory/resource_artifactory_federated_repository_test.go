package artifactory

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func skipFederatedRepo() (bool, string) {
	if len(os.Getenv("ARTIFACTORY_URL_2")) > 0 {
		return false, "Env var `ARTIFACTORY_URL_2` is set. Executing test."
	}

	return true, "Env var `ARTIFACTORY_URL_2` is not set. Skipping test."
}

func TestAccFederatedRepoWithMembers(t *testing.T) {
	if skip, reason := skipFederatedRepo(); skip {
		t.Skipf(reason)
	}

	name := fmt.Sprintf("terraform-federated-generic-%d-full", rand.Int())
	resourceType := "artifactory_federated_generic_repository"
	resourceName := fmt.Sprintf("%s.%s", resourceType, name)
	federatedMember1Url := fmt.Sprintf("%s/artifactory/%s", os.Getenv("ARTIFACTORY_URL"), name)
	federatedMember2Url := fmt.Sprintf("%s/artifactory/%s", os.Getenv("ARTIFACTORY_URL_2"), name)

	params := map[string]interface{}{
		"resourceType": resourceType,
		"name":         name,
		"member1Url":   federatedMember1Url,
		"member2Url":   federatedMember2Url,
	}
	federatedRepositoryConfig := executeTemplate("TestAccFederatedRepositoryConfigWithMembers", `
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
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, testCheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "member.0.url", federatedMember2Url),
					resource.TestCheckResourceAttr(resourceName, "member.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "member.1.url", federatedMember1Url),
					resource.TestCheckResourceAttr(resourceName, "member.1.enabled", "true"),
				),
			},
		},
	})
}

func federatedTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	if skip, reason := skipFederatedRepo(); skip {
		t.Skipf(reason)
	}

	name := fmt.Sprintf("terraform-federated-%s-%d", repoType, rand.Int())
	resourceType := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
	resourceName := fmt.Sprintf("%s.%s", resourceType, name)
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", os.Getenv("ARTIFACTORY_URL"), name)

	params := map[string]interface{}{
		"resourceType": resourceType,
		"name":         name,
		"memberUrl":    federatedMemberUrl,
	}
	federatedRepositoryConfig := executeTemplate("TestAccFederatedRepositoryConfig", `
		resource "{{ .resourceType }}" "{{ .name }}" {
			key         = "{{ .name }}"
			description = "Test federated repo for {{ .name }}"
			notes       = "Test federated repo for {{ .name }}"

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`, params)

	return t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, testCheckRepo),
		Steps: []resource.TestStep{
			{
				Config: federatedRepositoryConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", repoType),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test federated repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "notes", fmt.Sprintf("Test federated repo for %s", name)),

					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member.0.url", federatedMemberUrl),
					resource.TestCheckResourceAttr(resourceName, "member.0.enabled", "true"),
				),
			},
		},
	}
}

func TestAccFederatedRepoAllTypes(t *testing.T) {
	for _, repo := range federatedRepoTypesSupported {
		t.Run(fmt.Sprintf("TestFederated%sRepo", strings.Title(strings.ToLower(repo))), func(t *testing.T) {
			resource.Test(federatedTestCase(repo, t))
		})
	}
}

func TestAccFederatedRepoWithProjectKeyGH318(t *testing.T) {
	if skip, reason := skipFederatedRepo(); skip {
		t.Skipf(reason)
	}

	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", rand.Intn(100000000))
	projectEnv := randProjectEnv()
	repoName := fmt.Sprintf("%s-generic-federated", projectKey)

	_, fqrn, name := mkNames(repoName, "artifactory_federated_generic_repository")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", os.Getenv("ARTIFACTORY_URL"), name)

	params := map[string]interface{}{
		"name":         name,
		"projectKey":  projectKey,
		"projectEnv": projectEnv,
		"memberUrl":    federatedMemberUrl,
	}
	federatedRepositoryConfig := executeTemplate("TestAccFederatedRepositoryConfig", `
		resource "artifactory_federated_generic_repository" "{{ .name }}" {
			key         = "{{ .name }}"
			project_key = "{{ .projectKey }}"
	 		project_environments = ["{{ .projectEnv }}"]

			member {
				url     = "{{ .memberUrl }}"
				enabled = true
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck: func() {
			testAccPreCheck(t)
			createProject(t, projectKey)
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			deleteProject(t, projectKey)
			return testCheckRepo(id, request)
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
		},
	})
}
