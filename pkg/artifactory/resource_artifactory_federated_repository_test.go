package artifactory

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func federatedTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("terraform-federated-%s-%d-full", repoType, rand.Int())
	resourceType := fmt.Sprintf("artifactory_federated_%s_repository", repoType)
	resourceName := fmt.Sprintf("%s.%s", resourceType, name)
	//TODO: invalid URL will cause an error, to get 201, use the URL of created repository
	// Happy-path is to remove member completely, by default RT will assign the same repo as a member

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
				),
			},
		},
	}
}

func TestAccAllFederatedRepoTypesLocal(t *testing.T) {
	//TODO: test with all the package types
	for _, repo := range repoTypesLikeGenericFederated {
		t.Run(fmt.Sprintf("TestFederated%sRepo", strings.Title(strings.ToLower(repo))), func(t *testing.T) {
			resource.Test(federatedTestCase(repo, t))
		})
	}
}
