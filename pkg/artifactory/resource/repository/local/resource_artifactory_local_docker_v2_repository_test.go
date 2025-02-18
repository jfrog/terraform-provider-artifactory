package local_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccLocalDockerV2Repository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("dockerv2-local", "artifactory_local_docker_v2_repository")
	params := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
	}

	config := util.ExecuteTemplate("TestAccLocalDockerV2Repository", `
		resource "artifactory_local_docker_v2_repository" "{{ .name }}" {
			key = "{{ .name }}"
		}
	`, params)

	updatedConfig := util.ExecuteTemplate("TestAccLocalDockerV2Repository", `
		resource "artifactory_local_docker_v2_repository" "{{ .name }}" {
			key 	     = "{{ .name }}"
			tag_retention = {{ .retention }}
			max_unique_tags = {{ .max_tags }}
			block_pushing_schema1 = {{ .block }}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", "true"),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", "1"),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", "0"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "docker"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", params["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "docker"); return r }()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

func TestAccLocalDockerV2RepositoryWithDefaultMaxUniqueTagsGH370(t *testing.T) {
	_, fqrn, name := testutil.MkNames("dockerv2-local", "artifactory_local_docker_v2_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := util.ExecuteTemplate("TestAccLocalDockerV2Repository", `
		resource "artifactory_local_docker_v2_repository" "{{ .name }}" {
			key = "{{ .name }}"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", "0"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

func TestAccLocalDockerV2Repository_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("dockerv2-local", "artifactory_local_docker_v2_repository")
	params := map[string]interface{}{
		"block":     testutil.RandBool(),
		"retention": testutil.RandSelect(1, 5, 10),
		"max_tags":  testutil.RandSelect(0, 5, 10),
		"name":      name,
	}
	config := util.ExecuteTemplate("TestAccLocalDockerV2Repository", `
		resource "artifactory_local_docker_v2_repository" "{{ .name }}" {
			key 	     = "{{ .name }}"
			tag_retention = {{ .retention }}
			max_unique_tags = {{ .max_tags }}
			block_pushing_schema1 = {{ .block }}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "12.8.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "id", name),
					resource.TestCheckResourceAttr(fqrn, "key", name),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
