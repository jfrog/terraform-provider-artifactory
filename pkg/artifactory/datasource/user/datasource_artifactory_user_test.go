package user_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
)

func createUserUpdatable(t *testing.T, name string, email string) {
	userObj := user.User{
		Name:                     name,
		Email:                    email,
		Password:                 "Lizard123!",
		Admin:                    true,
		ProfileUpdatable:         true,
		DisableUIAccess:          false,
		InternalPasswordDisabled: false,
		Groups:                   []string{"readers"},
	}

	restyClient := acctest.GetTestResty(t)
	_, err := restyClient.R().SetBody(userObj).Put(user.UsersEndpointPath + name)

	if err != nil {
		t.Fatal(err)
	}
}

func deleteUser(t *testing.T, name string) error {
	restyClient := acctest.GetTestResty(t)
	_, err := restyClient.R().Delete(user.UsersEndpointPath + name)

	return err
}

func TestAccUser_basic_datasource(t *testing.T) {
	id := test.RandomInt()
	name := fmt.Sprintf("foobar-%d", id)
	email := name + "@test.com"

	temp := `
		data "artifactory_user" "{{ .name }}" {
			name  	= "{{ .name }}"
		}
	`

	config := util.ExecuteTemplate(name, temp, map[string]string{"name": name})
	fqrn := fmt.Sprintf("data.artifactory_user.%s", name)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			createUserUpdatable(t, name, email)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			return deleteUser(t, name)
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "email", email),
					resource.TestCheckResourceAttr(fqrn, "admin", "true"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
				),
			},
		},
	})
}
