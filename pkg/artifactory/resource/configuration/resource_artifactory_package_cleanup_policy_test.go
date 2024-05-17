package configuration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccPackageCleanupPolicy_full(t *testing.T) {
	_, fqrn, policyName := testutil.MkNames("test-package-cleanup-policy", "artifactory_package_cleanup_policy")
	_, fqrn, repoName := testutil.MkNames("test-docker-local", "artifactory_local_docker_v2_repository")

	temp := `
	resource "artifactory_local_docker_v2_repository" "{{ .repoName }}" {
		key             = "{{ .repoName }}"
		tag_retention   = 3
		max_unique_tags = 5
	}

	resource "artifactory_package_cleanup_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = true
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker"]
			repos = [artifactory_local_docker_v2_repository.{{ .repoName }}.key]
			include_packages = ["com/jfrog"]
			exclude_packages = ["com/jfrog/latest"]
			created_before_in_months = 1
			last_downloaded_before_in_months = 6
		}
	}`

	updatedTemp := `
	resource "artifactory_local_docker_v2_repository" "{{ .repoName }}" {
		key             = "{{ .repoName }}"
		tag_retention   = 3
		max_unique_tags = 5
	}

	resource "artifactory_package_cleanup_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 120
		enabled = false
		skip_trashcan = true
		
		search_criteria = {
			package_types = ["docker"]
			repos = [artifactory_local_docker_v2_repository.{{ .repoName }}.key]
			include_packages = ["com/jfrog", "foo"]
			exclude_packages = ["com/jfrog/latest"]
			created_before_in_months = 12
			last_downloaded_before_in_months = 24
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
			"repoName":   repoName,
		})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccPolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "60"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "false"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.0", "docker"),
					resource.TestCheckResourceAttr(fqrn, "repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repos.0", repoName),
					resource.TestCheckResourceAttr(fqrn, "include_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "include_packages.0", "com/jfrog"),
					resource.TestCheckResourceAttr(fqrn, "exclude_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "exclude_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "created_before_in_months", "1"),
					resource.TestCheckResourceAttr(fqrn, "last_downloaded_before_in_months", "6"),
				),
			},
			{
				Config: updatedTemp,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "120"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.0", "docker"),
					resource.TestCheckResourceAttr(fqrn, "repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repos.0", repoName),
					resource.TestCheckResourceAttr(fqrn, "include_packages.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "include_packages.0", "com/jfrog"),
					resource.TestCheckResourceAttr(fqrn, "include_packages.1", "foo"),
					resource.TestCheckResourceAttr(fqrn, "exclude_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "exclude_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "created_before_in_months", "12"),
					resource.TestCheckResourceAttr(fqrn, "last_downloaded_before_in_months", "24"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPolicyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		rs, ok := s.RootModule().Resources["artifactory_package_cleanup_policy."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		resp, err := client.R().
			SetPathParam("policyKey", rs.Primary.Attributes["key"]).
			Get(configuration.PackageCleanupPolicyEndpointPath)
		if err != nil {
			return err
		}

		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("error: Package Cleanup Policy %s still exists", rs.Primary.ID)
	}
}
