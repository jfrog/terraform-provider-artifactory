package virtual_test

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

func TestAccVirtualHexRepository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-virtual-test-repo", "artifactory_virtual_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")
	localRepoName := testutil.RandSelect("local-repo-1", "local-repo-2", "local-repo-3").(string)

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

		resource "artifactory_local_hex_repository" "{{ .local_repo_name }}" {
			key = "{{ .local_repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = [artifactory_local_hex_repository.{{ .local_repo_name }}.key]
			description = "Test repository for Hex"
			notes = "Internal notes"
			includes_pattern = "**/*"
			excludes_pattern = ""
			artifactory_requests_can_retrieve_remote_artifacts = true
		}
	`

	testData := map[string]interface{}{
		"kp_id":           kpId,
		"kp_name":         kpName,
		"repo_name":       name,
		"local_repo_name": localRepoName,
		"private_key":     os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":      os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccVirtualHexRepository_full", temp, testData)

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

		resource "artifactory_local_hex_repository" "{{ .local_repo_name }}" {
			key = "{{ .local_repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = [artifactory_local_hex_repository.{{ .local_repo_name }}.key]
			description = "Updated description"
			notes = "Updated notes"
			includes_pattern = "**/*"
			excludes_pattern = "*.tmp"
			artifactory_requests_can_retrieve_remote_artifacts = false
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccVirtualHexRepository_full", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
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
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "description", "Test repository for Hex"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", ""),
					resource.TestCheckResourceAttr(fqrn, "artifactory_requests_can_retrieve_remote_artifacts", "true"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("virtual", repository.HexPackageType)
						return r
					}()),
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
					resource.TestCheckResourceAttr(fqrn, "artifactory_requests_can_retrieve_remote_artifacts", "false"),
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

func TestAccVirtualHexRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-virtual", "artifactory_virtual_hex_repository")
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

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = []
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccVirtualHexRepository_basic", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
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
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("virtual", repository.HexPackageType)
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

func TestAccVirtualHexRepository_missingKeyPairRef(t *testing.T) {
	_, _, name := testutil.MkNames("hex-virtual", "artifactory_virtual_hex_repository")

	temp := `
		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
		}
	`

	testData := map[string]interface{}{
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccVirtualHexRepository_missingKeyPairRef", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s).*(Missing required argument|The argument.*is required|no definition was found).*hex_primary_keypair_ref.*`),
			},
		},
	})
}

func TestAccVirtualHexRepository_updateKeyPairRef(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-virtual", "artifactory_virtual_hex_repository")
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

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name1 }}.pair_name
			repositories = []
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

	config := util.ExecuteTemplate("TestAccVirtualHexRepository_updateKeyPairRef", temp, testData)

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

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
			repositories = []
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccVirtualHexRepository_updateKeyPairRef", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
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

