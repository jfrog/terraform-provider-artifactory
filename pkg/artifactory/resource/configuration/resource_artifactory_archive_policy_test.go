// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configuration_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

// INVALID TEST CASES

func TestAccArchivePolicy_invalid_key_validation(t *testing.T) {
	testCases := []struct {
		key        string
		errorRegex string
	}{
		{key: "1", errorRegex: ".*string length must be at least 3"},
		{key: "ab", errorRegex: ".*string length must be at least 3"},
		{key: "ab#", errorRegex: ".*only letters, numbers, underscore and hyphen are allowed"},
		{key: "ab1#", errorRegex: ".*only letters, numbers, underscore and hyphen are allowed"},
		{key: "test@key", errorRegex: ".*only letters, numbers, underscore and hyphen are allowed"},
		{key: "test key", errorRegex: ".*only letters, numbers, underscore and hyphen are allowed"},
		{key: "test.key", errorRegex: ".*only letters, numbers, underscore and hyphen are allowed"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.key, func(t *testing.T) {
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

			_, _, policyName := testutil.MkNames("test-archive-policy", "artifactory_archive_policy")

			config := fmt.Sprintf(`
			resource "artifactory_archive_policy" "%s" {
				key = "%s"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					created_before_in_months = 1
				}
			}`, policyName, testCase.key)

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

func TestAccArchivePolicy_invalid_condition_combinations(t *testing.T) {
	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	testCases := []struct {
		name        string
		config      string
		expectError bool
		errorRegex  string
	}{
		{
			name: "no_conditions_set",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-no-conditions"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
				}
			}`,
			expectError: true,
			errorRegex:  ".*A policy must use exactly one of the following condition types.*",
		},
		{
			name: "mixed_days_and_months",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-mixed-days-months"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					created_before_in_days = 30
					created_before_in_months = 6
				}
			}`,
			expectError: true,
			errorRegex:  ".*Cannot use both days-based conditions.*",
		},
		{
			name: "time_and_version_based",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-time-version"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					created_before_in_days = 30
					keep_last_n_versions = 10
				}
			}`,
			expectError: true,
			errorRegex:  ".*A policy can only use one type of condition.*",
		},
		{
			name: "time_and_properties_based",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-time-properties"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					created_before_in_days = 30
					included_properties = {
						"build.name" = ["my-app"]
					}
				}
			}`,
			expectError: true,
			errorRegex:  ".*A policy can only use one type of condition.*",
		},
		{
			name: "version_and_properties_based",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-version-properties"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					keep_last_n_versions = 10
					included_properties = {
						"build.name" = ["my-app"]
					}
				}
			}`,
			expectError: true,
			errorRegex:  ".*A policy can only use one type of condition.*",
		},
		{
			name: "all_three_condition_types",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-all-three"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					created_before_in_days = 30
					keep_last_n_versions = 10
					included_properties = {
						"build.name" = ["my-app"]
					}
				}
			}`,
			expectError: true,
			errorRegex:  ".*A policy can only use one type of condition.*",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						ExpectError: regexp.MustCompile(tc.errorRegex),
						PlanOnly:    true,
					},
				},
			})
		})
	}
}

