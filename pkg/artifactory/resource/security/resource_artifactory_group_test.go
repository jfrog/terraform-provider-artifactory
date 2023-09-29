package security_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccGroup_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, groupName := testutil.MkNames("test-group-upgrade-", "artifactory_group")
	temp := `
		resource "artifactory_group" "{{ .groupName }}" {
			name = "{{ .groupName }}"
		}
	`
	config := utilsdk.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "7.7.0",
						Source:            "registry.terraform.io/jfrog/artifactory",
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
				ConfigPlanChecks: acctest.ConfigPlanChecks,
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   config,
				PlanOnly:                 true,
				ConfigPlanChecks:         acctest.ConfigPlanChecks,
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
	config := utilsdk.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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

	config := utilsdk.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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

	config := utilsdk.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*can not be set to.*"),
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
			utilsdk.ExecuteTemplate(
				fmt.Sprint(step),
				template,
				map[string]string{"groupName": groupName},
			),
		)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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
			utilsdk.ExecuteTemplate(
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
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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

func testAccCheckGroupDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(utilsdk.ProvderMetadata).Client

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		resp, err := client.R().Head(security.GroupsEndpoint + rs.Primary.ID)
		if err != nil {
			if resp != nil && resp.StatusCode() == http.StatusNotFound {
				return nil
			}
			return err
		}

		return fmt.Errorf("error: Group %s still exists", rs.Primary.ID)
	}
}

func testAccDirectCheckGroupMembership(id string, expectedCount int) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(utilsdk.ProvderMetadata).Client

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
