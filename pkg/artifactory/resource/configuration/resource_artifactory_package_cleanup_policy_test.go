package configuration_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccPackageCleanupPolicy_migrate_schema_v0(t *testing.T) {
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
			include_all_projects = false
			included_projects = ["default"]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_months = 1
			last_downloaded_before_in_months = 6
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCleanupPolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "11.8.0",
					},
				},
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
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.0", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.include_all_projects", "false"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_projects.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_projects.0", "default"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "6"),
				),
			},
			{
				Config:                   config,
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "60"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "false"),
					resource.TestCheckNoResourceAttr(fqrn, "project_key"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.0", "docker"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.0", repoName),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.0", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.include_all_projects", "false"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_projects.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_projects.0", "default"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "6"),
				),
			},
		},
	})
}

func TestAccPackageCleanupPolicy_invalid_key(t *testing.T) {
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
			valid, err := util.CheckVersion(version, "7.90.1")
			if err != nil {
				t.Fatal(err)
			}
			if !valid {
				t.Skipf("Artifactory version %s is earlier than 7.90.1", version)
			}

			_, _, policyName := testutil.MkNames("test-package-cleanup-policy", "artifactory_package_cleanup_policy")

			temp := `
			resource "artifactory_package_cleanup_policy" "{{ .policyName }}" {
				key = "{{ .policyKey }}"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					repos = ["**"]
					package_types = ["docker"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					excluded_packages = ["com/jfrog/latest"]
					created_before_in_months = 1
					last_downloaded_before_in_months = 6
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

func TestAccPackageCleanupPolicy_validation_comprehensive(t *testing.T) {
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

	testCases := []struct {
		name        string
		config      string
		expectError bool
		errorRegex  string
	}{
		{
			name: "valid time-based conditions (months)",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-valid-time-months"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					created_before_in_months = 12
					last_downloaded_before_in_months = 6
				}
			}`,
			expectError: false,
		},
		{
			name: "valid time-based conditions (days)",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-valid-time-days"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					created_before_in_days = 365
					last_downloaded_before_in_days = 180
				}
			}`,
			expectError: false,
		},
		{
			name: "valid version-based condition",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-valid-version"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["maven"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					keep_last_n_versions = 5
				}
			}`,
			expectError: false,
		},
		{
			name: "valid properties-based condition",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-valid-properties"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					included_properties = {
						"test_key" = ["test_value"]
					}
				}
			}`,
			expectError: false,
		},
		{
			name: "invalid mixed time-based conditions (days and months)",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-invalid-mixed"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					created_before_in_months = 12
					created_before_in_days = 365
				}
			}`,
			expectError: true,
			errorRegex:  "Cannot use both days-based conditions",
		},
		{
			name: "invalid mixed condition types (time and version)",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-invalid-mixed-types"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					created_before_in_months = 12
					keep_last_n_versions = 5
				}
			}`,
			expectError: true,
			errorRegex:  "A policy can only use one type of condition",
		},
		{
			name: "invalid mixed condition types (time and properties)",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-invalid-time-props"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					created_before_in_months = 12
					included_properties = {
						"test_key" = ["test_value"]
					}
				}
			}`,
			expectError: true,
			errorRegex:  "A policy can only use one type of condition",
		},
		{
			name: "invalid mixed condition types (version and properties)",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-invalid-version-props"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					keep_last_n_versions = 5
					included_properties = {
						"test_key" = ["test_value"]
					}
				}
			}`,
			expectError: true,
			errorRegex:  "A policy can only use one type of condition",
		},
		{
			name: "invalid zero value for time-based condition (months)",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-invalid-zero-months"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					created_before_in_months = 0
				}
			}`,
			expectError: true,
			errorRegex:  "Time-based conditions must have a value greater than 0",
		},
		{
			name: "invalid zero value for time-based condition (days)",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-invalid-zero-days"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					created_before_in_days = 0
				}
			}`,
			expectError: true,
			errorRegex:  "Time-based conditions must have a value greater than 0",
		},
		{
			name: "invalid zero value for version-based condition",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-invalid-zero-version"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					keep_last_n_versions = 0
				}
			}`,
			expectError: true,
			errorRegex:  ".*Version-based condition \\(keep_last_n_versions\\) must have a value greater than\\s+0\\. Zero values are not allowed.*",
		},
		{
			name: "invalid properties with multiple keys",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-invalid-props-multi"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					included_properties = {
						"test_key1" = ["test_value1"]
						"test_key2" = ["test_value2"]
					}
				}
			}`,
			expectError: true,
			errorRegex:  "Properties-based conditions must have exactly one key",
		},
		{
			name: "invalid properties with multiple values",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-invalid-props-multi-val"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
					included_properties = {
						"test_key" = ["test_value1", "test_value2"]
					}
				}
			}`,
			expectError: true,
			errorRegex:  "The property value must be a list with exactly one string value",
		},
		{
			name: "no condition specified",
			config: `
			resource "artifactory_package_cleanup_policy" "test" {
				key = "test-no-condition"
				description = "Test policy"
				cron_expression = "0 0 2 ? * MON-SAT *"
				duration_in_minutes = 60
				enabled = true
				skip_trashcan = false
				
				search_criteria = {
					package_types = ["docker"]
					repos = ["**"]
					include_all_projects = true
					included_projects = []
					included_packages = ["**"]
				}
			}`,
			expectError: true,
			errorRegex:  "A policy must use exactly one of the following condition types",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tc.config,
						ExpectError: func() *regexp.Regexp {
							if tc.expectError {
								return regexp.MustCompile(tc.errorRegex)
							}
							return nil
						}(),
						Check: func(s *terraform.State) error {
							if tc.expectError {
								return nil
							}
							// For valid configurations, just verify the resource was created
							return nil
						},
					},
				},
			})
		})
	}
}

