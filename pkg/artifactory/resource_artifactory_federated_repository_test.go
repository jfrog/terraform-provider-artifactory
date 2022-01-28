package artifactory

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func skipFederatedRepo() (bool, error) {
	artifactory2Url := os.Getenv("ARTIFACTORY_URL_2")
	testFederatedRepo := os.Getenv("ARTIFACTORY_TEST_FEDERATED_REPO")

	if testFederatedRepo == "true" && len(artifactory2Url) > 0 {
		log.Println("Both env var `ARTIFACTORY_TEST_FEDERATED_REPO` and `ARTIFACTORY_URL_2` are set. Executing test.")
		return false, nil
	}

	log.Println("Either env var `ARTIFACTORY_TEST_FEDERATED_REPO` or `ARTIFACTORY_URL_2` is not set. Skipping test.")
	return true, nil
}

func TestAccFederatedRepoWithMembers(t *testing.T) {
	name := fmt.Sprintf("terraform-federated-generic-%d-full", rand.Int())
	resourceType := "artifactory_federated_generic_repository"
	resourceName := fmt.Sprintf("%s.%s", resourceType, name)
	artifactoryUrl := os.Getenv("ARTIFACTORY_URL")
	artifactory2Url := os.Getenv("ARTIFACTORY_URL_2")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", artifactoryUrl, name)
	federatedMember2Url := fmt.Sprintf("%s/artifactory/%s", artifactory2Url, name)

	const federatedRepositoryConfigFull = `
		resource "%s" "%[2]s" {
			key         = "%[2]s"
			description = "Test federated repo for %[2]s"
			notes       = "Test federated repo for %[2]s"

			member {
				url     = "%[3]s"
				enabled = true
			}

			member {
				url     = "%[4]s"
				enabled = true
			}
		}
	`

	cfg := fmt.Sprintf(federatedRepositoryConfigFull, resourceType, name, federatedMemberUrl, federatedMember2Url)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, testCheckRepo),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "member.0.url", federatedMember2Url),
					resource.TestCheckResourceAttr(resourceName, "member.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "member.1.url", federatedMemberUrl),
					resource.TestCheckResourceAttr(resourceName, "member.1.enabled", "true"),
				),
				SkipFunc: skipFederatedRepo,
			},
		},
	})
}

func federatedTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("terraform-federated-%s-%d-full", repoType, rand.Int())
	resourceType := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
	resourceName := fmt.Sprintf("%s.%s", resourceType, name)
	artifactoryUrl := os.Getenv("ARTIFACTORY_URL")
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", artifactoryUrl, name)

	const federatedRepositoryConfigFull = `
		resource "%s" "%[2]s" {
			key         = "%[2]s"
			description = "Test federated repo for %[2]s"
			notes       = "Test federated repo for %[2]s"

			member {
				url     = "%[3]s"
				enabled = true
			}
		}
	`

	cfg := fmt.Sprintf(federatedRepositoryConfigFull, resourceType, name, federatedMemberUrl)
	return t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, testCheckRepo),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", repoType),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test federated repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "notes", fmt.Sprintf("Test federated repo for %s", name)),

					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member.0.url", federatedMemberUrl),
					resource.TestCheckResourceAttr(resourceName, "member.0.enabled", "true"),
				),
				SkipFunc: skipFederatedRepo,
			},
		},
	}
}

func TestAccFederatedRepoAllTypes(t *testing.T) {
	for _, repo := range repoTypesLikeGeneric {
		t.Run(fmt.Sprintf("TestFederated%sRepo", strings.Title(strings.ToLower(repo))), func(t *testing.T) {
			resource.Test(federatedTestCase(repo, t))
		})
	}
}
