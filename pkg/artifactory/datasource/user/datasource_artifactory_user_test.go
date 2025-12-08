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

package user_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
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
