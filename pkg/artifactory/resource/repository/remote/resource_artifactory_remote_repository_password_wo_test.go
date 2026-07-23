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

package remote_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

// TestAccRemoteRepository_password_wo verifies that the write-only `password_wo`
// attribute authenticates against the remote registry without persisting the
// secret to Terraform state, and that bumping `password_wo_version` triggers an
// in-place update. Requires Terraform 1.11+ for write-only attribute support.
func TestAccRemoteRepository_password_wo(t *testing.T) {
	_, fqrn, name := testutil.MkNames("remote-generic-wo", "artifactory_remote_generic_repository")

	temp := `
		resource "artifactory_remote_generic_repository" "{{ .repo_name }}" {
			key                 = "{{ .repo_name }}"
			url                 = "https://registry.npmjs.org/"
			username            = "my-user"
			password_wo         = "{{ .password }}"
			password_wo_version = "{{ .version }}"
		}
	`

	config := util.ExecuteTemplate("TestAccRemoteRepository_password_wo", temp, map[string]interface{}{
		"repo_name": name,
		"password":  "initial-secret",
		"version":   "1",
	})

	// Rotate the secret: new value + bumped version to force an update.
	updatedConfig := util.ExecuteTemplate("TestAccRemoteRepository_password_wo", temp, map[string]interface{}{
		"repo_name": name,
		"password":  "rotated-secret",
		"version":   "2",
	})

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "username", "my-user"),
					resource.TestCheckResourceAttr(fqrn, "password_wo_version", "1"),
					// The write-only secret must never be stored in state.
					resource.TestCheckNoResourceAttr(fqrn, "password_wo"),
					// The regular password attribute is not used.
					resource.TestCheckNoResourceAttr(fqrn, "password"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "password_wo_version", "2"),
					resource.TestCheckNoResourceAttr(fqrn, "password_wo"),
				),
			},
		},
	})
}
