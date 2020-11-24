package artifactory

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const basicWatch = `
resource "artifactory_watch" "watch" {
	name        = "basic-watch"
	description = "basic watch"
	active      = false
}
`

func TestAccWatch_basicWatch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: basicWatch,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_watch.watch", "name", "basic-watch"),
					resource.TestCheckResourceAttr("artifactory_watch.watch", "description", "basic watch"),
				),
			},
		},
	})
}

const basicWatchWithPolicies = `
resource "artifactory_watch" "watch" {
	name        = "basic-watch"
	description = "basic watch"

	all_repositories {}

	policy {
		name = "policy1"
		type = "security"
	}

	policy {
		name = "example-2"
		type = "license"
	}
}
`

func TestAccWatch_basicWatchWithPolicies(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckCreatePolicy(t) },
		CheckDestroy: testAccCheckWatchMissing("artifactory_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      basicWatchWithPolicies,
				ExpectError: regexp.MustCompile("Got invalid watch: policy example-2 doen't exist"),
			},
		},
	})
}

const badPolicy = `
resource "artifactory_watch" "watch" {
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
		CheckDestroy: testAccCheckWatchMissing("artifactory_watch.watch"),
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
resource "artifactory_watch" "watch" {
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
		name = "policy1"
		type = "security"
	}
}
`

func TestAccWatch_allRepositories(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckCreatePolicy(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: allRepositories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_watch.watch", "name", "all-repositories"),
					resource.TestCheckResourceAttr("artifactory_watch.watch", "description", "all repositories"),

					testAccCheckWatchAttributes("artifactory_watch.watch", `repository_paths.\d+.include_patterns.\d+`, []string{"another-path/**", "path1/**"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `repository_paths.\d+.exclude_patterns.\d+`, []string{"another-path/ignore**", "path1/ignore/**"}),

					testAccCheckWatchAttributes("artifactory_watch.watch", `all_repositories.\d+.package_types.\d+`, []string{"Docker", "NuGet"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `all_repositories.\d+.paths.\d+`, []string{"path/*"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `all_repositories.\d+.names.\d+`, []string{"name1", "name2"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `all_repositories.\d+.mime_types.\d+`, []string{"application/zip"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `all_repositories.\d+.property.\d+.key`, []string{"field2", "field4"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `all_repositories.\d+.property.\d+.value`, []string{"value 2", "value 4"}),
				),
			},
		},
	})
}

const namedLocalRepository = `
resource "artifactory_local_repository" "example" {
	key 	     = "local-repo"
	package_type = "generic"
	xray_index   = true
}

resource "artifactory_watch" "watch" {
	name        = "named-local-repo"
	description = "all repositories"

	repository {
		name = artifactory_local_repository.example.key
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
		name = "policy1"
		type = "security"
	}
}
`

func TestAccWatch_namedLocalRepository(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckCreatePolicy(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: namedLocalRepository,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_watch.watch", "name", "named-local-repo"),
					resource.TestCheckResourceAttr("artifactory_watch.watch", "description", "all repositories"),

					testAccCheckWatchAttributes("artifactory_watch.watch", `repository_paths.\d+.include_patterns.\d+`, []string{"path1/**"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `repository_paths.\d+.exclude_patterns.\d+`, []string{"path1/ignore/**"}),

					testAccCheckWatchAttributes("artifactory_watch.watch", `repository.\d+.name`, []string{"local-repo"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `repository.\d+.package_types.\d+`, []string{"Generic"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `repository.\d+.paths.\d+`, []string{"path/*"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `repository.\d+.mime_types.\d+`, []string{"application/zip"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `repository.\d+.property.\d+.key`, []string{"field1", "field2"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `repository.\d+.property.\d+.value`, []string{"value 1", "value 2"}),
				),
			},
		},
	})
}

const allBuilds = `
resource "artifactory_watch" "watch" {
    name = "all_builds"
    description = "all_builds"
    active = true

    all_builds {
        bin_mgr_id = "default"
	}

	policy {
		name = "policy1"
		type = "security"
	}
}
`

func TestAccWatch_allBuilds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckCreatePolicy(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: allBuilds,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_watch.watch", "name", "all_builds"),
					resource.TestCheckResourceAttr("artifactory_watch.watch", "description", "all_builds"),

					testAccCheckWatchAttributes("artifactory_watch.watch", `all_builds.\d+.bin_mgr_id`, []string{"default"}),
				),
			},
		},
	})
}

const buildByName = `
resource "artifactory_watch" "watch" {
    name = "named_builds"
    description = "named_builds"
    active = true

    build {
        name = "build1"
        bin_mgr_id = "default"
	}

	policy {
		name = "policy1"
		type = "security"
	}
}
`

