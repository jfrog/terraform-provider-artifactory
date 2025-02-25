package remote_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

var curationPackageTypes = []string{
	repository.ConanPackageType,
	repository.DockerPackageType,
	repository.GemsPackageType,
	repository.GoPackageType,
	repository.GradlePackageType,
	repository.HuggingFacePackageType,
	repository.MavenPackageType,
	repository.NPMPackageType,
	repository.NugetPackageType,
	repository.PyPiPackageType,
}

func TestAccRemoteRepository_with_curated(t *testing.T) {
	for _, packageType := range curationPackageTypes {
		t.Run(packageType, func(t *testing.T) {
			rs := fmt.Sprintf("artifactory_remote_%s_repository", packageType)
			_, fqrn, resourceName := testutil.MkNames("test-remote-curated-repo", rs)

			const temp = `
				resource "artifactory_remote_{{ .package_type }}_repository" "{{ .name }}" {
					key                     		= "{{ .name }}"
					description 					= "This is a test"
					url                     		= "https://tempurl.org/"
					repo_layout_ref         		= "simple-default"
					curated                         = true
				}
			`

			testData := map[string]string{
				"name":         resourceName,
				"package_type": packageType,
			}

			config := util.ExecuteTemplate("TestAccRemoteRepository_with_curated", temp, testData)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(fqrn, "key", resourceName),
							resource.TestCheckResourceAttr(fqrn, "curated", "true"),
						),
					},
					{
						ResourceName:                         fqrn,
						ImportState:                          true,
						ImportStateId:                        resourceName,
						ImportStateVerify:                    true,
						ImportStateVerifyIdentifierAttribute: "key",
					},
				},
			})
		})
	}
}
