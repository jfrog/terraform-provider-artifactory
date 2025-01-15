package lifecycle_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccReleaseBundleV2Promotion_full(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2-promotion", "artifactory_release_bundle_v2_promotion")
	_, _, releaseBundleName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

	repoName := fmt.Sprintf("test-repo-%d", testutil.RandomInt())
	acctest.CreateRepo(t, repoName, "local", "maven", true, true)

	_, _, err := uploadTestFile(t, repoName)
	if err != nil {
		t.Fatalf("failed to upload file: %s", err)
	}

	keyPairName := fmt.Sprintf("test-keypair-%d", testutil.RandomInt())

	const template = `
	resource "artifactory_keypair" "{{ .keypair_name }}" {
		pair_name = "{{ .keypair_name }}"
		pair_type = "RSA"
		alias = "test-alias-{{ .keypair_name }}"
		private_key = <<EOF
{{ .private_key }}
EOF
		public_key = <<EOF
{{ .public_key }}
EOF
	}
	
	resource "artifactory_release_bundle_v2" "{{ .release_bundle_name }}" {
		name = "{{ .release_bundle_name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "aql"

		source = {
			aql = "items.find({\"repo\": {\"$match\": \"{{ .repo_name }}\"}})"
		}
	}

	resource "artifactory_release_bundle_v2_promotion" "{{ .name }}" {
		name = artifactory_release_bundle_v2.{{ .release_bundle_name }}.name
		version = artifactory_release_bundle_v2.{{ .release_bundle_name }}.version
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		environment = "DEV"
		included_repository_keys = ["{{ .repo_name }}"]
	}`

	testData := map[string]string{
		"name":                resourceName,
		"keypair_name":        keyPairName,
		"repo_name":           repoName,
		"release_bundle_name": releaseBundleName,
		"private_key":         os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":          os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccReleaseBundleV2_full", template, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			acctest.DeleteRepo(t, repoName)

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["release_bundle_name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "environment", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "included_repository_keys.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "included_repository_keys.0", testData["repo_name"]),
				),
			},
		},
	})
}

func TestAccReleaseBundleV2Promotion_full_with_project(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2-promotion", "artifactory_release_bundle_v2_promotion")
	_, _, releaseBundleName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

	_, _, projectName := testutil.MkNames("test-project-", "project")
	projectKey := fmt.Sprintf("test%d", testutil.RandomInt())

	repoName := fmt.Sprintf("test-repo-%d", testutil.RandomInt())
	acctest.CreateRepo(t, repoName, "local", "maven", true, true)

	_, _, err := uploadTestFile(t, repoName)
	if err != nil {
		t.Fatalf("failed to upload file: %s", err)
	}

	keyPairName := fmt.Sprintf("test-keypair-%d", testutil.RandomInt())

	const template = `
	resource "project" "{{ .project_name }}" {
		key          = "{{ .project_key }}"
		display_name = "{{ .project_key }}"
		admin_privileges {
			manage_members   = true
			manage_resources = true
			index_resources  = true
		}
	}

	resource "project_repository" "{{ .project_repo_name }}" {
		project_key = project.{{ .project_name }}.key
		key         = "{{ .repo_name }}"
	}

	resource "artifactory_keypair" "{{ .keypair_name }}" {
		pair_name = "{{ .keypair_name }}"
		pair_type = "RSA"
		alias = "test-alias-{{ .keypair_name }}"
		private_key = <<EOF
{{ .private_key }}
EOF
		public_key = <<EOF
{{ .public_key }}
EOF
	}
	
	resource "artifactory_release_bundle_v2" "{{ .release_bundle_name }}" {
		name = "{{ .release_bundle_name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		project_key = project.{{ .project_name }}.key
		skip_docker_manifest_resolution = true
		source_type = "aql"

		source = {
			aql = "items.find({\"repo\": {\"$match\": \"{{ .repo_name }}\"}})"
		}
	}

	resource "artifactory_release_bundle_v2_promotion" "{{ .name }}" {
		name = artifactory_release_bundle_v2.{{ .release_bundle_name }}.name
		version = artifactory_release_bundle_v2.{{ .release_bundle_name }}.version
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		project_key = project.{{ .project_name }}.key
		environment = "DEV"
		included_repository_keys = ["{{ .repo_name }}"]
	}`

	testData := map[string]string{
		"name":                resourceName,
		"keypair_name":        keyPairName,
		"repo_name":           repoName,
		"release_bundle_name": releaseBundleName,
		"project_name":        projectName,
		"project_key":         projectKey,
		"project_repo_name":   fmt.Sprintf("%s-%s", projectKey, repoName),
		"private_key":         os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":          os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccReleaseBundleV2_full", template, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		CheckDestroy: func(*terraform.State) error {
			acctest.DeleteRepo(t, repoName)

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["release_bundle_name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "project_key", testData["project_key"]),
					resource.TestCheckResourceAttr(fqrn, "environment", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "included_repository_keys.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "included_repository_keys.0", testData["repo_name"]),
				),
			},
		},
	})
}
