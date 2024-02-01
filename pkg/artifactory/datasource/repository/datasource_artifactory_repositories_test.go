package repository_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccDataSourceRepositories(t *testing.T) {
	_, _, genericRepoName := testutil.MkNames("generic-local", "artifactory_local_generic_repository")
	_, _, alpineRepoName := testutil.MkNames("alpine", "artifactory_local_alpine_repository")
	_, fqrn, name := testutil.MkNames("all-local", "data.artifactory_repositories")

	params := map[string]interface{}{
		"genericRepoName": genericRepoName,
		"alpineRepoName":  alpineRepoName,
		"name":            name,
	}
	config := util.ExecuteTemplate("TestAccLocalRepositories", `
		resource "artifactory_local_generic_repository" "{{ .genericRepoName }}" {
		  key         = "{{ .genericRepoName }}"
		  description = "Test repo for {{ .genericRepoName }}"
		  notes       = "Test repo for {{ .genericRepoName }}"
		}

		resource "artifactory_local_alpine_repository" "{{ .alpineRepoName }}" {
		  count = 5
		  key         = "{{ .alpineRepoName }}-local-${count.index}"
		  description = "Test local repo for {{ .alpineRepoName }}-${count.index}"
		}

		resource "artifactory_remote_alpine_repository" "{{ .alpineRepoName }}" {
		  key         = "{{ .alpineRepoName }}-remote"
		  description = "Test remote repo for {{ .alpineRepoName }}"
		  url         = "http://tempurl.org"
		}

		data "artifactory_repositories" "{{ .name }}" {
		  repository_type = "local"
		  package_type    = "alpine"

		  // ensure all repos are created first before fetching
		  depends_on = [
			artifactory_local_generic_repository.{{ .genericRepoName }},
			artifactory_local_alpine_repository.{{ .alpineRepoName }},
			artifactory_remote_alpine_repository.{{ .alpineRepoName }},
		  ]
		}

		output "repo_key_0" {
			value = tolist(data.artifactory_repositories.{{ .name }}.repos)[0].key
		}
	`, params)

	artifactoryURL := acctest.GetArtifactoryUrl(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repos.#", "5"),
					resource.TestCheckResourceAttr(fqrn, "repos.0.key", fmt.Sprintf("%s-local-0", params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.1.key", fmt.Sprintf("%s-local-1", params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.2.key", fmt.Sprintf("%s-local-2", params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.3.key", fmt.Sprintf("%s-local-3", params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.4.key", fmt.Sprintf("%s-local-4", params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.0.type", "LOCAL"),
					resource.TestCheckResourceAttr(fqrn, "repos.0.description", fmt.Sprintf("Test local repo for %s-0", params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.0.url", fmt.Sprintf("%s/artifactory/%s-local-0", artifactoryURL, params["alpineRepoName"])),
					resource.TestCheckResourceAttr(fqrn, "repos.0.package_type", "Alpine"),
					resource.TestCheckOutput("repo_key_0", fmt.Sprintf("%s-local-0", params["alpineRepoName"])),
				),
			},
		},
	})
}
