package local_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func TestAccDataSourceLocalAllRepository(t *testing.T) {
	_, _, repoName := testutil.MkNames("alpine-local", "artifactory_local_alpine_repository")
	_, fqrn, name := testutil.MkNames("all-local", "data.artifactory_local_all_repository")

	params := map[string]interface{}{
		"repoName": repoName,
		"name":     name,
	}
	config := utilsdk.ExecuteTemplate("TestAccLocalAllRepository", `
		resource "artifactory_local_alpine_repository" "{{ .repoName }}" {
		  count = 5
		  key         = "{{ .repoName }}-${count.index}"
		  description = "Test repo for {{ .repoName }}-${count.index}"
		  notes       = "Test repo for {{ .repoName }}-${count.index}"
		}

		data "artifactory_local_all_repository" "{{ .name }}" {
		  package_type = "alpine"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repos.#", "5"),
					resource.TestCheckResourceAttr(fqrn, "repos.0.key", fmt.Sprintf("%s-0", params["repoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.1.key", fmt.Sprintf("%s-1", params["repoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.2.key", fmt.Sprintf("%s-2", params["repoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.3.key", fmt.Sprintf("%s-3", params["repoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.4.key", fmt.Sprintf("%s-4", params["repoName"])),
				),
			},
		},
	})
}
