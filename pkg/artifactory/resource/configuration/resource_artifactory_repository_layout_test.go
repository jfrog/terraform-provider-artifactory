package configuration_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLayout_full(t *testing.T) {
	_, fqrn, name := test.MkNames("test", "artifactory_repository_layout")

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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccLayoutDestroy(name),
		Steps: []resource.TestStep{
			{
				Config: layoutConfig,
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
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLayout_importNotFound(t *testing.T) {
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        config,
				ResourceName:  "artifactory_repository_layout.not-exist-test",
				ImportStateId: "not-exist-test",
				ImportState:   true,
				ExpectError:   regexp.MustCompile("Cannot import non-existent remote object"),
			},
		},
	})
}

func testAccLayoutDestroy(name string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProvderMetadata).Client

		_, ok := s.RootModule().Resources["artifactory_repository_layout."+name]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", name)
		}

		layouts := &configuration.Layouts{}

		response, err := client.R().SetResult(&layouts).Get("artifactory/api/system/configuration")
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

func TestAccLayout_validate_distinctive_descriptor_path_pattern(t *testing.T) {
	_, fqrn, name := test.MkNames("test", "artifactory_repository_layout")

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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.VerifyDeleted(fqrn, acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config:      layoutConfig,
				ExpectError: regexp.MustCompile("descriptor_path_pattern must be set when distinctive_descriptor_path_pattern is true"),
			},
		},
	})
}
