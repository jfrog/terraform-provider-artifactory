package security_test

import (
	"log"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
)

func createGroup(groupName string, description string, id string, t *testing.T) {
	group := security.Group{
		Name:            groupName,
		Description:     description,
		ExternalId:      id,
		AutoJoin:        false,
		AdminPrivileges: false,
		Realm:           "Realm name internal",
		RealmAttributes: "Realm attributes for use by internal",
		UsersNames:      []string{"admin", "anonymous"},
		WatchManager:    true,
		PolicyManager:   false,
		ReportsManager:  true,
	}

	restyClient := acctest.GetTestResty(t)
	_, err := restyClient.R().SetBody(group).Put(security.GroupsEndpoint + group.Name)

	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Create group %s done.", group.Name)
}

func deleteGroup(t *testing.T, groupName string) error {
	restyClient := acctest.GetTestResty(t)
	_, err := restyClient.R().Delete(security.GroupsEndpoint + groupName)

	return err
}

func TestAccGroup_basic_datasource(t *testing.T) {
	id, tempFqrn, groupName := test.MkNames("test-group-full", "artifactory_group")
	temp := `
		data "artifactory_group" "{{ .groupName }}" {
			name  = "{{ .groupName }}"
		}
	`
	fqrn := "data." + tempFqrn
	config := util.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	description := "test-group full body"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			createGroup(groupName, description, strconv.Itoa(id), t)
		},
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", groupName),
					resource.TestCheckResourceAttr(fqrn, "description", description),
					resource.TestCheckResourceAttr(fqrn, "external_id", strconv.Itoa(id)),
					resource.TestCheckResourceAttr(fqrn, "auto_join", "false"),
					resource.TestCheckResourceAttr(fqrn, "admin_privileges", "false"),
					resource.TestCheckResourceAttr(fqrn, "realm", "realm name internal"),
					resource.TestCheckResourceAttr(fqrn, "realm_attributes", "Realm attributes for use by internal"),
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "0"), //Include users set to false, so no users should be in this list.
					resource.TestCheckResourceAttr(fqrn, "watch_manager", "true"),
					resource.TestCheckResourceAttr(fqrn, "policy_manager", "false"),
					resource.TestCheckResourceAttr(fqrn, "reports_manager", "true"),
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			return deleteGroup(t, groupName)
		},
	})
}

func TestAccGroup_basic_datasource_includeusers_true(t *testing.T) {
	id, tempFqrn, groupName := test.MkNames("test-group-full", "artifactory_group")
	temp := `
    data "artifactory_group" "{{ .groupName }}" {
      name  = "{{ .groupName }}"
      include_users = true
    }
	`
	fqrn := "data." + tempFqrn
	config := util.ExecuteTemplate(groupName, temp, map[string]string{"groupName": groupName})

	description := "test-group full body. Include users false"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			createGroup(groupName, description, strconv.Itoa(id), t)
		},
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", groupName),
					resource.TestCheckResourceAttr(fqrn, "description", description),
					resource.TestCheckResourceAttr(fqrn, "users_names.#", "2"),
					resource.TestCheckResourceAttr(fqrn, "users_names.0", "admin"),
					resource.TestCheckResourceAttr(fqrn, "users_names.1", "anonymous"),
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			return deleteGroup(t, groupName)
		},
	})
}
