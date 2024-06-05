package user_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccDataSourceUser_basic(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("foobar-%d", id)
	email := name + "@test.com"

	temp := `
	resource "artifactory_managed_user" "{{ .name }}" {
		name              = "{{ .name }}"
		password          = "Passw0rd!123"
		email             = "{{ .email }}"
		groups            = ["readers"]
		admin             = false
		profile_updatable = true
		disable_ui_access = false
	}

	data "artifactory_user" "{{ .name }}" {
		name = artifactory_managed_user.{{ .name }}.name
	}`

	config := util.ExecuteTemplate(name, temp, map[string]string{
		"name":  name,
		"email": email,
	})

	fqrn := fmt.Sprintf("data.artifactory_user.%s", name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
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