func TestAccPackageCleanupPolicy_all_package_types(t *testing.T) {
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

	// All supported package types as a valid HCL list
	//allPackageTypes := "[\"alpine\", \"ansible\", \"chef\", \"cargo\", \"composer\" \"cocoapods\", \"conan\", \"conda\", \"debian\", \"docker\", \"gems\", \"generic\", \"go\", \"gradle\", \"helm\", \"helmoci\", \"huggingfaceml\", \"machinelearning\", \"maven\", \"npm\", \"nuget\", \"oci\", \"puppet\", \"pypi\", \"sbt\", \"swift\", \"terraform\", \"terraformbackend\", \"yum\"]"
	allPackageTypes := "[\"alpine\", \"ansible\", \"cargo\", \"cocoapods\", \"conan\", \"conda\", \"debian\", \"docker\", \"gems\", \"generic\", \"go\", \"gradle\", \"helm\", \"helmoci\", \"huggingfaceml\", \"machinelearning\", \"maven\", \"npm\", \"nuget\", \"oci\", \"pypi\", \"sbt\", \"terraform\", \"terraformbackend\", \"yum\"]"
	temp := `
	resource "artifactory_package_cleanup_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy with all package types"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = true
		skip_trashcan = false
		
		search_criteria = {
			package_types = {{ .packageTypes }}
			repos = ["**"]
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			created_before_in_months = 12
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName":   policyName,
			"packageTypes": allPackageTypes,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccCleanupPolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy with all package types"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "25"),
				),
			},
		},
	})
}

func TestAccPackageCleanupPolicy_days_based_conditions(t *testing.T) {
	client := acctest.GetTestResty(t)
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

	_, fqrn, policyName := testutil.MkNames("test-package-cleanup-policy", "artifactory_package_cleanup_policy")

	temp := `
	resource "artifactory_package_cleanup_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy with days-based conditions"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = true
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker"]
			repos = ["**"]
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_days = 30
			last_downloaded_before_in_days = 60
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
		CheckDestroy:             testAccCleanupPolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy with days-based conditions"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_days", "30"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_days", "60"),
				),
			},
		},
	})
}

func TestAccPackageCleanupPolicy_included_properties(t *testing.T) {
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

	temp := `
	resource "artifactory_package_cleanup_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy with included properties"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = true
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker"]
			repos = ["**"]
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			included_properties = {
				"test_key" = ["test_value"]
			}
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
		CheckDestroy:             testAccCleanupPolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy with included properties"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_properties.test_key.0", "test_value"),
				),
			},
		},
	})
}

func TestAccPackageCleanupPolicy_excluded_properties(t *testing.T) {
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

	temp := `
	resource "artifactory_package_cleanup_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy with excluded properties"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = true
		skip_trashcan = false
		
		search_criteria = {
			package_types = ["docker"]
			repos = ["**"]
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			excluded_properties = {
				"test_key" = ["test_value"]
			}
			created_before_in_months = 1
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
		CheckDestroy:             testAccCleanupPolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy with excluded properties"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_properties.test_key.0", "test_value"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "1"),
				),
			},
		},
	})
}

