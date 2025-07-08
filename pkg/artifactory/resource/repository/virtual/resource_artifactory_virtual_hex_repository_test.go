package virtual_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
)

func TestAccVirtualHexRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-repo", "artifactory_virtual_hex_repository")
	_, _, repoName := testutil.MkNames("local-repo", "artifactory_local_hex_repository")

	config := fmt.Sprintf(`
		resource "artifactory_local_hex_repository" "%s" {
			key = "%s"
			hex_primary_keypair_ref = "mykey"
		}

		resource "artifactory_virtual_hex_repository" "%s" {
			key          = "%s"
			repositories = [artifactory_local_hex_repository.%s.key]
			hex_primary_keypair_ref = "mykey"
		}
	`, repoName, repoName, name, name, repoName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", repository.HexPackageType),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", "mykey"),
				),
			},
		},
	})
}
