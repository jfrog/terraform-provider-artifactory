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

package security_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccGlobalEnvironment_full(t *testing.T) {
	_, fqrn, envName := testutil.MkNames("test-global-env-", "artifactory_global_environment")

	temp := `
		resource "artifactory_global_environment" "{{ .name }}" {
			name = "{{ .envName }}"
		}
	`
	config := util.ExecuteTemplate(envName, temp, map[string]string{"name": envName, "envName": envName})

	newEnvName := fmt.Sprintf("%s-new", envName)
	updatedConfig := util.ExecuteTemplate(newEnvName, temp, map[string]string{"name": envName, "envName": newEnvName})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGlobalEnvironmentDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", envName),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", newEnvName),
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

func TestAccGlobalEnvironment_invalid_name(t *testing.T) {
	testCases := []struct {
		name       string
		errorRegex string
	}{
		{name: "1", errorRegex: ".*must start with a letter and contain letters, digits and `-`"},
		{name: "a#", errorRegex: ".*must start with a letter and contain letters, digits and `-`"},
		{name: "a12345678901234567890123456789012", errorRegex: `.*name string length must be between 1 and 32.*`},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, name := testutil.MkNames("test-", "artifactory_global_environment")

			temp := `
				resource "artifactory_global_environment" "{{ .name }}" {
					name = "{{ .envName }}"
				}
			`
			config := util.ExecuteTemplate(
				testCase.name,
				temp,
				map[string]string{
					"envName": testCase.name,
					"name":    name,
				},
			)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
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

func testAccCheckGlobalEnvironmentDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		environments := security.GlobalEnvironmentsAPIModel{}

		_, err := client.R().
			SetResult(&environments).
			Get("access/api/v1/environments")
		if err != nil {
			return err
		}

		found := false
		for _, env := range environments {
			if env.Name == rs.Primary.ID {
				found = true
			}
		}

		if found {
			return fmt.Errorf("error: global environment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
