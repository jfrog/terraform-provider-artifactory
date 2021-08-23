package artifactory

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	artifactory "github.com/jfrog/jfrog-client-go/artifactory/services"
)

const groupBasic = `
resource "artifactory_group" "test-group" {
	name  = "terraform-group"
}`

func TestAccGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckGroupDestroy("artifactory_group.test-group"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: groupBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_group.test-group", "name", "terraform-group"),
				),
			},
		},
	})
}

const groupFull = `
resource "artifactory_group" "test-group" {
	name             = "terraform-group"
    description 	 = "Test group"
	auto_join        = true
	admin_privileges = false
	realm            = "test"
	realm_attributes = "Some attribute"
}`

const groupUserUpdate1 = `
resource "artifactory_group" "test-group" {
	name             = "terraform-group"
    description 	 = "Test group"
	auto_join        = true
	admin_privileges = false
	realm            = "test"
	realm_attributes = "Some attribute"
	users_names = ["anonymous"]
}`

const groupUserUpdate2 = `
resource "artifactory_group" "test-group" {
	name             = "terraform-group"
    description 	 = "Test group"
	auto_join        = true
	admin_privileges = false
	realm            = "test"
	realm_attributes = "Some attribute"
	users_names = ["anonymous", "admin"]
}`

const groupUserUpdate3 = `
resource "artifactory_group" "test-group" {
	name             = "terraform-group"
    description 	 = "Test group"
	auto_join        = true
	admin_privileges = false
	realm            = "test"
	realm_attributes = "Some attribute"
	users_names = ["admin"]
}`

func TestAccGroup_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckGroupDestroy("artifactory_group.test-group"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: groupFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_group.test-group", "name", "terraform-group"),
					resource.TestCheckResourceAttr("artifactory_group.test-group", "auto_join", "true"),
					resource.TestCheckResourceAttr("artifactory_group.test-group", "admin_privileges", "false"),
					resource.TestCheckResourceAttr("artifactory_group.test-group", "realm", "test"),
					resource.TestCheckResourceAttr("artifactory_group.test-group", "realm_attributes", "Some attribute"),
					resource.TestCheckResourceAttr("artifactory_group.test-group", "users_names.#", "0"),
				),
			},
			{
				Config: groupUserUpdate1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_group.test-group", "users_names.#", "1"),
					resource.TestCheckResourceAttr("artifactory_group.test-group", "users_names.1386592683", "anonymous"),
				),
			},
			{
				Config: groupUserUpdate2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_group.test-group", "users_names.#", "2"),
					resource.TestCheckResourceAttr("artifactory_group.test-group", "users_names.3672628397", "admin"),
					resource.TestCheckResourceAttr("artifactory_group.test-group", "users_names.1386592683", "anonymous"),
				),
			},
			{
				Config: groupUserUpdate3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_group.test-group", "users_names.#", "1"),
					resource.TestCheckResourceAttr("artifactory_group.test-group", "users_names.3672628397", "admin"),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apis := testAccProvider.Meta().(*ArtClient)
		client := apis.ArtNew
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		group, err := client.GetGroup(artifactory.GroupParams{GroupDetails: artifactory.Group{Name: rs.Primary.ID}})
		if err != nil {
			return err
		}
		if group == nil {
			return nil
		}

		return fmt.Errorf("error: Group %s still exists", rs.Primary.ID)
	}
}
