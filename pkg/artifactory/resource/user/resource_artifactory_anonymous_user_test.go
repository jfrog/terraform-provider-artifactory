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
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccAnonymousUser_Importable(t *testing.T) {
	const anonymousUserConfig = `
		resource "artifactory_anonymous_user" "anonymous" {
		}
	`

	fqrn := "artifactory_anonymous_user.anonymous"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:           anonymousUserConfig,
				ResourceName:     fqrn,
				ImportState:      true,
				ImportStateId:    "anonymous",
				ImportStateCheck: validator.CheckImportState("anonymous", "id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", "anonymous"),
				),
			},
		},
	})
}

func TestAccAnonymousUser_NotCreatable(t *testing.T) {

	const anonymousUserConfig = `
		resource "artifactory_anonymous_user" "anonymous" {
			name = "anonymous"
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      anonymousUserConfig,
				ExpectError: regexp.MustCompile(".*Anonymous Artifactory user cannot be created.*"),
			},
		},
	})
}
