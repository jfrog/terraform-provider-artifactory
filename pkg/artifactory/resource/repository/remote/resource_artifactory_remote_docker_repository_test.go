package remote_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccRemoteDockerRepository_with_external_dependencies_patterns(t *testing.T) {
	_, testCase := mkNewRemoteTestCase(repository.DockerPackageType, t, map[string]interface{}{
		"external_dependencies_enabled":  true,
		"enable_token_authentication":    true,
		"block_pushing_schema1":          true,
		"priority_resolution":            false,
		"external_dependencies_patterns": []interface{}{"**/hub.docker.io/**", "**/bintray.jfrog.io/**"},
		"missed_cache_period_seconds":    1800, // https://github.com/jfrog/terraform-provider-artifactory/issues/225
	})
	resource.Test(t, testCase)
}

func TestAccRemoteDockerRepository_DependenciesTrueEmptyListFails(t *testing.T) {
	const config = `
		resource "artifactory_remote_docker_repository" "remote-docker-repo-basic" {
			key                     		= "remote-docker"
			url                     		= "https://registry.npmjs.org/"
			retrieval_cache_period_seconds 	= 70
			enable_token_authentication    	= true
			block_pushing_schema1          	= true
			priority_resolution            	= false
			external_dependencies_patterns  = ["**/hub.docker.io/**"]
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*Attribute \"external_dependencies_enabled\" must be specified when.*"),
			},
		},
	})
}

func TestAccRemoteDockerRepository_full(t *testing.T) {
	id, fqrn, name := testutil.MkNames("docker-remote-", "artifactory_remote_docker_repository")
	var testData = map[string]string{
		"resource_name":                  name,
		"repo_name":                      fmt.Sprintf("docker-remote-%d", id),
		"url":                            "https://registry-1.docker.io/",
		"assumed_offline_period_secs":    "300",
		"retrieval_cache_period_seconds": "43200",
		"missed_cache_period_seconds":    "7200",
		"excludes_pattern":               "nopat3,nopat2,nopat1",
		"includes_pattern":               "pat3,pat2,pat1",
		"project_id":                     "",
		"notes":                          "internal description",
		"proxy":                          "",
		"username":                       "admin",
		"password":                       "password1",
		"xray_index":                     "false",
		"archive_browsing_enabled":       "false",
		"list_remote_folder_items":       "true",
		"external_dependencies_enabled":  "true",
		"enable_token_authentication":    "true",
	}
	var testDataUpdated = map[string]string{
		"resource_name":                  name,
		"repo_name":                      fmt.Sprintf("docker-remote-%d", id),
		"url":                            "https://registry-1.docker.io/",
		"assumed_offline_period_secs":    "301",
		"retrieval_cache_period_seconds": "43201",
		"missed_cache_period_seconds":    "7201",
		"excludes_pattern":               "nopat3,nopat2,nopat1",
		"includes_pattern":               "pat3,pat2,pat1",
		"project_id":                     "fake-project-id",
		"notes":                          "internal description",
		"proxy":                          "",
		"username":                       "admin1",
		"password":                       "password",
		"xray_index":                     "true",
		"archive_browsing_enabled":       "true",
		"list_remote_folder_items":       "false",
		"external_dependencies_enabled":  "true",
		"enable_token_authentication":    "false",
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, repoTemplate, testData),
				Check:  resource.ComposeTestCheckFunc(verifyRepository(fqrn, testData)),
			},
			{
				Config: util.ExecuteTemplate(fqrn, repoTemplate, testDataUpdated),
				Check:  resource.ComposeTestCheckFunc(verifyRepository(fqrn, testDataUpdated)),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(name, "key"),
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

const repoTemplate = `
resource "artifactory_remote_docker_repository" "{{ .resource_name }}" {
  key                            = "{{ .repo_name }}"
  url                            = "{{ .url }}"
  assumed_offline_period_secs    = {{ .assumed_offline_period_secs }}

  retrieval_cache_period_seconds = {{ .retrieval_cache_period_seconds }}
  missed_cache_period_seconds    = {{ .missed_cache_period_seconds }}
  excludes_pattern               = "{{ .excludes_pattern }}"
  includes_pattern               = "{{ .includes_pattern }}"
  project_id                     = "{{ .project_id }}"
  notes                          = "{{ .notes }}"
  proxy                          = "{{ .proxy }}"
  property_sets                  = [
    "artifactory",
  ]
  username                  = "{{ .username }}"
  password                  = "{{ .password }}"
  xray_index                = {{ .xray_index }}
  archive_browsing_enabled  = {{ .archive_browsing_enabled }}
}
`
