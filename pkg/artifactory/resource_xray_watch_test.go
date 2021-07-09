package artifactory

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const basicWatch = `
resource "artifactory_xray_watch" "watch" {
	name        = "basic-watch"
	description = "basic watch"
	active      = false
}
`

func TestAccWatch_basicWatch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_xray_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: basicWatch,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "name", "basic-watch"),
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "description", "basic watch"),
				),
			},
			{
				ResourceName:      "artifactory_xray_watch.watch",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const basicInitialWatchWithPolicy = `
resource "artifactory_local_repository" "repo1" {
	key                  = "watch-test-repo1"
	package_type         = "generic"
	repo_layout_ref      = "simple-default"
	xray_index 			 = true
}
  
resource "artifactory_local_repository" "repo2" {
	key                  = "watch-test-repo2"
	package_type         = "generic"
	repo_layout_ref      = "simple-default"
	xray_index 			 = true
}

resource "artifactory_xray_policy" "test" {
	name  = "basic-initial-policy"
	description = "watch-test"
	type = "security"

	rules {
		name = "rule1"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}
  
resource "artifactory_xray_watch" "watch" {
	name        = "named-local-repo-test1"
	description = "all repositories"
	active = false
  
	repository {
	  name = artifactory_local_repository.repo1.key
	  repo_type = "local"
	}

	policy {
		name = artifactory_xray_policy.test.id
		type = "security"
	}
}
`

const basicWatchWithPolicyAdditionalRepo = `
resource "artifactory_local_repository" "repo1" {
	key                  = "watch-test-repo1"
	package_type         = "generic"
	repo_layout_ref      = "simple-default"
	xray_index 			 = true
}
  
resource "artifactory_local_repository" "repo2" {
	key              = "watch-test-repo2"
	package_type     = "generic"
	repo_layout_ref  = "simple-default"
	xray_index 			 = true
}

resource "artifactory_xray_policy" "test" {
	name  = "basic-initial-policy"
	description = "watch-test"
	type = "security"

	rules {
		name = "rule1"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "artifactory_xray_watch" "watch" {
	name        = "named-local-repo-test1"
	description = "all repositories"
	active = false
  
	repository {
		name = artifactory_local_repository.repo1.key
		repo_type = "local"
	}

	repository {
		name = artifactory_local_repository.repo2.key
		repo_type = "local"
	}

	policy {
		name = artifactory_xray_policy.test.name
		type = "security"
	}
}
`

func TestAccWatch_basicWatchWithPolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_xray_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: basicInitialWatchWithPolicy,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "name", "named-local-repo-test1"),
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "description", "all repositories"),
				),
			},
			{
				Config: basicWatchWithPolicyAdditionalRepo,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "name", "named-local-repo-test1"),
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "description", "all repositories"),
				),
			},
			{
				ResourceName:      "artifactory_xray_watch.watch",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const basicWatchWithPolicies = `
resource "artifactory_xray_policy" "security" {
	name  = "basic-with-policies-1"
	description = "watch-test"
	type = "security"

	rules {
		name = "rule1"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "artifactory_xray_policy" "license" {
	name  = "basic-with-policies-2"
	description = "watch-test"
	type = "license"

	rules {
		name = "rule1"
		priority = 1
		criteria {
			allowed_licenses = ["BSD-4-Clause"]
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "artifactory_xray_watch" "watch" {
	name        = "basic-watch-with-policies"
	description = "basic watch"

	all_repositories {}

	policy {
		name = artifactory_xray_policy.security.name
		type = "security"
	}

	policy {
		name = artifactory_xray_policy.license.name
		type = "license"
	}
}
`

func TestAccWatch_basicWatchWithPolicies(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_xray_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: basicWatchWithPolicies,
				// ExpectError: regexp.MustCompile("Got invalid watch: policy example-2 doen't exist"),
			},
			{
				ResourceName:      "artifactory_xray_watch.watch",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const badPolicy = `
resource "artifactory_xray_watch" "watch" {
	name        = "basic-watch"
	description = "basic watch"

	all_repositories {}

	policy {
		name = "example"
		type = "bad"
	}
}
`

func TestAccWatch_badPolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchMissing("artifactory_xray_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      badPolicy,
				ExpectError: regexp.MustCompile("policy type bad must be security or license"),
			},
		},
	})
}

