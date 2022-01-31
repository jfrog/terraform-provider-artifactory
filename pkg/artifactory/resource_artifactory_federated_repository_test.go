package artifactory

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

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

	cfg := fmt.Sprintf(federatedRepositoryConfigFull, resourceType, name, federatedMember1Url, federatedMember2Url)

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

	name := fmt.Sprintf("terraform-federated-%s-%d-full", repoType, rand.Int())
	resourceType := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
	resourceName := fmt.Sprintf("%s.%s", resourceType, name)
	federatedMemberUrl := fmt.Sprintf("%s/artifactory/%s", os.Getenv("ARTIFACTORY_URL"), name)

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