func TestAccPackageCleanupPolicy_invalid_conditions(t *testing.T) {
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

	_, _, policyName := testutil.MkNames("test-package-cleanup-policy", "artifactory_package_cleanup_policy")

	temp := `
	resource "artifactory_package_cleanup_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = true
		skip_trashcan = false
		
		search_criteria = {
			repos = ["**"]
			package_types = ["docker"]
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_months = 0
			last_downloaded_before_in_months = 0
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
				ExpectError: regexp.MustCompile("Time-based conditions must have a value greater than 0"),
			},
		},
	})
}

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
			include_all_projects = true
			included_projects = []
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			created_before_in_months = 1
			last_downloaded_before_in_months = 6
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccCleanupPolicyDestroy(fqrn),
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
					resource.TestCheckNoResourceAttr(fqrn, "project_key"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.0", "docker"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.0", repoName),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.0", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.include_all_projects", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "6"),
				),
			},
		},
	})
}

func TestAccPackageCleanupPolicy_with_project_key(t *testing.T) {
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
	_, _, projectKey := testutil.MkNames("testproj", "project")

	temp := `
	resource "project" "{{ .projectKey }}" {
		key = "{{ .projectKey }}"
		display_name = "Test Project"
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

	resource "artifactory_local_docker_v2_repository" "test" {
		key                 = "{{ .projectKey }}-{{ .repoName }}"
		tag_retention       = 3
		max_unique_tags     = 5
		project_key         = project.{{ .projectKey }}.key
		project_environments = ["DEV"]
	}

	resource "artifactory_package_cleanup_policy" "{{ .policyName }}" {
		key = "{{ .policyName }}"
		description = "Test policy"
		cron_expression = "0 0 2 ? * MON-SAT *"
		duration_in_minutes = 60
		enabled = false
		skip_trashcan = false
		search_criteria = {
			package_types = ["docker"]
			repos = [artifactory_local_docker_v2_repository.test.key]
			included_packages = ["**"]
			excluded_packages = ["com/jfrog/latest"]
			include_all_projects = false
			included_projects = [ project.{{ .projectKey }}.key ]
			created_before_in_months = 1
			last_downloaded_before_in_months = 6
		}
	}`

	config := util.ExecuteTemplate(
		policyName,
		temp,
		map[string]string{
			"policyName": policyName,
			"repoName":   repoName,
			"projectKey": projectKey,
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
		CheckDestroy: testAccCleanupPolicyDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", policyName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test policy"),
					resource.TestCheckResourceAttr(fqrn, "cron_expression", "0 0 2 ? * MON-SAT *"),
					resource.TestCheckResourceAttr(fqrn, "duration_in_minutes", "60"),
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "skip_trashcan", "false"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.package_types.0", "docker"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.repos.0", fmt.Sprintf("%s-%s", projectKey, repoName)),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_packages.0", "**"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.excluded_packages.0", "com/jfrog/latest"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.include_all_projects", "false"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.included_projects.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.created_before_in_months", "1"),
					resource.TestCheckResourceAttr(fqrn, "search_criteria.last_downloaded_before_in_months", "6"),
				),
			},
		},
	})
}

func testAccCleanupPolicyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		client := acctest.GetTestResty(nil)
		version, err := util.GetArtifactoryVersion(client)
		if err != nil {
			return err
		}
		valid, err := util.CheckVersion(version, "7.90.1")
		if err != nil {
			return err
		}
		if !valid {
			return nil
		}

		response, err := client.R().
			SetPathParam("policyKey", rs.Primary.ID).
			Get(configuration.PackageCleanupPolicyEndpointPath)

		if err != nil {
			return err
		}

		if response.StatusCode() == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("error: Package cleanup policy %s still exists", rs.Primary.ID)
	}
}