const allRepositories = `
resource "artifactory_xray_policy" "test" {
	name  = "all-repositories"
	description = "watch-test"
	type = "security"

	rules {
		name = "rule1"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "artifactory_xray_watch" "watch" {
	name        = "all-repositories"
	description = "all repositories"

	all_repositories {
        package_types = ["NuGet", "Docker"]
        paths = ["path/*"]
        names = ["name2", "name1"]
        mime_types = ["application/zip"]

        property {
            key = "field4"
            value = "value 4"
        }
        property {
            key = "field2"
            value = "value 2"
        }
	}
	
	repository_paths {
        include_patterns = [
			"path1/**",
			"another-path/**",
        ]
        exclude_patterns = [
			"path1/ignore/**",
			"another-path/ignore**",
        ]
	}
	
	policy {
		name = artifactory_xray_policy.test.id
		type = "security"
	}
}
`

func TestAccWatch_allRepositories(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_xray_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: allRepositories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "name", "all-repositories"),
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "description", "all repositories"),

					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository_paths.\d+.include_patterns.\d+`, []string{"another-path/**", "path1/**"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository_paths.\d+.exclude_patterns.\d+`, []string{"another-path/ignore**", "path1/ignore/**"}),

					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `all_repositories.\d+.package_types.\d+`, []string{"Docker", "NuGet"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `all_repositories.\d+.paths.\d+`, []string{"path/*"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `all_repositories.\d+.names.\d+`, []string{"name1", "name2"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `all_repositories.\d+.mime_types.\d+`, []string{"application/zip"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `all_repositories.\d+.property.\d+.key`, []string{"field2", "field4"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `all_repositories.\d+.property.\d+.value`, []string{"value 2", "value 4"}),
				),
			},
			{
				ResourceName:      "artifactory_xray_watch.watch",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const namedLocalRepository = `
resource "artifactory_xray_policy" "test" {
	name  = "named-repository"
	description = "watch-test"
	type = "security"

	rules {
		name = "rule1"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "artifactory_local_repository" "example" {
	key 	     = "named-local-repo"
	package_type = "generic"
	xray_index   = true
}

resource "artifactory_xray_watch" "watch" {
	name        = "named-local-repo"
	description = "local repo"

	repository {
		name = artifactory_local_repository.example.key
		repo_type = "local"
        package_types = ["Generic"]
        paths = ["path/*"]
        mime_types = ["application/zip"]

        property {
            key = "field1"
            value = "value 1"
        }
        property {
            key = "field2"
            value = "value 2"
        }
	}

	repository_paths {
        include_patterns = [
            "path1/**"
        ]
        exclude_patterns = [
            "path1/ignore/**"
        ]
	}

	policy {
		name = artifactory_xray_policy.test.id
		type = "security"
	}
}
`

func TestAccWatch_namedLocalRepository(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_xray_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: namedLocalRepository,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "name", "named-local-repo"),
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "description", "local repo"),

					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository_paths.\d+.include_patterns.\d+`, []string{"path1/**"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository_paths.\d+.exclude_patterns.\d+`, []string{"path1/ignore/**"}),

					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.name`, []string{"named-local-repo"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.package_types.\d+`, []string{"Generic"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.paths.\d+`, []string{"path/*"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.mime_types.\d+`, []string{"application/zip"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.property.\d+.key`, []string{"field1", "field2"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.property.\d+.value`, []string{"value 1", "value 2"}),
				),
			},
			{
				ResourceName:      "artifactory_xray_watch.watch",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const namedRemoteRepository = `
resource "artifactory_xray_policy" "test" {
	name  = "all-repositories"
	description = "watch-test"
	type = "security"

	rules {
		name = "rule1"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "artifactory_remote_repository" "example" {
	key             = "remote-repo"
    package_type    = "npm"
	url             = "https://registry.npmjs.org/"
	repo_layout_ref = "npm-default"
	xray_index      = true

	content_synchronisation {
		enabled = false
	}
}

resource "artifactory_xray_watch" "watch" {
	name        = "named-remote-repo"
	description = "remote repo"

	repository {
		name = artifactory_remote_repository.example.key
		repo_type = "remote"
        package_types = ["Generic"]
        paths = ["path/*"]
        mime_types = ["application/zip"]

        property {
            key = "field1"
            value = "value 1"
        }
        property {
            key = "field2"
            value = "value 2"
        }
	}

	repository_paths {
        include_patterns = [
            "path1/**"
        ]
        exclude_patterns = [
            "path1/ignore/**"
        ]
	}

	policy {
		name = artifactory_xray_policy.test.id
		type = "security"
	}
}
`

func TestAccWatch_namedRemoteRepository(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_xray_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: namedRemoteRepository,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "name", "named-remote-repo"),
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "description", "remote repo"),

					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository_paths.\d+.include_patterns.\d+`, []string{"path1/**"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository_paths.\d+.exclude_patterns.\d+`, []string{"path1/ignore/**"}),

					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.name`, []string{"remote-repo"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.package_types.\d+`, []string{"Generic"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.paths.\d+`, []string{"path/*"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.mime_types.\d+`, []string{"application/zip"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.property.\d+.key`, []string{"field1", "field2"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `repository.\d+.property.\d+.value`, []string{"value 1", "value 2"}),
				),
			},
			{
				ResourceName:      "artifactory_xray_watch.watch",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const allBuilds = `
resource "artifactory_xray_policy" "test" {
	name  = "all-builds"
	description = "watch-test"
	type = "security"

	rules {
		name = "rule1"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "artifactory_xray_watch" "watch" {
    name = "all_builds"
    description = "all_builds"
    active = true

    all_builds {
        bin_mgr_id = "default"
	}

	policy {
		name = artifactory_xray_policy.test.id
		type = "security"
	}
}
`

func TestAccWatch_allBuilds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_xray_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: allBuilds,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "name", "all_builds"),
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "description", "all_builds"),

					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `all_builds.\d+.bin_mgr_id`, []string{"default"}),
				),
			},
			{
				ResourceName:      "artifactory_xray_watch.watch",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const buildByName = `
resource "artifactory_xray_policy" "test" {
	name  = "build-by-name"
	description = "watch-test"
	type = "security"

	rules {
		name = "rule1"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "artifactory_xray_watch" "watch" {
    name = "named_builds"
    description = "named_builds"
    active = true

    build {
        name = "build1"
        bin_mgr_id = "default"
	}

	policy {
		name = artifactory_xray_policy.test.id
		type = "security"
	}
}
`

// There isn't a resource to create a build and index it
// Therefore, we expect this test to fail because xray cannot watch a non-existant and unindexed build.
func TestAccWatch_buildByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchMissing("artifactory_xray_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      buildByName,
				ExpectError: regexp.MustCompile("Got invalid watch: build build1 doesn't exist"),
			},
		},
	})
}

const buildByPattern = `
resource "artifactory_xray_policy" "test" {
	name  = "build-by-pattern"
	description = "watch-test"
	type = "security"

	rules {
		name = "rule1"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "artifactory_xray_watch" "watch" {
    name = "pattern_builds"
    description = "pattern_builds"
    active = true

    all_builds {
        include_patterns = ["hello/**", "apache/**"]
        exclude_patterns = ["apache/bad*", "world/**"]
        bin_mgr_id = "default"        
	}
	
	policy {
		name = artifactory_xray_policy.test.id
		type = "security"
	}
}
`

func TestAccWatch_buildByPattern(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_xray_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: buildByPattern,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "name", "pattern_builds"),
					resource.TestCheckResourceAttr("artifactory_xray_watch.watch", "description", "pattern_builds"),

					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `all_builds.\d+.bin_mgr_id`, []string{"default"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `all_builds.\d+.include_patterns.\d+`, []string{"apache/**", "hello/**"}),
					testAccCheckWatchAttributes("artifactory_xray_watch.watch", `all_builds.\d+.exclude_patterns.\d+`, []string{"apache/bad*", "world/**"}),
				),
			},
			{
				ResourceName:      "artifactory_xray_watch.watch",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// This function helps to approximately assert against the state as there isn't a way to access map objects by index
// We use a regex to search in the keys in the state file to find the values.
// The keys are in the format something.<number> = value
// For example to assert on a list of exclude_patterns,
// The data looks like:
// - all_builds.0.exclude_patterns.2959604672 = apache/bad*
// - all_builds.0.exclude_patterns.9997913 = world/**
// The searchString would be "all_builds.\d+.exclude_patterns.\\d+"
// The expectedValues would be []string{"apache/bad*", "world/**"}),
func testAccCheckWatchAttributes(id string, searchString string, expectedValues []string) func(*terraform.State) error {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		r := regexp.MustCompile(searchString)
		found := false
		matches := []string{}

		for key, realValue := range rs.Primary.Attributes {
			if r.MatchString(key) {
				found = true
				matches = append(matches, realValue)
			}
		}

		if !found {
			return fmt.Errorf("error: element %s not found in state", searchString)
		}

		sort.Strings(matches)
		sort.Strings(expectedValues)

		if reflect.DeepEqual(matches, expectedValues) {
			return nil
		}

		return fmt.Errorf("error: expected values do not match real values, %s, %s", matches, expectedValues)
	}
}

func testAccCheckWatchDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.JfrogXray
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		w, err := client.GetWatch(rs.Primary.ID)

		if w == nil || strings.Contains(err.Error(), "404") {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Watch %s still exists", rs.Primary.ID)
		}
	}
}

func testAccCheckWatchMissing(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[id]

		if !ok {
			return nil
		}

		return fmt.Errorf("err: Resource id[%s] not found", id)

	}
}
