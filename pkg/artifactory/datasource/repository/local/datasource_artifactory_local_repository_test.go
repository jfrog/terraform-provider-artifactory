package local_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestAccLocalAllRepoTypes(t *testing.T) {
	for _, repoType := range local.RepoTypesLikeGeneric {
		title := fmt.Sprintf("%s", cases.Title(language.AmericanEnglish).String(strings.ToLower(repoType)))
		t.Run(title, func(t *testing.T) {
			resource.Test(mkTestCase(repoType, t))
		})
	}
}

func mkTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("terraform-local-%s-%d-full", repoType, test.RandomInt())
	resourceName := fmt.Sprintf("data.artifactory_local_%s_repository.%s", repoType, name)
	xrayIndex := test.RandBool()

	params := map[string]interface{}{
		"repoType":  repoType,
		"name":      name,
		"xrayIndex": xrayIndex,
	}
	config := util.ExecuteTemplate("TestAccLocalRepository", `
		resource "artifactory_local_{{ .repoType }}_repository" "{{ .name }}" {
		  key         = "{{ .name }}"
		  description = "Test repo for {{ .name }}"
		  notes       = "Test repo for {{ .name }}"
		  xray_index  = {{ .xrayIndex }}
		}

		data "artifactory_local_{{ .repoType }}_repository" "{{ .name }}" {
		  key = artifactory_local_{{ .repoType }}_repository.{{ .name }}.id
		}
	`, params)

	return t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", repoType),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test repo for %s", name)),
				),
			},
		},
	}
}