func TestAccVirtualHexRepository_withRepositories(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-virtual", "artifactory_virtual_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")
	localRepoName1 := testutil.RandSelect("local-repo-1", "local-repo-2", "local-repo-3").(string)
	localRepoName2 := testutil.RandSelect("local-repo-4", "local-repo-5", "local-repo-6").(string)

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

		resource "artifactory_local_hex_repository" "{{ .local_repo_name1 }}" {
			key = "{{ .local_repo_name1 }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_local_hex_repository" "{{ .local_repo_name2 }}" {
			key = "{{ .local_repo_name2 }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = [
				artifactory_local_hex_repository.{{ .local_repo_name1 }}.key,
				artifactory_local_hex_repository.{{ .local_repo_name2 }}.key
			]
		}
	`

	testData := map[string]interface{}{
		"kp_id":            kpId,
		"kp_name":          kpName,
		"repo_name":        name,
		"local_repo_name1": localRepoName1,
		"local_repo_name2": localRepoName2,
		"private_key":      os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":       os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccVirtualHexRepository_withRepositories", temp, testData)

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

		resource "artifactory_local_hex_repository" "{{ .local_repo_name1 }}" {
			key = "{{ .local_repo_name1 }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_local_hex_repository" "{{ .local_repo_name2 }}" {
			key = "{{ .local_repo_name2 }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = [artifactory_local_hex_repository.{{ .local_repo_name1 }}.key]
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccVirtualHexRepository_withRepositories", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
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
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "2"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
				),
			},
		},
	})
}

func TestAccVirtualHexRepository_withProject(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-hex-virtual", projectKey)
	_, fqrn, name := testutil.MkNames(repoName, "artifactory_virtual_hex_repository")
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

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = []
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

	config := util.ExecuteTemplate("TestAccVirtualHexRepository_withProject", temp, testData)

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

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = []
			project_key = "{{ .project_key }}"
			project_environments = ["DEV"]
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccVirtualHexRepository_withProject", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
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

func TestAccVirtualHexRepository_allFields(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-virtual", "artifactory_virtual_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")
	localRepoName := testutil.RandSelect("local-repo-1", "local-repo-2", "local-repo-3").(string)

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

		resource "artifactory_local_hex_repository" "{{ .local_repo_name }}" {
			key = "{{ .local_repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = [artifactory_local_hex_repository.{{ .local_repo_name }}.key]
			description = "Test description"
			notes = "Test notes"
			includes_pattern = "**/*"
			excludes_pattern = "*.tmp"
			repo_layout_ref = "simple-default"
			artifactory_requests_can_retrieve_remote_artifacts = true
			default_deployment_repo = artifactory_local_hex_repository.{{ .local_repo_name }}.key
		}
	`

	testData := map[string]interface{}{
		"kp_id":           kpId,
		"kp_name":         kpName,
		"repo_name":       name,
		"local_repo_name": localRepoName,
		"private_key":     os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":      os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccVirtualHexRepository_allFields", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
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
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "description", "Test description"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Test notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "*.tmp"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "artifactory_requests_can_retrieve_remote_artifacts", "true"),
					resource.TestCheckResourceAttr(fqrn, "default_deployment_repo", localRepoName),
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

func TestAccVirtualHexRepository_invalidKey(t *testing.T) {
	_, _, name := testutil.MkNames("hex-virtual", "artifactory_virtual_hex_repository")
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

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .invalid_key }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"invalid_key": "invalid key with spaces",
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccVirtualHexRepository_invalidKey", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`.*invalid.*key.*`),
			},
		},
	})
}

func TestAccVirtualHexRepository_updateRepositories(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-virtual", "artifactory_virtual_hex_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")
	localRepoName1 := testutil.RandSelect("local-repo-1", "local-repo-2", "local-repo-3").(string)
	localRepoName2 := testutil.RandSelect("local-repo-4", "local-repo-5", "local-repo-6").(string)

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

		resource "artifactory_local_hex_repository" "{{ .local_repo_name1 }}" {
			key = "{{ .local_repo_name1 }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_local_hex_repository" "{{ .local_repo_name2 }}" {
			key = "{{ .local_repo_name2 }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = [artifactory_local_hex_repository.{{ .local_repo_name1 }}.key]
		}
	`

	testData := map[string]interface{}{
		"kp_id":            kpId,
		"kp_name":          kpName,
		"repo_name":        name,
		"local_repo_name1": localRepoName1,
		"local_repo_name2": localRepoName2,
		"private_key":      os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":       os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccVirtualHexRepository_updateRepositories", temp, testData)

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

		resource "artifactory_local_hex_repository" "{{ .local_repo_name1 }}" {
			key = "{{ .local_repo_name1 }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_local_hex_repository" "{{ .local_repo_name2 }}" {
			key = "{{ .local_repo_name2 }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
		}

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = [
				artifactory_local_hex_repository.{{ .local_repo_name1 }}.key,
				artifactory_local_hex_repository.{{ .local_repo_name2 }}.key
			]
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccVirtualHexRepository_updateRepositories", updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "2"),
				),
			},
		},
	})
}

func TestAccVirtualHexRepository_emptyRepositories(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-virtual", "artifactory_virtual_hex_repository")
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

		resource "artifactory_virtual_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			repositories = []
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccVirtualHexRepository_emptyRepositories", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
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
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
				),
			},
		},
	})
}