func TestAccArchivePolicy_invalid_zero_values(t *testing.T) {
	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	testCases := []struct {
		name       string
		config     string
		errorRegex string
	}{
		{
			name: "zero_created_before_in_days",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-zero-days"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					created_before_in_days = 0
				}
			}`,
			errorRegex: ".*Time-based conditions must have a value greater than 0.*",
		},
		{
			name: "zero_last_downloaded_before_in_days",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-zero-download-days"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					last_downloaded_before_in_days = 0
				}
			}`,
			errorRegex: ".*Time-based conditions must have a value greater than 0.*",
		},
		{
			name: "zero_created_before_in_months",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-zero-months"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					created_before_in_months = 0
				}
			}`,
			errorRegex: ".*Time-based conditions must have a value greater than 0.*",
		},
		{
			name: "zero_last_downloaded_before_in_months",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-zero-download-months"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					last_downloaded_before_in_months = 0
				}
			}`,
			errorRegex: ".*Time-based conditions must have a value greater than 0.*",
		},
		{
			name: "zero_keep_last_n_versions",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-zero-version"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					keep_last_n_versions = 0
				}
			}`,
			errorRegex: ".*Version-based condition \\(keep_last_n_versions\\) must have a value greater than\\s+0\\. Zero values are not allowed.*",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						ExpectError: regexp.MustCompile(tc.errorRegex),
						PlanOnly:    true,
					},
				},
			})
		})
	}
}

func TestAccArchivePolicy_invalid_properties_validation(t *testing.T) {
	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	testCases := []struct {
		name       string
		config     string
		errorRegex string
	}{
		{
			name: "included_properties_multiple_keys",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-invalid-included-multiple-keys"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					included_properties = {
						"build.name" = ["my-app"]
						"build.number" = ["123"]
					}
				}
			}`,
			errorRegex: ".*Properties-based conditions must have exactly one key.*",
		},
		{
			name: "excluded_properties_multiple_keys",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-invalid-excluded-multiple-keys"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					keep_last_n_versions = 10
					excluded_properties = {
						"build.name" = ["legacy-app"]
						"team" = ["deprecated"]
					}
				}
			}`,
			errorRegex: ".*Properties-based conditions must have exactly one key.*",
		},
		{
			name: "included_properties_multiple_values",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-invalid-included-multiple-values"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					included_properties = {
						"build.name" = ["my-app", "my-service"]
					}
				}
			}`,
			errorRegex: ".*The property value must be a list with exactly one string value.*",
		},
		{
			name: "excluded_properties_multiple_values",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-invalid-excluded-multiple-values"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					keep_last_n_versions = 10
					excluded_properties = {
						"build.name" = ["legacy-app", "old-app"]
					}
				}
			}`,
			errorRegex: ".*The property value must be a list with exactly one string value.*",
		},
		{
			name: "included_properties_empty_list",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-invalid-included-empty-list"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					included_properties = {
						"build.name" = []
					}
				}
			}`,
			errorRegex: ".*The property value must be a list with exactly one string value.*",
		},
		{
			name: "excluded_properties_empty_list",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-invalid-excluded-empty-list"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					keep_last_n_versions = 10
					excluded_properties = {
						"build.name" = []
					}
				}
			}`,
			errorRegex: ".*The property value must be a list with exactly one string value.*",
		},
		{
			name: "included_properties_empty_map",
			config: `
			resource "artifactory_archive_policy" "test" {
				key = "test-invalid-included-no-keys"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					included_properties = {}
				}
			}`,
			errorRegex: ".*A policy must use exactly one of the following condition types.*",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						ExpectError: regexp.MustCompile(tc.errorRegex),
						PlanOnly:    true,
					},
				},
			})
		})
	}
}

func TestAccArchivePolicy_invalid_project_level_policy(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, _, policyName := testutil.MkNames("test-project-policy", "artifactory_archive_policy")
	_, _, projectKey := testutil.MkNames("testproj", "project")

	config := fmt.Sprintf(`
	resource "project" "%s" {
		key = "%s"
		display_name = "%s"
		description  = "Test Project"
		admin_privileges {
			manage_members   = true
			manage_resources = true
			index_resources  = true
		}
		max_storage_in_gibibytes   = 10
		block_deployments_on_limit = false
		email_notification         = true
	}

	resource "artifactory_archive_policy" "%s" {
		key = "%s"
		description = "Test policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		project_key = project.%s.key
		
		search_criteria = {
			package_types = ["docker"]
			repos = ["**"]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			include_all_projects = true
			included_projects = []
			created_before_in_months = 1
		}
	}`, projectKey, projectKey, projectKey, policyName, policyName, projectKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*Cannot include all projects for a project level policy.*"),
			},
		},
	})
}

// VALID TEST CASES

func TestAccArchivePolicy_valid_time_based_days(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, _, policyName := testutil.MkNames("test-days-policy", "artifactory_archive_policy")

	testCases := []struct {
		name         string
		createdDays  int
		downloadDays int
		description  string
	}{
		{
			name:         "single_created_condition",
			createdDays:  30,
			downloadDays: 0,
			description:  "Policy with only created_before_in_days condition",
		},
		{
			name:         "single_download_condition",
			createdDays:  0,
			downloadDays: 60,
			description:  "Policy with only last_downloaded_before_in_days condition",
		},
		{
			name:         "both_conditions",
			createdDays:  30,
			downloadDays: 60,
			description:  "Policy with both days-based conditions",
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policyKey := fmt.Sprintf("%s-%d", policyName, i)

			var createdCondition, downloadCondition string
			if tc.createdDays > 0 {
				createdCondition = fmt.Sprintf("created_before_in_days = %d", tc.createdDays)
			}
			if tc.downloadDays > 0 {
				downloadCondition = fmt.Sprintf("last_downloaded_before_in_days = %d", tc.downloadDays)
			}

			config := fmt.Sprintf(`
			resource "artifactory_archive_policy" "%s" {
				key = "%s"
				description = "%s"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					%s
					%s
				}
			}`, policyKey, policyKey, tc.description, createdCondition, downloadCondition)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				CheckDestroy:             testAccArchivePolicyDestroy(fmt.Sprintf("artifactory_archive_policy.%s", policyKey)),
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(fmt.Sprintf("artifactory_archive_policy.%s", policyKey), "key", policyKey),
							resource.TestCheckResourceAttr(fmt.Sprintf("artifactory_archive_policy.%s", policyKey), "description", tc.description),
							resource.TestCheckResourceAttr(fmt.Sprintf("artifactory_archive_policy.%s", policyKey), "enabled", "false"),
						),
					},
				},
			})
		})
	}
}

func TestAccArchivePolicy_valid_time_based_months(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-months-policy", "artifactory_archive_policy")

	config := fmt.Sprintf(`
	resource "artifactory_archive_policy" "%s" {
		key = "%s"
		description = "Policy with months-based conditions"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			repos = ["**"]
			package_types = ["docker"]
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_months = 6
			last_downloaded_before_in_months = 12
		}
	}`, policyName, policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Policy with months-based conditions"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "6"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "12"),
				),
			},
		},
	})
}

func TestAccArchivePolicy_valid_version_based(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-version-policy", "artifactory_archive_policy")

	config := fmt.Sprintf(`
	resource "artifactory_archive_policy" "%s" {
		key = "%s"
		description = "Policy with version-based condition"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			repos = ["**"]
			package_types = ["docker"]
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			keep_last_n_versions = 10
		}
	}`, policyName, policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Policy with version-based condition"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.keep_last_n_versions", "10"),
				),
			},
		},
	})
}

func TestAccArchivePolicy_valid_properties_based(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, _, policyName := testutil.MkNames("test-properties-policy", "artifactory_archive_policy")

	testCases := []struct {
		name        string
		description string
		config      string
	}{
		{
			name:        "included_properties_only",
			description: "Policy with included_properties only",
			config: `
			included_properties = {
				"build.name" = ["my-app"]
			}`,
		},
		{
			name:        "excluded_properties_with_time_based",
			description: "Policy with excluded_properties and time-based condition",
			config: `
			created_before_in_days = 30
			excluded_properties = {
				"build.name" = ["legacy-app"]
			}`,
		},
		{
			name:        "excluded_properties_with_version_based",
			description: "Policy with excluded_properties and version-based condition",
			config: `
			keep_last_n_versions = 5
			excluded_properties = {
				"team" = ["deprecated"]
			}`,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policyKey := fmt.Sprintf("%s-%d", policyName, i)

			config := fmt.Sprintf(`
			resource "artifactory_archive_policy" "%s" {
				key = "%s"
				description = "%s"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = false
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					%s
				}
			}`, policyKey, policyKey, tc.description, tc.config)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				CheckDestroy:             testAccArchivePolicyDestroy(fmt.Sprintf("artifactory_archive_policy.%s", policyKey)),
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(fmt.Sprintf("artifactory_archive_policy.%s", policyKey), "key", policyKey),
							resource.TestCheckResourceAttr(fmt.Sprintf("artifactory_archive_policy.%s", policyKey), "description", tc.description),
						),
					},
				},
			})
		})
	}
}

func TestAccArchivePolicy_valid_all_package_types(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-all-packages", "artifactory_archive_policy")

	// All supported package types
	allPackageTypes := []string{
		"alpine", "ansible", "cargo", "chef", "cocoapods", "composer", "conan", "conda",
		"debian", "docker", "gems", "generic", "go", "gradle", "helm", "helmoci",
		"huggingfaceml", "machinelearning", "maven", "npm", "nuget", "oci", "opkg",
		"puppet", "pypi", "sbt", "swift", "terraform", "terraformbackend", "vagrant", "yum",
	}

	packageTypesStr := `["` + strings.Join(allPackageTypes, `", "`) + `"]`

	config := fmt.Sprintf(`
	resource "artifactory_archive_policy" "%s" {
		key = "%s"
		description = "Policy with all supported package types"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			repos = ["**"]
			package_types = %s
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_days = 30
		}
	}`, policyName, policyName, packageTypesStr)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Policy with all supported package types"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", fmt.Sprintf("%d", len(allPackageTypes))),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_days", "30"),
				),
			},
		},
	})
}

func TestAccArchivePolicy_valid_with_project_association(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-project-association", "artifactory_archive_policy")
	_, _, repoName := testutil.MkNames("test-docker-local", "artifactory_local_docker_v2_repository")
	_, _, projectKey := testutil.MkNames("testproj", "project")

	config := fmt.Sprintf(`
	resource "artifactory_local_docker_v2_repository" "%s" {
		key             = "%s"
		tag_retention   = 3
		max_unique_tags = 5

		lifecycle {
			ignore_changes = ["project_key", "project_environments"]
		}
	}

	resource "project" "%s" {
		key = "%s"
		display_name = "%s"
		description  = "Test Project"
		admin_privileges {
			manage_members   = true
			manage_resources = true
			index_resources  = true
		}
		max_storage_in_gibibytes   = 10
		block_deployments_on_limit = false
		email_notification         = true
	}

	resource "project_repository" "%s-%s" {
		project_key = project.%s.key
		key = artifactory_local_docker_v2_repository.%s.key
	}

	resource "artifactory_archive_policy" "%s" {
		key = "%s"
		description = "Policy with project association"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker"]
			repos = [project_repository.%s-%s.key]
			included_projects = [project.%s.key]
			include_all_projects = false
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_months = 7
			last_downloaded_before_in_months = 6
		}
	}`, repoName, repoName, projectKey, projectKey, projectKey, projectKey, repoName, projectKey, repoName, policyName, policyName, projectKey, repoName, projectKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		CheckDestroy: testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Policy with project association"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.0", "docker"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.0", repoName),
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

func TestAccArchivePolicy_valid_project_level_policy(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-project-level", "artifactory_archive_policy")
	_, _, repoName := testutil.MkNames("test-docker-local", "artifactory_local_docker_v2_repository")
	_, _, projectKey := testutil.MkNames("testproj", "project")

	config := fmt.Sprintf(`
	resource "artifactory_local_docker_v2_repository" "%s" {
		key             = "%s"
		tag_retention   = 3
		max_unique_tags = 5

		lifecycle {
			ignore_changes = ["project_key", "project_environments"]
		}
	}

	resource "project" "%s" {
		key = "%s"
		display_name = "%s"
		description  = "Test Project"
		admin_privileges {
			manage_members   = true
			manage_resources = true
			index_resources  = true
		}
		max_storage_in_gibibytes   = 10
		block_deployments_on_limit = false
		email_notification         = true
	}

	resource "project_repository" "%s-%s" {
		project_key = project.%s.key
		key = artifactory_local_docker_v2_repository.%s.key
	}

	resource "artifactory_archive_policy" "%s" {
		key = "%s-%s"
		description = "Project-level policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		project_key = project_repository.%s-%s.project_key
		
		search_criteria = {
			package_types = ["docker"]
			repos = [project_repository.%s-%s.key]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			include_all_projects = false
			included_projects = []
			created_before_in_months = 1
			last_downloaded_before_in_months = 6
		}
	}`, repoName, repoName, projectKey, projectKey, projectKey, projectKey, repoName, projectKey, repoName, policyName, projectKey, policyName, projectKey, repoName, projectKey, repoName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		CheckDestroy: testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", fmt.Sprintf("%s-%s", projectKey, policyName)),
					resource.TestCheckResourceAttr(fqrn, "description", "Project-level policy"),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.include_all_projects", "false"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        fmt.Sprintf("%s-%s:%s", projectKey, policyName, projectKey),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccArchivePolicy_valid_update_scenarios(t *testing.T) {
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}

	client := acctest.GetTestResty(t)
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.101.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.101.0", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-update", "artifactory_archive_policy")

	initialConfig := fmt.Sprintf(`
	resource "artifactory_archive_policy" "%s" {
		key = "%s"
		description = "Initial policy description"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			repos = ["**"]
			package_types = ["docker"]
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_days = 30
		}
	}`, policyName, policyName)

	updatedConfig := fmt.Sprintf(`
	resource "artifactory_archive_policy" "%s" {
		key = "%s"
		description = "Updated policy description"
		cron_expression = "0 0 3 ? * MON-SAT *"
		duration_in_minutes = 120
		enabled = true
		skip_trashcan = true
		
		search_criteria = {
			repos = ["**"]
			package_types = ["docker", "maven", "npm"]
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest", "org/deprecated"]
			created_before_in_days = 60
			last_downloaded_before_in_days = 90
		}
	}`, policyName, policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Initial policy description"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "60"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "false"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "1"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Updated policy description"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "120"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "3"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "2"),
				),
			},
		},
	})
}

func TestAccArchivePolicy_with_variable_last_downloaded_before_in_days(t *testing.T) {
	client := acctest.GetTestResty(t)
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.111.2")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.111.2", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-archive-policy", "artifactory_archive_policy")

	temp := `
	variable "archive_policy_last_downloaded_before_in_days" {
		type = number
		default = 10
	}

	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy with variable for last_downloaded_before_in_days"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker", "generic", "helm", "helmoci", "nuget", "terraform"]
			repos = ["**"]
			include_all_projects = false
			included_projects = ["default"]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			last_downloaded_before_in_days = var.archive_policy_last_downloaded_before_in_days
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy with variable for last_downloaded_before_in_days"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_days", "10"),
				),
			},
		},
	})
}

