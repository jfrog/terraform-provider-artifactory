package security_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccGroup_basic(t *testing.T) {
	_, fqrn, groupName := test.MkNames("test-group-full", "artifactory_group")
	temp := `
		resource "artifactory_group" "{{ .groupName }}" {
			name  = "{{ .groupName }}"
		}
	`
	config := util.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", groupName),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(groupName, "name"),
			},
		},
	})
}

func TestAccGroup_full(t *testing.T) {
	_, fqrn, groupName := test.MkNames("test-group-full", "artifactory_group")
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
			users_names = ["anonymous"]
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
			users_names = ["anonymous", "admin"]
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy(fqrn),
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

func TestAccGroup_unmanagedmembers(t *testing.T) {
	_, fqrn, groupName := test.MkNames("test-group-unmanagedmembers", "artifactory_group")

	templates := []string{
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
			auto_join        = false
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy(fqrn),
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

func testAccCheckGroupDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProvderMetadata).Client

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
		client := acctest.Provider.Meta().(util.ProvderMetadata).Client

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		group := security.Group{}
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
