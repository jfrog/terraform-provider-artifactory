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

package lifecycle_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

type artifactUploadResponse struct {
	Path      string                         `json:"path"`
	Checksums artifactUploadChecksumResponse `json:"checksums"`
}

type artifactUploadChecksumResponse struct {
	SHA256 string `json:"sha256"`
}

func uploadTestFile(t *testing.T, repoKey string) (string, string, error) {
	body, err := os.ReadFile("../../../../samples/multi1-3.7-20220310.233748-1.jar")
	if err != nil {
		return "", "", err
	}
	uri := fmt.Sprintf("/artifactory/%s/org/jfrog/test/multi1/3.7-SNAPSHOT/multi1-3.7-SNAPSHOT.jar", repoKey)

	var result artifactUploadResponse
	_, err = acctest.GetTestResty(t).R().
		SetHeader("Content-Type", "application/java-archive").
		SetBody(body).
		SetResult(&result).
		Put(uri)
	if err != nil {
		return "", "", err
	}

	return result.Path, result.Checksums.SHA256, nil
}

type Build struct {
	Version string        `json:"version"`
	Name    string        `json:"name"`
	Number  string        `json:"number"`
	Started string        `json:"started"`
	Modules []BuildModule `json:"modules"`
}

type BuildModule struct {
	ID        string                `json:"id"`
	Artifacts []BuildModuleArtifact `json:"artifacts"`
}

type BuildModuleArtifact struct {
	Type string `json:"type"`
	Name string `json:"name"`
	SHA1 string `json:"sha1"`
}

func uploadBuild(t *testing.T, name, number, projectKey string) error {
	build := Build{
		Version: "1.0.1",
		Name:    name,
		Number:  number,
		Started: time.Now().Format("2006-01-02T15:04:05.000Z0700"),
		Modules: []BuildModule{
			{
				ID: "org.jfrog.test:multi1:3.7-SNAPSHOT",
				Artifacts: []BuildModuleArtifact{
					{
						Type: "jar",
						Name: "multi1-3.7-SNAPSHOT.jar",
						SHA1: "f142780623aed30ba41d15a8db1ec24da8fd67e8",
					},
				},
			},
		},
	}

	restyClient := acctest.GetTestResty(t)

	req := restyClient.R()

	if projectKey != "" {
		req.SetQueryParam("project", projectKey)
	}

	res, err := req.
		SetBody(build).
		Put("artifactory/api/build")

	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("%s", res.String())
	}

	return nil
}

func deleteBuild(t *testing.T, name, projectKey string) error {
	type Build struct {
		Name      string `json:"buildName"`
		BuildRepo string `json:"buildRepo"`
		DeleteAll bool   `json:"deleteAll"`
	}

	build := Build{
		Name:      name,
		DeleteAll: true,
	}

	restyClient := acctest.GetTestResty(t)

	req := restyClient.R()

	if projectKey != "" {
		build.BuildRepo = fmt.Sprintf("%s-build-info", projectKey)
	}

	res, err := req.
		SetBody(build).
		Post("artifactory/api/build/delete")

	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("%s", res.String())
	}

	return nil
}

func TestAccReleaseBundleV2_full_aql(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

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

	resource "artifactory_release_bundle_v2" "{{ .name }}" {
		name = "{{ .name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "aql"

		source = {
			aql = "items.find({\"repo\": {\"$match\": \"{{ .repo_name }}\"}})"
		}
	}`

	testData := map[string]string{
		"name":         resourceName,
		"keypair_name": keyPairName,
		"repo_name":    repoName,
		"private_key":  os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":   os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
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
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "source.aql", fmt.Sprintf("items.find({\"repo\": {\"$match\": \"%s\"}})", repoName)),
				),
			},
		},
	})
}

func TestAccReleaseBundleV2_full_artifacts(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

	repoName := fmt.Sprintf("test-repo-%d", testutil.RandomInt())
	acctest.CreateRepo(t, repoName, "local", "maven", true, true)

	artifactPath, artifactChecksum, err := uploadTestFile(t, repoName)
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

	resource "artifactory_release_bundle_v2" "{{ .name }}" {
		name = "{{ .name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "artifacts"

		source = {
			artifacts = [{
				path = "{{ .artifact_path }}"
				sha256 = "{{ .artifact_checksum }}"
			}]
		}
	}`

	testData := map[string]string{
		"name":              resourceName,
		"keypair_name":      keyPairName,
		"artifact_path":     fmt.Sprintf("%s%s", repoName, artifactPath),
		"artifact_checksum": artifactChecksum,
		"private_key":       os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":        os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
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
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "source_type", "artifacts"),
					resource.TestCheckResourceAttr(fqrn, "source.artifacts.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "source.artifacts.0.path", testData["artifact_path"]),
					resource.TestCheckResourceAttr(fqrn, "source.artifacts.0.sha256", testData["artifact_checksum"]),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
				),
			},
		},
	})
}