func TestAccArchivePolicy_with_variable_created_before_in_days(t *testing.T) {
	client := acctest.GetTestResty(t)
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.111.2")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.111.2", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-archive-policy", "artifactory_archive_policy")

	temp := `
	variable "archive_policy_created_before_in_days" {
		type = number
		default = 45
	}

	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy with variable for created_before_in_days"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker", "generic", "helm", "helmoci", "nuget", "terraform"]
			repos = ["**"]
			include_all_projects = false
			included_projects = ["default"]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_days = var.archive_policy_created_before_in_days
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy with variable for created_before_in_days"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_days", "45"),
				),
			},
		},
	})
}

func TestAccArchivePolicy_with_variable_keep_last_n_versions(t *testing.T) {
	client := acctest.GetTestResty(t)
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.111.2")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.111.2", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-archive-policy", "artifactory_archive_policy")

	temp := `
	variable "archive_policy_keep_last_n_versions" {
		type = number
		default = 5
	}

	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy with variable for keep_last_n_versions"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker", "helm", "helmoci", "nuget", "maven", "npm"]
			repos = ["**"]
			include_all_projects = false
			included_projects = ["default"]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			keep_last_n_versions = var.archive_policy_keep_last_n_versions
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy with variable for keep_last_n_versions"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.keep_last_n_versions", "5"),
				),
			},
		},
	})
}

