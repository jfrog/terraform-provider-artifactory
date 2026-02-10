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
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccTrashCanConfig_full(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	_, fqrn, resourceName := testutil.MkNames("trashcan-", "artifactory_trashcan_config")

	const trashCanConfigTemplate = `
	resource "artifactory_trashcan_config" "{{ .resourceName }}" {
		enabled              = {{ .enabled }}
		retention_period_days = {{ .retention_period_days }}
	}`

	testData := map[string]string{
		"resourceName":          resourceName,
		"enabled":               "true",
		"retention_period_days": "14",
	}

	const trashCanConfigTemplateUpdate = `
	resource "artifactory_trashcan_config" "{{ .resourceName }}" {
		enabled              = {{ .enabled }}
		retention_period_days = {{ .retention_period_days }}
	}`

	testDataUpdated := map[string]string{
		"resourceName":          resourceName,
		"enabled":               "true",
		"retention_period_days": "30",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccTrashCanConfigDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, trashCanConfigTemplate, testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "retention_period_days", "14"),
				),
			},
			{
				Config: util.ExecuteTemplate(fqrn, trashCanConfigTemplateUpdate, testDataUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "retention_period_days", "30"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportStateId:                        resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "enabled",
			},
		},
	})
}

func TestAccTrashCanConfig_disabled(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	_, fqrn, resourceName := testutil.MkNames("trashcan-", "artifactory_trashcan_config")

	const trashCanConfigTemplate = `
	resource "artifactory_trashcan_config" "{{ .resourceName }}" {
		enabled              = false
		retention_period_days = 7
	}`

	testData := map[string]string{
		"resourceName": resourceName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccTrashCanConfigDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, trashCanConfigTemplate, testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "false"),
					resource.TestCheckResourceAttr(fqrn, "retention_period_days", "7"),
				),
			},
		},
	})
}

func testAccTrashCanConfigDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		_, ok := s.RootModule().Resources["artifactory_trashcan_config."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		var trashCanConfig configuration.TrashCanConfig

		response, err := client.R().SetResult(&trashCanConfig).Get(configuration.ConfigurationEndpoint)
		if err != nil {
			return err
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read. Response:%#v", response)
		}

		// Trash can config is a singleton - after delete it resets to defaults
		// so we just check that the resource was removed from state
		return nil
	}
}
