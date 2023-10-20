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
	_, _, genericRepoName := testutil.MkNames("generic-local", "artifactory_local_generic_repository")
	_, _, alpineRepoName := testutil.MkNames("alpine-local", "artifactory_local_alpine_repository")
	_, fqrn, name := testutil.MkNames("all-local", "data.artifactory_local_all_repository")

	params := map[string]interface{}{
		"genericRepoName": genericRepoName,
		"alpineRepoName":  alpineRepoName,
		"name":            name,
	}
	config := utilsdk.ExecuteTemplate("TestAccLocalAllRepository", `
		resource "artifactory_local_generic_repository" "{{ .genericRepoName }}" {
		  key         = "{{ .genericRepoName }}"
		  description = "Test repo for {{ .genericRepoName }}"
		  notes       = "Test repo for {{ .genericRepoName }}"
		}

		resource "artifactory_local_alpine_repository" "{{ .alpineRepoName }}" {
		  count = 5
		  key         = "{{ .alpineRepoName }}-${count.index}"
		  description = "Test repo for {{ .alpineRepoName }}-${count.index}"
		  notes       = "Test repo for {{ .alpineRepoName }}-${count.index}"
		}

		data "artifactory_local_all_repository" "{{ .name }}" {
		  package_type = "alpine"

		  // ensure all repos are created first before fetching
		  depends_on = [
			artifactory_local_generic_repository.{{ .genericRepoName }},
			artifactory_local_alpine_repository.{{ .alpineRepoName }},
		  ]
		}

		output "repo_key_0" {
			value = tolist(data.artifactory_local_all_repository.{{ .name }}.repos)[0].key
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
					resource.TestCheckResourceAttr(fqrn, "repos.0.key", fmt.Sprintf("%s-0", params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.1.key", fmt.Sprintf("%s-1", params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.2.key", fmt.Sprintf("%s-2", params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.3.key", fmt.Sprintf("%s-3", params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.4.key", fmt.Sprintf("%s-4", params["alpineRepoName"])),
					resource.TestCheckOutput("repo_key_0", fmt.Sprintf("%s-0", params["alpineRepoName"])),
				),
			},
		},
	})
}
