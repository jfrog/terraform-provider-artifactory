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
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccGroup_UpgradeFromSDKv2(t *testing.T) {
	providerHost := os.Getenv("TF_ACC_PROVIDER_HOST")
	if providerHost == "registry.opentofu.org" {
		t.Skipf("provider host is registry.opentofu.org. Previous version of Artifactory provider is unknown to OpenTofu.")
	}

	_, fqrn, groupName := testutil.MkNames("test-group-upgrade-", "artifactory_group")
	temp := `
		resource "artifactory_group" "{{ .groupName }}" {
			name = "{{ .groupName }}"
		}
	`
	config := util.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "7.7.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", groupName),
					resource.TestCheckNoResourceAttr(fqrn, "admin_privileges"),
					resource.TestCheckNoResourceAttr(fqrn, "auto_join"),
					resource.TestCheckNoResourceAttr(fqrn, "description"),
					resource.TestCheckNoResourceAttr(fqrn, "detach_all_users"),
					resource.TestCheckNoResourceAttr(fqrn, "external_id"),
					resource.TestCheckResourceAttr(fqrn, "policy_manager", "false"),
					resource.TestCheckResourceAttr(fqrn, "reports_manager", "false"),
					resource.TestCheckNoResourceAttr(fqrn, "realm"),
					resource.TestCheckNoResourceAttr(fqrn, "realm_attributes"),
					resource.TestCheckNoResourceAttr(fqrn, "users_names"),
					resource.TestCheckResourceAttr(fqrn, "watch_manager", "false"),
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

func TestAccGroup_defaults(t *testing.T) {
	_, fqrn, groupName := testutil.MkNames("test-group-basic-", "artifactory_group")
	temp := `
		resource "artifactory_group" "{{ .groupName }}" {
			name  = "{{ .groupName }}"
		}
	`
	config := util.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", groupName),
					resource.TestCheckResourceAttr(fqrn, "auto_join", "false"),
					resource.TestCheckResourceAttr(fqrn, "admin_privileges", "false"),
					resource.TestCheckResourceAttr(fqrn, "realm", "internal"),
					resource.TestCheckNoResourceAttr(fqrn, "detach_all_users"),
					resource.TestCheckResourceAttr(fqrn, "watch_manager", "false"),
					resource.TestCheckResourceAttr(fqrn, "policy_manager", "false"),
					resource.TestCheckResourceAttr(fqrn, "reports_manager", "false"),
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "0"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(groupName, "name"),
				ImportStateVerifyIgnore: []string{"detach_all_users", "users_names"}, // `detach_all_users` attribute is not being sent via API, can't be imported, `users_names` can't be imported due to API specifics
			},
		},
	})
}

func TestAccGroup_full(t *testing.T) {
	_, fqrn, groupName := testutil.MkNames("test-group-full", "artifactory_group")
	temp := `
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group"
			external_id      = "externalID"
			auto_join        = true
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
			detach_all_users = true
			watch_manager    = true
			policy_manager   = true
			reports_manager  = true
			users_names 	 = ["anonymous", "admin"]
		}
	`

	config := util.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", groupName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test group"),
					resource.TestCheckResourceAttr(fqrn, "external_id", "externalID"),
					resource.TestCheckResourceAttr(fqrn, "auto_join", "true"),
					resource.TestCheckResourceAttr(fqrn, "admin_privileges", "false"),
					resource.TestCheckResourceAttr(fqrn, "realm", "test"),
					resource.TestCheckResourceAttr(fqrn, "realm_attributes", "Some attribute"),
					resource.TestCheckResourceAttr(fqrn, "detach_all_users", "true"),
					resource.TestCheckResourceAttr(fqrn, "watch_manager", "true"),
					resource.TestCheckResourceAttr(fqrn, "policy_manager", "true"),
					resource.TestCheckResourceAttr(fqrn, "reports_manager", "true"),
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "users_names.0", "admin"),
					resource.TestCheckResourceAttr(fqrn, "users_names.1", "anonymous"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(groupName, "name"),
				ImportStateVerifyIgnore: []string{"detach_all_users", "users_names"}, // `detach_all_users` attribute is not being sent via API, can't be imported, `users_names` can't be imported due to API specifics
			},
		},
	})
}

