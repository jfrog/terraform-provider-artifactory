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
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccLocalHexRepository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-local-test-repo", "artifactory_local_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	temp := `
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
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			description = "Test repository for Hex"
			notes = "Internal notes"
			includes_pattern = "**/*"
			excludes_pattern = ""
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccLocalHexRepository_full", temp, testData)

	updatedTemp := `
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
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			description = "Updated description"
			notes = "Updated notes"
			includes_pattern = "**/*"
			excludes_pattern = "*.tmp"
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccLocalHexRepository_full", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test repository for Hex"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", ""),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.HexPackageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "description", "Updated description"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Updated notes"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "*.tmp"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccLocalHexRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-local", "artifactory_local_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	temp := `
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
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccLocalHexRepository_basic", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.HexPackageType)
						return r
					}()),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccLocalHexRepository_missingKeyPairRef(t *testing.T) {
	_, _, name := testutil.MkNames("hex-local", "artifactory_local_hex_repository")

	temp := `
		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}
	`

	testData := map[string]interface{}{
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccLocalHexRepository_missingKeyPairRef", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s).*(Missing required argument|The argument.*is required|no definition was found).*hex_primary_keypair_ref.*`),
			},
		},
	})
}

func TestAccLocalHexRepository_updateKeyPairRef(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-local", "artifactory_local_hex_repository")
	kpId1, kpFqrn1, kpName1 := testutil.MkNames("keypair1", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("keypair2", "artifactory_keypair")

	temp := `
		resource "artifactory_keypair" "{{ .kp_name1 }}" {
			pair_name  = "{{ .kp_name1 }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id1 }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
		}

		resource "artifactory_keypair" "{{ .kp_name2 }}" {
			pair_name  = "{{ .kp_name2 }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id2 }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name1 }}.pair_name
		}
	`

	testData := map[string]interface{}{
		"kp_id1":      kpId1,
		"kp_name1":    kpName1,
		"kp_id2":      kpId2,
		"kp_name2":    kpName2,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccLocalHexRepository_updateKeyPairRef", temp, testData)

	updatedTemp := `
		resource "artifactory_keypair" "{{ .kp_name1 }}" {
			pair_name  = "{{ .kp_name1 }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id1 }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
		}

		resource "artifactory_keypair" "{{ .kp_name2 }}" {
			pair_name  = "{{ .kp_name2 }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id2 }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccLocalHexRepository_updateKeyPairRef", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn1, "pair_name", security.VerifyKeyPair),
			acctest.VerifyDeleted(t, kpFqrn2, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName1),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName2),
				),
			},
		},
	})
}

func TestAccLocalHexRepository_withProject(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-hex-local", projectKey)
	_, fqrn, name := testutil.MkNames(repoName, "artifactory_local_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	temp := `
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
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			project_key = "{{ .project_key }}"
			project_environments = ["DEV", "PROD"]
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"project_key": projectKey,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccLocalHexRepository_withProject", temp, testData)

	updatedTemp := `
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
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			project_key = "{{ .project_key }}"
			project_environments = ["DEV"]
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccLocalHexRepository_withProject", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
			func(state *terraform.State) error {
				acctest.DeleteProject(t, projectKey)
				return nil
			},
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "2"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.0", "DEV"),
				),
			},
		},
	})
}

func TestAccLocalHexRepository_allFields(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-local", "artifactory_local_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	temp := `
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
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			description = "Test description"
			notes = "Test notes"
			includes_pattern = "**/*"
			excludes_pattern = "*.tmp"
			repo_layout_ref = "simple-default"
			blacked_out = false
			xray_index = true
			priority_resolution = false
			archive_browsing_enabled = true
			download_direct = false
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccLocalHexRepository_allFields", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "description", "Test description"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Test notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "*.tmp"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "blacked_out", "false"),
					resource.TestCheckResourceAttr(fqrn, "xray_index", "true"),
					resource.TestCheckResourceAttr(fqrn, "priority_resolution", "false"),
					resource.TestCheckResourceAttr(fqrn, "archive_browsing_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "download_direct", "false"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

func TestAccLocalHexRepository_invalidKey(t *testing.T) {
	_, _, name := testutil.MkNames("hex-local", "artifactory_local_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	// Skip if keypair env vars are not set
	privateKey := os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY")
	publicKey := os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY")
	if privateKey == "" || publicKey == "" {
		t.Skip("JFROG_TEST_RSA_PRIVATE_KEY and JFROG_TEST_RSA_PUBLIC_KEY must be set for this test")
	}

	temp := `
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
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .invalid_key }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"invalid_key": "invalid key with spaces",
		"private_key": privateKey,
		"public_key":  publicKey,
	}

	config := util.ExecuteTemplate("TestAccLocalHexRepository_invalidKey", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`.*(invalid|cannot contain spaces|must not contain).*key.*`),
			},
		},
	})
}
