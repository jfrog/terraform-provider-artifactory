package user_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccDataSourceUser_basic(t *testing.T) {
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
			acctest.CreateUserUpdatable(t, name, email)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			return acctest.DeleteUser(t, name)
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "email", email),
					resource.TestCheckResourceAttr(fqrn, "admin", "false"),
					resource.TestCheckResourceAttr(fqrn, "profile_updatable", "true"),
					resource.TestCheckResourceAttr(fqrn, "disable_ui_access", "false"),
					resource.TestCheckResourceAttr(fqrn, "groups.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "id", name),
				),
			},
		},
	})
}
