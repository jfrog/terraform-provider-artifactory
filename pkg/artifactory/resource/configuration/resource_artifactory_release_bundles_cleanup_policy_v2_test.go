package configuration_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccReleaseBundleV2Cleanup_invalid_key(t *testing.T) {
	testCases := []struct {
		key        string
		errorRegex string
	}{
		{key: "1", errorRegex: ".*string length must be at least 3"},
		{key: "ab#", errorRegex: ".*only letters, numbers, underscore and hyphen are allowed"},
		{key: "ab1#", errorRegex: ".*only letters, numbers, underscore and hyphen are allowed"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.key, func(t *testing.T) {
			client := acctest.GetTestResty(t)
			version, err := util.GetArtifactoryVersion(client)
			if err != nil {
				t.Fatal(err)
			}
			valid, err := util.CheckVersion(version, "7.104.2")
			if err != nil {
				t.Fatal(err)
			}
			if !valid {
				t.Skipf("Artifactory version %s is earlier than 7.104.2", version)
			}

			_, _, policyName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2_cleanup_policy")

			temp := `
			resource "artifactory_release_bundle_v2_cleanup_policy" "{{ .policyName }}" {
				key = "{{ .policyKey }}"
				description = "test release bundle cleanup policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				search_criteria = {
					include_all_projects = true
					included_projects = []
					release_bundles = [
					{
						name = "**"
						project_key = "test"
					},
					{
						name = "**"
						project_key = "test2"
					}
					]
					exclude_promoted_environments = [
					"**"
					]
				}
			}`

			config := util.ExecuteTemplate(
				policyName,
				temp,
				map[string]string{
					"policyName": policyName,
					"policyKey":  testCase.key,
				},
			)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      config,
						ExpectError: regexp.MustCompile(testCase.errorRegex),
					},
				},
			})
		})
	}
}

func TestAccReleaseBundleV2Cleanup_full(t *testing.T) {
	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.104.2")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.104.2", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2_cleanup_policy")

	temp := `
		resource "artifactory_release_bundle_v2_cleanup_policy" "{{ .policyName }}" {
			key = "{{ .policyName }}"
			description = "test release bundle cleanup policy"
			cron_expression = "0 0 2 ? * MON-SAT *"
			duration_in_minutes = 60
			enabled = true
			search_criteria = {
				include_all_projects = true
				included_projects = []
				release_bundles = [
				{
					name = "**"
					project_key = "test"
				},
				{
					name = "**"
					project_key = "test2"
				}
				]
				exclude_promoted_environments = [
				"**"
				]
			}
		}`

	updatedTemp := `
		resource "artifactory_release_bundle_v2_cleanup_policy" "{{ .policyName }}" {
			key = "{{ .policyName }}"
			description = "test release bundle cleanup policy"
			cron_expression = "0 0 2 ? * MON-SAT *"
			duration_in_minutes = 60
			enabled = true
			search_criteria = {
				include_all_projects = true
				included_projects = []
				release_bundles = [
				{
					name = "**"
					project_key = "test2"
				},
				{
					name = "**"
					project_key = "test3"
				}
				]
				exclude_promoted_environments = [
				"**"
				]
			}
		}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
		},
	)

	updatedConfig := util.ExecuteTemplate(
		policyName,
		updatedTemp,
		map[string]string{
			"policyName": policyName,
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
		CheckDestroy: testAccReleaseBundleV2CleanupPolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "test release bundle cleanup policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "60"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "item_type", "releaseBundle"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "24"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.exclude_promoted_environments.0", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.include_all_projects", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.release_bundles.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.release_bundles.0.name", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.release_bundles.0.project_key", "test"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.release_bundles.1.name", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.release_bundles.1.project_key", "test2"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_projects.#", "0"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "test release bundle cleanup policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "60"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "item_type", "releaseBundle"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "24"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.exclude_promoted_environments.0", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.include_all_projects", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.release_bundles.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.release_bundles.0.name", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.release_bundles.0.project_key", "test2"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.release_bundles.1.name", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.release_bundles.1.project_key", "test3"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_projects.#", "0"),
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

func testAccReleaseBundleV2CleanupPolicyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		client := acctest.Provider.Meta().(util.ProviderMetadata).Client
		resp, err := client.R().
			SetPathParam("policyKey", rs.Primary.Attributes["key"]).
			Get("artifactory/api/cleanup/bundles/policies/{policyKey}")
		if err != nil {
			return err
		}

		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("error: Resource  Policy %s still exists", rs.Primary.ID)
	}
}