func TestAccArchivePolicy_with_variable_duration_in_minutes(t *testing.T) {
	client := acctest.GetTestResty(t)
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.102.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.102.0", version)
	}

	_, fqrn, policyName := testutil.MkNames("test-archive-policy", "artifactory_archive_policy")

	temp := `
	variable "archive_policy_duration_in_minutes" {
		type = number
		default = 120
	}

	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy with variable for duration_in_minutes"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = var.archive_policy_duration_in_minutes
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker", "generic", "helm", "helmoci", "nuget", "terraform"]
			repos = ["**"]
			include_all_projects = false
			included_projects = ["default"]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_days = 30
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccArchivePolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy with variable for duration_in_minutes"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "120"),
				),
			},
		},
	})
}

func TestAccArchivePolicy_with_variable_no_default_should_fail(t *testing.T) {
	client := acctest.GetTestResty(t)
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.111.2")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.111.2", version)
	}

	_, _, policyName := testutil.MkNames("test-archive-policy", "artifactory_archive_policy")

	temp := `
	variable "archive_policy_last_downloaded_before_in_days" {
		type = number
		# No default - should require value
	}

	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy with variable without default"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker", "generic", "helm", "helmoci", "nuget", "terraform"]
			repos = ["**"]
			include_all_projects = false
			included_projects = ["default"]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			last_downloaded_before_in_days = var.archive_policy_last_downloaded_before_in_days
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*(No value for required variable|Missing required argument|Required variable not set|variable.*must be set).*"),
				PlanOnly:    true,
			},
		},
	})
}

