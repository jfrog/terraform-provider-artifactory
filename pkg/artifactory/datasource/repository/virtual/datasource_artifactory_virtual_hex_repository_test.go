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

package virtual_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccDataSourceVirtualHexRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-virtual-test-repo-basic", "data.artifactory_virtual_hex_repository")
	kpId, _, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")
	kpFqrn := "artifactory_keypair." + kpName
	virtualRepositoryBasic := util.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_local_hex_repository" "local-hex" {
			key                     = "{{ .repo_name }}-local"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			depends_on              = [artifactory_keypair.{{ .kp_name }}]
		}

		resource "artifactory_remote_hex_repository" "remote-hex" {
			key                     = "{{ .repo_name }}-remote"
			url                     = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key              = <<EOF
{{ .public_key }}
EOF
			depends_on              = [artifactory_keypair.{{ .kp_name }}]
		}

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key                     = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories            = [
				artifactory_local_hex_repository.local-hex.key,
				artifactory_remote_hex_repository.remote-hex.key
			]
			depends_on              = [
				artifactory_keypair.{{ .kp_name }},
				artifactory_local_hex_repository.local-hex,
				artifactory_remote_hex_repository.remote-hex
			]
		}

		data "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = artifactory_virtual_hex_repository.{{ .repo_name }}.key
		}
	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "hex"),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef(virtual.Rclass, "hex"); return r }()),
				),
			},
		},
	})
}