func TestAccGroup_bool_conflict(t *testing.T) {
	_, fqrn, groupName := testutil.MkNames("test-group-full", "artifactory_group")
	temp := `
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group"
			external_id      = "externalID"
			auto_join        = true
			admin_privileges = true
			realm            = "test"
			realm_attributes = "Some attribute"
			detach_all_users = true
			watch_manager    = true
			policy_manager   = true
			reports_manager  = true
			users_names 	 = ["anonymous", "admin"]
		}
	`

	config := util.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*can not be set to.*"),
			},
		},
	})
}

func TestAccGroup_name_too_long(t *testing.T) {
	_, fqrn, groupName := testutil.MkNames("test-group-full", "artifactory_group")

	groupName = fmt.Sprintf("%s%s", groupName, strings.Repeat("X", 60))
	temp := `
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group"
			external_id      = "externalID"
			auto_join        = true
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
			detach_all_users = true
			watch_manager    = true
			policy_manager   = true
			reports_manager  = true
			users_names 	 = ["anonymous", "admin"]
		}
	`

	config := util.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*Attribute name string length must be between 1 and 64.*"),
			},
		},
	})
}

func TestAccGroup_unmanaged_members_update(t *testing.T) {
	_, fqrn, groupName := testutil.MkNames("test-group-unmanaged-members", "artifactory_group")

	templates := []string{
		`
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group 0"
			auto_join        = true
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
			users_names = ["anonymous", "admin"]
		}
		`,
		`
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group 1"
			auto_join        = false
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
		}
		`,
		`
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group 2"
			auto_join        = false
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
			detach_all_users = true
		}
		`,
	}
	var configs []string
	for step, template := range templates {
		configs = append(
			configs,
			util.ExecuteTemplate(
				fmt.Sprint(step),
				template,
				map[string]string{"groupName": groupName},
			),
		)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: configs[0],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", groupName),
					resource.TestCheckResourceAttr(fqrn, "auto_join", "true"),
					resource.TestCheckResourceAttr(fqrn, "admin_privileges", "false"),
					resource.TestCheckResourceAttr(fqrn, "realm", "test"),
					resource.TestCheckResourceAttr(fqrn, "realm_attributes", "Some attribute"),
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "2"),
					testAccDirectCheckGroupMembership(fqrn, 2),
				),
			},
			{
				Config: configs[1],
				Check: resource.ComposeTestCheckFunc(
					testAccDirectCheckGroupMembership(fqrn, 2),
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "0"),
				),
			},
			{
				Config: configs[2],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "detach_all_users", "true"),
					testAccDirectCheckGroupMembership(fqrn, 0),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(groupName, "name"),
				ImportStateVerifyIgnore: []string{"detach_all_users"}, // this attribute is not being sent via API, can't be imported
			},
		},
	})
}