// There isn't a resource to create a build and index it
// Therefore, we expect this test to fail because xray cannot watch a non-existant and unindexed build.
func TestAccWatch_buildByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckCreatePolicy(t) },
		CheckDestroy: testAccCheckWatchMissing("artifactory_watch.watch"),
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
resource "artifactory_watch" "watch" {
    name = "pattern_builds"
    description = "pattern_builds"
    active = true

    all_builds {
        include_patterns = ["hello/**", "apache/**"]
        exclude_patterns = ["apache/bad*", "world/**"]
        bin_mgr_id = "default"        
	}
	
	policy {
		name = "policy1"
		type = "security"
	}
}
`

func TestAccWatch_buildByPattern(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckCreatePolicy(t) },
		CheckDestroy: testAccCheckWatchDestroy("artifactory_watch.watch"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: buildByPattern,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_watch.watch", "name", "pattern_builds"),
					resource.TestCheckResourceAttr("artifactory_watch.watch", "description", "pattern_builds"),

					testAccCheckWatchAttributes("artifactory_watch.watch", `all_builds.\d+.bin_mgr_id`, []string{"default"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `all_builds.\d+.include_patterns.\d+`, []string{"apache/**", "hello/**"}),
					testAccCheckWatchAttributes("artifactory_watch.watch", `all_builds.\d+.exclude_patterns.\d+`, []string{"apache/bad*", "world/**"}),
				),
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
		deletePolicy("policy1")

		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.XrayClient
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		_, resp, err := client.GetWatch(rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound {
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
		deletePolicy("policy1")

		_, ok := s.RootModule().Resources[id]
		if !ok {
			return nil
		}

		return fmt.Errorf("err: Resource id[%s] not found", id)

	}
}

// The following methods are to create and destroy a policy
// We need a policy defined so watches can be set as active
// This code should be replace by a policy resource when that is created.

func testAccPreCheckCreatePolicy(t *testing.T) {
	testAccPreCheck(t)
	createPolicy("policy1")
}

func createPolicy(policyName string) error {
	apis := testAccProvider.Meta().(*ArtClient)
	client := apis.XrayClient

	data := ArtifactoryPolicy{
		Name:        policyName,
		Description: "example policy",
		Type:        "security",
		Rules: []ArtifactoryPolicyRules{{
			Name:     "sec_rule",
			Priority: 1,
			Criteria: map[string]string{
				"min_severity": "medium",
			},
			Actions: ArtifactoryPolicyActions{
				Webhooks: []string{},
				BlockDownload: ArtifactoryPolicyActionsBlockDownload{
					Active:    true,
					Unscanned: false,
				},
				BlockReleaseBundleDistribution: true,
				FailBuild:                      true,
				NotifyDeployer:                 true,
				NotifyWatchRecipients:          true,
			},
		}},
	}

	requestContent, err := json.Marshal(data)
	if err != nil {
		return errors.New("failed marshalling policy " + policyName)
	}
	xrayHTTPDetails := *client.Client().ArtDetails
	httpClientsDetails := xrayHTTPDetails.CreateHttpClientDetails()
	httpClientsDetails.Headers["Content-Type"] = "application/json"

	url := xrayHTTPDetails.GetUrl()

	resp, _, err := client.Client().SendPost(url+"api/v2/policies", requestContent, &httpClientsDetails)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return errors.New("Policy is not Created - " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func deletePolicy(policyName string) error {
	apis := testAccProvider.Meta().(*ArtClient)
	client := apis.XrayClient
	xrayHTTPDetails := *client.Client().ArtDetails
	httpClientsDetails := xrayHTTPDetails.CreateHttpClientDetails()
	httpClientsDetails.Headers["Content-Type"] = "application/json"
	url := xrayHTTPDetails.GetUrl()

	resp, _, err := client.Client().SendDelete(url+"api/v2/policies/"+policyName, nil, &httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to delete policy " + resp.Status)
	}
	return nil
}

type ArtifactoryPolicy struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Type        string                   `json:"type"`
	Rules       []ArtifactoryPolicyRules `json:"rules"`
}

type ArtifactoryPolicyRules struct {
	Name     string                   `json:"name"`
	Priority int                      `json:"priority"`
	Criteria map[string]string        `json:"criteria"`
	Actions  ArtifactoryPolicyActions `json:"actions"`
}

type ArtifactoryPolicyActions struct {
	Webhooks                       []string                              `json:"webhooks"`
	BlockDownload                  ArtifactoryPolicyActionsBlockDownload `json:"block_download"`
	BlockReleaseBundleDistribution bool                                  `json:"block_release_bundle_distribution"`
	FailBuild                      bool                                  `json:"fail_build"`
	NotifyDeployer                 bool                                  `json:"notify_deployer"`
	NotifyWatchRecipients          bool                                  `json:"notify_watch_recipients"`
}

type ArtifactoryPolicyActionsBlockDownload struct {
	Active    bool `json:"active"`
	Unscanned bool `json:"unscanned"`
}
