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

package local_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
)

func TestAccDataSourceLocalComposerRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("local-composer-repo", "data.artifactory_local_composer_repository")

	config := fmt.Sprintf(`
		resource "artifactory_local_composer_repository" "%s" {
			key                         = "%s"
			enable_composer_v1_indexing = true
		}

		data "artifactory_local_composer_repository" "%s" {
			key = artifactory_local_composer_repository.%s.key
		}
	`, name, name, name, name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "composer"),
					resource.TestCheckResourceAttr(fqrn, "enable_composer_v1_indexing", "true"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "composer"); return r }()),
				),
			},
		},
	})
}
