package local_test

import (
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalMachineLearningRepository(t *testing.T) {
	acctest.SkipIfNotSupportedVersion(t, "7.102.0")

	_, fqrn, name := testutil.MkNames("machinelearning-local", "artifactory_local_machinelearning_repository")

	temp := `
	resource "artifactory_local_machinelearning_repository" "{{ .name }}" {
		key                      = "{{ .name }}"
		blacked_out              = {{ .blacked_out }}
		xray_index               = {{ .xray_index }}
		property_sets            = ["{{ .property_set }}"]
		archive_browsing_enabled = {{ .archive_browsing_enabled }}
	}`

	params := map[string]interface{}{
		"name":                     name,
		"blacked_out":              testutil.RandBool(),
		"xray_index":               testutil.RandBool(),
		"property_set":             "artifactory",
		"archive_browsing_enabled": testutil.RandBool(),
	}
	config := util.ExecuteTemplate("TestAccLocalMachineLearningRepository", temp, params)

	updatedParams := map[string]interface{}{
		"name":                     name,
		"blacked_out":              testutil.RandBool(),
		"xray_index":               testutil.RandBool(),
		"property_set":             "artifactory",
		"archive_browsing_enabled": testutil.RandBool(),
	}
	updatedConfig := util.ExecuteTemplate("TestAccLocalMachineLearningRepository", temp, updatedParams)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "blacked_out", strconv.FormatBool(params["blacked_out"].(bool))),
					resource.TestCheckResourceAttr(fqrn, "xray_index", strconv.FormatBool(params["xray_index"].(bool))),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.MachineLearningType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
					resource.TestCheckResourceAttr(fqrn, "property_sets.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "property_sets.0", params["property_set"].(string)),
					resource.TestCheckResourceAttr(fqrn, "archive_browsing_enabled", strconv.FormatBool(params["archive_browsing_enabled"].(bool))),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "blacked_out", strconv.FormatBool(updatedParams["blacked_out"].(bool))),
					resource.TestCheckResourceAttr(fqrn, "xray_index", strconv.FormatBool(updatedParams["xray_index"].(bool))),
					resource.TestCheckResourceAttr(fqrn, "property_sets.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "property_sets.0", updatedParams["property_set"].(string)),
					resource.TestCheckResourceAttr(fqrn, "archive_browsing_enabled", strconv.FormatBool(updatedParams["archive_browsing_enabled"].(bool))),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}
