package configuration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccPackageCleanupPolicy_full(t *testing.T) {
	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.90.1")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.90.1", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-package-cleanup-policy", "artifactory_package_cleanup_policy")
	_, _, repoName := testutil.MkNames("test-docker-local", "artifactory_local_docker_v2_repository")

	temp := `
	resource "artifactory_local_docker_v2_repository" "{{ .repoName }}" {
		key             = "{{ .repoName }}"
		tag_retention   = 3
		max_unique_tags = 5
	}

	resource "project" "myproject" {
		key = "myproj"
		display_name = "My Project"
		description  = "My Project"
		admin_privileges {
			manage_members   = true
			manage_resources = true
			index_resources  = true
		}
		max_storage_in_gibibytes   = 10
		block_deployments_on_limit = false
		email_notification         = true
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
			included_projects = [project.myproject.key]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
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
			package_types = ["docker", "maven", "gradle"]
			include_all_repos = true
			included_packages = ["com/jfrog", "**"]
			excluded_packages = ["com/jfrog/latest"]
			include_all_projects = true
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
		},
	)

	updatedConfig := util.ExecuteTemplate(
		policyName,
		updatedTemp,
		map[string]string{
			"policyName": policyName,
			"repoName":   repoName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		CheckDestroy: testAccPolicyDestroy(fqrn),
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
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.0", repoName),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.0", "com/jfrog"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "6"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "120"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "3"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "docker"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "maven"),
					resource.TestCheckTypeSetElemAttr(fqrn, "search_criteria.package_types.*", "gradle"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.0", repoName),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.0", "com/jfrog"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.1", "foo"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "12"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "24"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        policyName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func testAccPolicyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		client := acctest.Provider.Meta().(util.ProviderMetadata).Client
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
