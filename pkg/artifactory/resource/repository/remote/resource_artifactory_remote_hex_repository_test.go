package remote_test

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

func TestAccRemoteHexRepository_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-remote-test-repo", "artifactory_remote_hex_repository")
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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
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

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_full", temp, testData)

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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
			description = "Updated description"
			notes = "Updated notes"
			includes_pattern = "**/*"
			excludes_pattern = "*.tmp"
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccRemoteHexRepository_full", updatedTemp, testData)

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
					resource.TestCheckResourceAttr(fqrn, "url", "https://repo.hex.pm"),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttrSet(fqrn, "public_key"),
					resource.TestCheckResourceAttr(fqrn, "description", "Test repository for Hex"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", ""),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("remote", repository.HexPackageType)
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
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"password", "public_key"},
			},
		},
	})
}

func TestAccRemoteHexRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-remote", "artifactory_remote_hex_repository")
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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_basic", temp, testData)

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
					resource.TestCheckResourceAttr(fqrn, "url", "https://repo.hex.pm"),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttrSet(fqrn, "public_key"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("remote", repository.HexPackageType)
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
				ImportStateVerifyIgnore:              []string{"password", "public_key"},
			},
		},
	})
}

func TestAccRemoteHexRepository_missingKeyPairRef(t *testing.T) {
	_, _, name := testutil.MkNames("hex-remote", "artifactory_remote_hex_repository")

	temp := `
		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			public_key = "test-public-key"
		}
	`

	testData := map[string]interface{}{
		"repo_name": name,
	}

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_missingKeyPairRef", temp, testData)

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

func TestAccRemoteHexRepository_missingPublicKey(t *testing.T) {
	_, _, name := testutil.MkNames("hex-remote", "artifactory_remote_hex_repository")
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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
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

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_missingPublicKey", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s).*(Missing required argument|The argument.*is required|no definition was found).*public_key.*`),
			},
		},
	})
}

func TestAccRemoteHexRepository_missingURL(t *testing.T) {
	_, _, name := testutil.MkNames("hex-remote", "artifactory_remote_hex_repository")
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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_missingURL", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s).*(Missing required argument|The argument.*is required|no definition was found).*url.*`),
			},
		},
	})
}

func TestAccRemoteHexRepository_updateKeyPairRef(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-remote", "artifactory_remote_hex_repository")
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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name1 }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
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

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_updateKeyPairRef", temp, testData)

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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccRemoteHexRepository_updateKeyPairRef", updatedTemp, testData)

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

func TestAccRemoteHexRepository_updatePublicKey(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-remote", "artifactory_remote_hex_repository")
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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_updatePublicKey", temp, testData)

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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApqREcFDt5vV21JVe2QNB
Edvzk6w36aNFhVGWN5toNJRjRJ6m4hIuG4KaXtDWVLjnvct6MYMfqhC79HAGwyF+
IqR6Q6a5bbFSsImgBJwz1oadoVKD6ZNetAuCIK84cjMrEFRkELtEIPNHblCzUkkM
3rS9+DPlnfG8hBvGi6tvQIuZmXGCxF/73hU0/MyGhbmEjIKRtG6b0sJYKelRLTPW
XgK7s5pESgiwf2YC/2MGDXjAJfpfCd0RpLdvd4eRiXtVlE9qO9bND94E7PgQ/xqZ
J1i2xWFndWa6nfFnRxZmCStCOZWYYPlaxr+FZceFbpMwzTNs4g3d4tLNUcbKAIH4
0wIDAQAB
-----END PUBLIC KEY-----
EOF
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccRemoteHexRepository_updatePublicKey", updatedTemp, testData)

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
					resource.TestCheckResourceAttrSet(fqrn, "public_key"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttrSet(fqrn, "public_key"),
				),
			},
		},
	})
}

func TestAccRemoteHexRepository_withProject(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", testutil.RandomInt())
	repoName := fmt.Sprintf("%s-hex-remote", projectKey)
	_, fqrn, name := testutil.MkNames(repoName, "artifactory_remote_hex_repository")
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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
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

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_withProject", temp, testData)

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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
			project_key = "{{ .project_key }}"
			project_environments = ["DEV"]
		}
	`

	updatedConfig := util.ExecuteTemplate("TestAccRemoteHexRepository_withProject", updatedTemp, testData)

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

func TestAccRemoteHexRepository_allFields(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-remote", "artifactory_remote_hex_repository")
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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
			description = "Test description"
			notes = "Test notes"
			includes_pattern = "**/*"
			excludes_pattern = "*.tmp"
			repo_layout_ref = "simple-default"
			hard_fail = true
			offline = false
			store_artifacts_locally = true
			socket_timeout_millis = 20000
			retrieval_cache_period_seconds = 7200
			missed_cache_period_seconds = 1800
			synchronize_properties = true
			block_mismatching_mime_types = true
			bypass_head_requests = false
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_allFields", temp, testData)

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
					resource.TestCheckResourceAttr(fqrn, "url", "https://repo.hex.pm"),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttrSet(fqrn, "public_key"),
					resource.TestCheckResourceAttr(fqrn, "description", "Test description"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Test notes"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "*.tmp"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "simple-default"),
					resource.TestCheckResourceAttr(fqrn, "hard_fail", "true"),
					resource.TestCheckResourceAttr(fqrn, "offline", "false"),
					resource.TestCheckResourceAttr(fqrn, "store_artifacts_locally", "true"),
					resource.TestCheckResourceAttr(fqrn, "socket_timeout_millis", "20000"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "7200"),
					resource.TestCheckResourceAttr(fqrn, "missed_cache_period_seconds", "1800"),
					resource.TestCheckResourceAttr(fqrn, "synchronize_properties", "true"),
					resource.TestCheckResourceAttr(fqrn, "block_mismatching_mime_types", "true"),
					resource.TestCheckResourceAttr(fqrn, "bypass_head_requests", "false"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"password", "public_key"},
			},
		},
	})
}

func TestAccRemoteHexRepository_invalidURL(t *testing.T) {
	_, _, name := testutil.MkNames("hex-remote", "artifactory_remote_hex_repository")
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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			url = "not-a-valid-url"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
		}
	`

	testData := map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_invalidURL", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s).*(invalid|must be a valid URL|Invalid Attribute Value).*url.*`),
			},
		},
	})
}

func TestAccRemoteHexRepository_invalidKey(t *testing.T) {
	_, _, name := testutil.MkNames("hex-remote", "artifactory_remote_hex_repository")
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

		resource "artifactory_remote_hex_repository" "{{ .repo_name }}" {
			key = "{{ .invalid_key }}"
			url = "https://repo.hex.pm"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			public_key = <<EOF
{{ .public_key }}
EOF
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

	config := util.ExecuteTemplate("TestAccRemoteHexRepository_invalidKey", temp, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
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
