package artifactory

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func TestAccGroup_basic(t *testing.T) {
	_, rfqn, groupName := utils.MkNames("test-group-full", "artifactory_group")
	temp := `
		resource "artifactory_group" "{{ .groupName }}" {
			name  = "{{ .groupName }}"
		}
	`
	config := utils.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckGroupDestroy(rfqn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rfqn, "name", groupName),
				),
			},
		},
	})
}

func TestAccGroup_full(t *testing.T) {
	_, rfqn, groupName := utils.MkNames("test-group-full", "artifactory_group")

	templates := []string{
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

	configs := []string{}
	for step, template := range templates {
		configs = append(configs, utils.ExecuteTemplate(fmt.Sprint(step), template, map[string]string{"groupName": groupName}))

	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckGroupDestroy(rfqn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configs[0],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rfqn, "name", groupName),
					resource.TestCheckResourceAttr(rfqn, "auto_join", "true"),
					resource.TestCheckResourceAttr(rfqn, "admin_privileges", "false"),
					resource.TestCheckResourceAttr(rfqn, "realm", "test"),
					resource.TestCheckResourceAttr(rfqn, "realm_attributes", "Some attribute"),
					resource.TestCheckResourceAttr(rfqn, "users_names.#", "0"),
					testAccDirectCheckGroupMembership(rfqn, 0),
				),
			},
			{
				Config: configs[1],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rfqn, "users_names.#", "2"),
					resource.TestCheckResourceAttr(rfqn, "users_names.0", "admin"),
					resource.TestCheckResourceAttr(rfqn, "users_names.1", "anonymous"),
					testAccDirectCheckGroupMembership(rfqn, 2),
				),
			},
			{
				Config: configs[2],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rfqn, "users_names.#", "1"),
					resource.TestCheckResourceAttr(rfqn, "users_names.0", "anonymous"),
					testAccDirectCheckGroupMembership(rfqn, 1),
				),
			},
			{
				Config: configs[3],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rfqn, "users_names.#", "0"),
					testAccDirectCheckGroupMembership(rfqn, 1),
				),
			},
			{
				Config: configs[4],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rfqn, "users_names.#", "2"),
					resource.TestCheckResourceAttr(rfqn, "users_names.0", "admin"),
					resource.TestCheckResourceAttr(rfqn, "users_names.1", "anonymous"),
					testAccDirectCheckGroupMembership(rfqn, 2),
				),
			},
			{
				Config: configs[5],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rfqn, "users_names.#", "0"),
					resource.TestCheckResourceAttr(rfqn, "detach_all_users", "true"),
					testAccDirectCheckGroupMembership(rfqn, 0),
					resource.TestCheckResourceAttr(rfqn, "watch_manager", "false"),
					resource.TestCheckResourceAttr(rfqn, "policy_manager", "false"),
					resource.TestCheckResourceAttr(rfqn, "reports_manager", "false"),
				),
			},
			{
				Config: configs[6],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rfqn, "name", groupName),
					resource.TestCheckResourceAttr(rfqn, "watch_manager", "true"),
					resource.TestCheckResourceAttr(rfqn, "policy_manager", "true"),
					resource.TestCheckResourceAttr(rfqn, "reports_manager", "true"),
				),
			},
		},
	})
}

func TestAccGroup_unmanagedmembers(t *testing.T) {
	_, rfqn, groupName := utils.MkNames("test-group-unmanagedmembers", "artifactory_group")

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
	configs := []string{}
	for step, template := range templates {
		configs = append(configs, utils.ExecuteTemplate(fmt.Sprint(step), template, map[string]string{"groupName": groupName}))

	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckGroupDestroy(rfqn),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configs[0],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rfqn, "name", groupName),
					resource.TestCheckResourceAttr(rfqn, "auto_join", "true"),
					resource.TestCheckResourceAttr(rfqn, "admin_privileges", "false"),
					resource.TestCheckResourceAttr(rfqn, "realm", "test"),
					resource.TestCheckResourceAttr(rfqn, "realm_attributes", "Some attribute"),
					resource.TestCheckResourceAttr(rfqn, "users_names.#", "2"),
					testAccDirectCheckGroupMembership(rfqn, 2),
				),
			},
			{
				Config: configs[1],
				Check: resource.ComposeTestCheckFunc(
					testAccDirectCheckGroupMembership(rfqn, 2),
					resource.TestCheckResourceAttr(rfqn, "users_names.#", "0"),
				),
			},
			{
				Config: configs[2],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rfqn, "users_names.#", "0"),
					resource.TestCheckResourceAttr(rfqn, "detach_all_users", "true"),
					testAccDirectCheckGroupMembership(rfqn, 0),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := testAccProviders["artifactory"]()
		client := provider.Meta().(*resty.Client)

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		resp, err := client.R().Head(groupsEndpoint + rs.Primary.ID)
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
		provider, _ := testAccProviders["artifactory"]()
		client := provider.Meta().(*resty.Client)

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		group := Group{}
		_, err := client.R().SetResult(&group).Get(groupsEndpoint + rs.Primary.ID + "?includeUsers=true")
		if err != nil {
			return err
		}

		if len(group.UsersNames) != expectedCount {
			return fmt.Errorf("error: Group %s has wrong number of members. Expected: %d  Actual: %d", rs.Primary.ID, expectedCount, len(group.UsersNames))
		}

		return nil
	}
}
