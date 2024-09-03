package configuration_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccRepositoryLayout_UpgradeFromSDKv2(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	_, fqrn, name := testutil.MkNames("test", "artifactory_repository_layout")

	config := util.ExecuteTemplate("layout", `
		resource "artifactory_repository_layout" "{{ .name }}" {
			name                                = "{{ .name }}"
			artifact_path_pattern               = "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).[ext]"
			distinctive_descriptor_path_pattern = false
			folder_integration_revision_regexp  = "SNAPSHOT"
			file_integration_revision_regexp    = "SNAPSHOT|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))"
		}
	`, map[string]interface{}{
		"name": name,
	})

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "10.6.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "artifact_path_pattern", "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).[ext]"),
					resource.TestCheckResourceAttr(fqrn, "distinctive_descriptor_path_pattern", "false"),
					resource.TestCheckResourceAttr(fqrn, "descriptor_path_pattern", ""),
					resource.TestCheckResourceAttr(fqrn, "folder_integration_revision_regexp", "SNAPSHOT"),
					resource.TestCheckResourceAttr(fqrn, "file_integration_revision_regexp", "SNAPSHOT|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks:         testutil.ConfigPlanChecks(""),
			},
		},
	})
}

func TestAccRepositoryLayout_full(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	_, fqrn, name := testutil.MkNames("test", "artifactory_repository_layout")

	layoutConfig := util.ExecuteTemplate("layout", `
		resource "artifactory_repository_layout" "{{ .name }}" {
			name                                = "{{ .name }}"
			artifact_path_pattern               = "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).[ext]"
			distinctive_descriptor_path_pattern = false
			folder_integration_revision_regexp  = "SNAPSHOT"
			file_integration_revision_regexp    = "SNAPSHOT|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))"
		}
	`, map[string]interface{}{
		"name": name,
	})

	layoutUpdatedConfig := util.ExecuteTemplate("layout", `
		resource "artifactory_repository_layout" "{{ .name }}" {
			name                                = "{{ .name }}"
			artifact_path_pattern               = "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).[ext]"
			distinctive_descriptor_path_pattern = true
			descriptor_path_pattern             = "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).pom"
			folder_integration_revision_regexp  = "Foo"
			file_integration_revision_regexp    = "Foo|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))"
		}
	`, map[string]interface{}{
		"name": name,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccRepositoryLayoutDestroy(name),
		Steps: []resource.TestStep{
			{
				Config: layoutConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "artifact_path_pattern", "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).[ext]"),
					resource.TestCheckResourceAttr(fqrn, "distinctive_descriptor_path_pattern", "false"),
					resource.TestCheckNoResourceAttr(fqrn, "descriptor_path_pattern"),
					resource.TestCheckResourceAttr(fqrn, "folder_integration_revision_regexp", "SNAPSHOT"),
					resource.TestCheckResourceAttr(fqrn, "file_integration_revision_regexp", "SNAPSHOT|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))"),
				),
			},
			{
				Config: layoutUpdatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "artifact_path_pattern", "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).[ext]"),
					resource.TestCheckResourceAttr(fqrn, "distinctive_descriptor_path_pattern", "true"),
					resource.TestCheckResourceAttr(fqrn, "descriptor_path_pattern", "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).pom"),
					resource.TestCheckResourceAttr(fqrn, "folder_integration_revision_regexp", "Foo"),
					resource.TestCheckResourceAttr(fqrn, "file_integration_revision_regexp", "Foo|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}

func TestAccRepositoryLayout_importNotFound(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	config := `
		resource "artifactory_repository_layout" "not-exist-test" {
			name                                = "not-exist-test"
			artifact_path_pattern               = "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).[ext]"
			distinctive_descriptor_path_pattern = true
			descriptor_path_pattern             = "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).pom"
			folder_integration_revision_regexp  = "Foo"
			file_integration_revision_regexp    = "Foo|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))"
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:                               config,
				ResourceName:                         "artifactory_repository_layout.not-exist-test",
				ImportStateId:                        "not-exist-test",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "name",
				ExpectError:                          regexp.MustCompile("Cannot import non-existent remote object"),
			},
		},
	})
}

func testAccRepositoryLayoutDestroy(name string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		_, ok := s.RootModule().Resources["artifactory_repository_layout."+name]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", name)
		}

		var layouts configuration.RepositoryLayoutsAPIModel
		response, err := client.R().SetResult(&layouts).Get(configuration.ConfigurationEndpoint)
		if err != nil {
			return err
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read. Response: %#v", response)
		}

		for _, layout := range layouts.Layouts {
			if layout.Name == name {
				return fmt.Errorf("error: Layout with name: %s still exists.", name)
			}
		}
		return nil
	}
}

func TestAccRepositoryLayout_validate_distinctive_descriptor_path_pattern(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test", "artifactory_repository_layout")

	layoutConfig := util.ExecuteTemplate("layout", `
		resource "artifactory_repository_layout" "{{ .name }}" {
			name                                = "{{ .name }}"
			artifact_path_pattern               = "[orgPath]/[module]/[baseRev](-[folderItegRev])/[module]-[baseRev](-[fileItegRev])(-[classifier]).[ext]"
			distinctive_descriptor_path_pattern = true
			folder_integration_revision_regexp  = "SNAPSHOT"
			file_integration_revision_regexp    = "SNAPSHOT|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))"
		}
	`, map[string]interface{}{
		"name": name,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config:      layoutConfig,
				ExpectError: regexp.MustCompile(".*descriptor_path_pattern must be set when distinctive_descriptor_path_pattern\n.*is true.*"),
			},
		},
	})
}
