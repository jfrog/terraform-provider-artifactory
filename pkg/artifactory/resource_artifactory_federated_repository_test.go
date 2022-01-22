package artifactory

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFederatedRepoWithMembers(t *testing.T) {
	name := fmt.Sprintf("terraform-federated-generic-%d-full", rand.Int())
	resourceType := "artifactory_federated_generic_repository"
	resourceName := fmt.Sprintf("%s.%s", resourceType, name)
	artifactoryUrl := os.Getenv("ARTIFACTORY_URL")
	federatedMemberUrl := fmt.Sprintf("%s/%s", artifactoryUrl, name)

	const federatedRepositoryConfigFull = `
		resource "%s" "%[2]s" {
			key         = "%[2]s"
			description = "Test federated repo for %[2]s"
			notes       = "Test federated repo for %[2]s"

			member {
				url     = "http://tempurl.org/test"
				enabled = true
			}
		}
	`

	cfg := fmt.Sprintf(federatedRepositoryConfigFull, resourceType, name, federatedMemberUrl)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, testCheckRepo),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"), // In addition to the one specificed in HCL, Artifactory also add another entry with the source repo
					resource.TestCheckResourceAttr(resourceName, "member.0.url", federatedMemberUrl),
					resource.TestCheckResourceAttr(resourceName, "member.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "member.1.url", "http://tempurl.org/test"),
					resource.TestCheckResourceAttr(resourceName, "member.1.enabled", "true"),
				),
			},
		},
	})
}

func federatedTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("terraform-federated-%s-%d-full", repoType, rand.Int())
	resourceType := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
	resourceName := fmt.Sprintf("%s.%s", resourceType, name)

	const federatedRepositoryConfigFull = `
		resource "%s" "%[2]s" {
			key         = "%[2]s"
			description = "Test federated repo for %[2]s"
			notes       = "Test federated repo for %[2]s"
		}
	`

	cfg := fmt.Sprintf(federatedRepositoryConfigFull, resourceType, name)
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

					resource.TestCheckResourceAttr(resourceName, "member.#", "1"), // Artifactory also add another entry with the source repo
					resource.TestCheckResourceAttr(resourceName, "member.0.url", fmt.Sprintf("%s/artifactory/%s", os.Getenv("ARTIFACTORY_URL"), name)),
					resource.TestCheckResourceAttr(resourceName, "member.0.enabled", "true"),
				),
			},
		},
	}
}

func TestAccAllFederatedRepoTypes(t *testing.T) {
	for _, repo := range repoTypesLikeGeneric {
		t.Run(fmt.Sprintf("TestFederated%sRepo", strings.Title(strings.ToLower(repo))), func(t *testing.T) {
			resource.Test(federatedTestCase(repo, t))
		})
	}
}