func TestAccReleaseBundleV2_full_builds(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

	repoName := fmt.Sprintf("test-repo-%d", testutil.RandomInt())
	acctest.CreateRepo(t, repoName, "local", "maven", true, true)

	_, _, err := uploadTestFile(t, repoName)
	if err != nil {
		t.Fatalf("failed to upload file: %s", err)
	}

	keyPairName := fmt.Sprintf("test-keypair-%d", testutil.RandomInt())
	buildName := fmt.Sprintf("test-build-%d", testutil.RandomInt())

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

	resource "artifactory_release_bundle_v2" "{{ .name }}" {
		name = "{{ .name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "builds"

		source = {
			builds = [{
				name = "{{ .build_name }}"
				number = "{{ .build_number }}"
			}]
		}
	}`

	testData := map[string]string{
		"name":         resourceName,
		"keypair_name": keyPairName,
		"build_name":   buildName,
		"build_number": "1",
		"private_key":  os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":   os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccReleaseBundleV2_full_builds", template, testData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			if err := uploadBuild(t, buildName, "1", ""); err != nil {
				t.Fatalf("failed to upload build: %s", err)
			}
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			acctest.DeleteRepo(t, repoName)

			if err := deleteBuild(t, buildName, ""); err != nil {
				return err
			}

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "source_type", "builds"),
					resource.TestCheckResourceAttr(fqrn, "source.builds.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "source.builds.0.name", testData["build_name"]),
					resource.TestCheckResourceAttr(fqrn, "source.builds.0.number", "1"),
				),
			},
		},
	})
}

func TestAccReleaseBundleV2_full_release_bundles(t *testing.T) {
	_, _, resourceName1 := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")
	_, fqrn, resourceName2 := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

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

	resource "artifactory_release_bundle_v2" "{{ .name1 }}" {
		name = "{{ .name1 }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "aql"

		source = {
			aql = "items.find({\"repo\": {\"$match\": \"{{ .repo_name }}\"}})"
		}
	}

	resource "artifactory_release_bundle_v2" "{{ .name2 }}" {
		name = "{{ .name2 }}"
		version = "2.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "release_bundles"

		source = {
			release_bundles = [{
				name = artifactory_release_bundle_v2.{{ .name1 }}.name
				version = artifactory_release_bundle_v2.{{ .name1 }}.version
			}]
		}
	}`

	testData := map[string]string{
		"name1":        resourceName1,
		"name2":        resourceName2,
		"keypair_name": keyPairName,
		"repo_name":    repoName,
		"private_key":  os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":   os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	config := util.ExecuteTemplate("TestAccReleaseBundleV2_full_builds", template, testData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			acctest.DeleteRepo(t, repoName)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["name2"]),
					resource.TestCheckResourceAttr(fqrn, "version", "2.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "source_type", "release_bundles"),
					resource.TestCheckResourceAttr(fqrn, "source.release_bundles.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "source.release_bundles.0.name", testData["name1"]),
					resource.TestCheckResourceAttr(fqrn, "source.release_bundles.0.version", "1.0.0"),
				),
			},
		},
	})
}

func TestAccReleaseBundleV2_full_aql_with_project(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

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

	resource "artifactory_release_bundle_v2" "{{ .name }}" {
		name = "{{ .name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		project_key = project.{{ .project_name }}.key
		skip_docker_manifest_resolution = true
		source_type = "aql"

		source = {
			aql = "items.find({\"repo\": {\"$match\": \"{{ .repo_name }}\"}})"
		}
	}`

	testData := map[string]string{
		"name":         resourceName,
		"keypair_name": keyPairName,
		"repo_name":    repoName,
		"project_name": projectName,
		"project_key":  projectKey,
		"private_key":  os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":   os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
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
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "source.aql", fmt.Sprintf("items.find({\"repo\": {\"$match\": \"%s\"}})", repoName)),
				),
			},
		},
	})
}