func TestAccGroup_full_update(t *testing.T) {
	_, fqrn, groupName := testutil.MkNames("test-group-full", "artifactory_group")
	externalId := "test-external-id"

	templates := []string{
		`
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group"
			external_id      = "{{ .externalId }}"
			auto_join        = true
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
		}
		`,
		`
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group"
			auto_join        = true
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
			users_names = ["anonymous", "admin"]
		}
		`,
		`
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group"
			auto_join        = true
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
			users_names      = ["anonymous"]
		}
		`,
		`
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group"
			auto_join        = true
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
		}
		`,
		`
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group"
			auto_join        = false
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
			users_names      = ["anonymous", "admin"]
		}
		`,
		`
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group"
			auto_join        = false
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
			detach_all_users = true
		}
		`,
		`
		resource "artifactory_group" "{{ .groupName }}" {
			name             = "{{ .groupName }}"
			description 	 = "Test group"
			auto_join        = false
			admin_privileges = false
			realm            = "test"
			realm_attributes = "Some attribute"
			watch_manager    = true
			policy_manager   = true
			reports_manager  = true
		}
		`,
	}

	var configs []string
	for step, template := range templates {
		configs = append(
			configs,
			util.ExecuteTemplate(
				fmt.Sprint(step),
				template,
				map[string]string{
					"groupName":  groupName,
					"externalId": externalId,
				},
			),
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: configs[0],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", groupName),
					resource.TestCheckResourceAttr(fqrn, "external_id", externalId),
					resource.TestCheckResourceAttr(fqrn, "auto_join", "true"),
					resource.TestCheckResourceAttr(fqrn, "admin_privileges", "false"),
					resource.TestCheckResourceAttr(fqrn, "realm", "test"),
					resource.TestCheckResourceAttr(fqrn, "realm_attributes", "Some attribute"),
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "0"),
					testAccDirectCheckGroupMembership(fqrn, 0),
				),
			},
			{
				Config: configs[1],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "users_names.0", "admin"),
					resource.TestCheckResourceAttr(fqrn, "users_names.1", "anonymous"),
					testAccDirectCheckGroupMembership(fqrn, 2),
				),
			},
			{
				Config: configs[2],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "users_names.0", "anonymous"),
					testAccDirectCheckGroupMembership(fqrn, 1),
				),
			},
			{
				Config: configs[3],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "0"),
					testAccDirectCheckGroupMembership(fqrn, 1),
				),
			},
			{
				Config: configs[4],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "users_names.0", "admin"),
					resource.TestCheckResourceAttr(fqrn, "users_names.1", "anonymous"),
					testAccDirectCheckGroupMembership(fqrn, 2),
				),
			},
			{
				Config: configs[5],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "detach_all_users", "true"),
					testAccDirectCheckGroupMembership(fqrn, 0),
					resource.TestCheckResourceAttr(fqrn, "watch_manager", "false"),
					resource.TestCheckResourceAttr(fqrn, "policy_manager", "false"),
					resource.TestCheckResourceAttr(fqrn, "reports_manager", "false"),
				),
			},
			{
				Config: configs[6],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", groupName),
					resource.TestCheckResourceAttr(fqrn, "watch_manager", "true"),
					resource.TestCheckResourceAttr(fqrn, "policy_manager", "true"),
					resource.TestCheckResourceAttr(fqrn, "reports_manager", "true"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(groupName, "name"),
				ImportStateVerifyIgnore: []string{"detach_all_users"}, // this attribute is not being sent via API, can't be imported
			},
		},
	})
}

func TestAccGroup_update_name(t *testing.T) {
	_, fqrn, groupName := testutil.MkNames("test-group-name-", "artifactory_group")

	temp := `
		resource "artifactory_group" "{{ .groupName }}" {
			name  = "{{ .groupName }}"
		}
	`
	config := util.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	updatedTemp := `
		resource "artifactory_group" "{{ .groupName }}" {
			name  = "{{ .groupName }}-updated"
		}
	`

	updatedConfig := util.ExecuteTemplate(groupName, updatedTemp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", groupName),
				),
			},
			{
				Config: updatedConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fqrn, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func testAccCheckGroupDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		resp, err := client.R().Head(security.GroupsEndpoint + rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("error: Group %s still exists", rs.Primary.ID)
	}
}

func testAccDirectCheckGroupMembership(id string, expectedCount int) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		group := security.ArtifactoryGroupResourceAPIModel{}
		_, err := client.R().SetResult(&group).Get(security.GroupsEndpoint + rs.Primary.ID + "?includeUsers=true")
		if err != nil {
			return err
		}

		if len(group.UsersNames) != expectedCount {
			return fmt.Errorf("error: Group %s has wrong number of members. Expected: %d  Actual: %d", rs.Primary.ID, expectedCount, len(group.UsersNames))
		}

		return nil
	}
}