func TestAccArchivePolicy_with_variable_duration_in_minutes_no_default_should_fail(t *testing.T) {
	client := acctest.GetTestResty(t)
	archivePolicyEnabled := os.Getenv("JFROG_ARCHIVE_POLICY_ENABLED")
	if strings.ToLower(archivePolicyEnabled) != "true" {
		t.Skipf("JFROG_ARCHIVE_POLICY_ENABLED env var is not set to 'true'")
	}
	version, err := util.GetArtifactoryVersion(client)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := util.CheckVersion(version, "7.102.0")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Skipf("Artifactory version %s is earlier than 7.102.0", version)
	}

	_, _, policyName := testutil.MkNames("test-archive-policy", "artifactory_archive_policy")

	temp := `
	variable "archive_policy_duration_in_minutes" {
		type = number
		# No default - should require value
	}

	resource "artifactory_archive_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy with variable for duration_in_minutes without default"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = var.archive_policy_duration_in_minutes
		enabled = false
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker", "generic", "helm", "helmoci", "nuget", "terraform"]
			repos = ["**"]
			include_all_projects = false
			included_projects = ["default"]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_days = 30
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*(No value for required variable|Missing required argument|Required variable not set|variable.*must be set).*"),
				PlanOnly:    true,
			},
		},
	})
}

func testAccArchivePolicyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		client := acctest.Provider.Meta().(util.ProviderMetadata).Client
		resp, err := client.R().
			SetPathParam("policyKey", rs.Primary.Attributes["key"]).
			Get("artifactory/api/archive/v2/packages/policies/{policyKey}")
		if err != nil {
			return err
		}

		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("error: Archive Policy %s still exists", rs.Primary.ID)
	}
}
