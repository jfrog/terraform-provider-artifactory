package artifact_test

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccArtifact_filepath(t *testing.T) {
	_, _, repoName := testutil.MkNames("test-generic-local", "artifactory_local_generic_repository")
	_, fqrn, name := testutil.MkNames("test-artifact-", "artifactory_artifact")

	temp := `
	resource "artifactory_local_generic_repository" "{{ .repoName }}" {
		key = "{{ .repoName }}"
	}

	resource "artifactory_artifact" "{{ .name }}" {
		repository = artifactory_local_generic_repository.{{ .repoName }}.key
		path = "{{ .path }}"
		file_path = "{{ .filePath }}"
	}`

	testData := map[string]string{
		"name":     name,
		"repoName": repoName,
		"path":     "/foo/bar/multi1-3.7-20220310.233748-1.jar",
		"filePath": "../../../../samples/multi1-3.7-20220310.233748-1.jar",
	}
	config := util.ExecuteTemplate(name, temp, testData)

	updatedTestData := map[string]string{
		"name":     name,
		"repoName": repoName,
		"path":     "/foo/bar/multi1-3.7-20220310.233748-1.jar",
		"filePath": "../../../../samples/multi1-3.7-20220310.233859-2.jar",
	}
	updatedConfig := util.ExecuteTemplate(name, temp, updatedTestData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccCheckArtifactDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repository", repoName),
					resource.TestCheckResourceAttr(fqrn, "path", testData["path"]),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_md5"),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_sha1"),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_sha256"),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
					resource.TestCheckResourceAttrSet(fqrn, "created_by"),
					resource.TestCheckResourceAttrSet(fqrn, "download_uri"),
					resource.TestCheckResourceAttrSet(fqrn, "mime_type"),
					resource.TestCheckResourceAttr(fqrn, "size", "1034"),
					resource.TestCheckResourceAttrSet(fqrn, "uri"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repository", repoName),
					resource.TestCheckResourceAttr(fqrn, "path", updatedTestData["path"]),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_md5"),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_sha1"),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_sha256"),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
					resource.TestCheckResourceAttrSet(fqrn, "created_by"),
					resource.TestCheckResourceAttrSet(fqrn, "download_uri"),
					resource.TestCheckResourceAttrSet(fqrn, "mime_type"),
					resource.TestCheckResourceAttr(fqrn, "size", "1034"),
					resource.TestCheckResourceAttrSet(fqrn, "uri"),
				),
			},
		},
	})
}

func TestAccArtifact_content_base64(t *testing.T) {
	_, _, repoName := testutil.MkNames("test-generic-local", "artifactory_local_generic_repository")
	_, fqrn, name := testutil.MkNames("test-artifact-", "artifactory_artifact")

	temp := `
	resource "artifactory_local_generic_repository" "{{ .repoName }}" {
		key = "{{ .repoName }}"
	}

	resource "artifactory_artifact" "{{ .name }}" {
		repository = artifactory_local_generic_repository.{{ .repoName }}.key
		path = "{{ .path }}"
		content_base64 = "{{ .content }}"
	}`

	data, err := os.ReadFile("../../../../samples/multi1-3.7-20220310.233748-1.jar")
	if err != nil {
		t.Fatalf("failed to read file. %v", err)
		return
	}

	testData := map[string]string{
		"name":     name,
		"repoName": repoName,
		"path":     "/foo/bar/multi1-3.7-20220310.233748-1.jar",
		"content":  base64.StdEncoding.EncodeToString(data),
	}
	config := util.ExecuteTemplate(name, temp, testData)

	updatedData, err := os.ReadFile("../../../../samples/multi1-3.7-20220310.233859-2.jar")
	if err != nil {
		t.Fatalf("failed to read file. %v", err)
		return
	}

	updatedTestData := map[string]string{
		"name":     name,
		"repoName": repoName,
		"path":     "/foo/bar/multi1-3.7-20220310.233748-1.jar",
		"content":  base64.StdEncoding.EncodeToString(updatedData),
	}
	updatedConfig := util.ExecuteTemplate(name, temp, updatedTestData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             testAccCheckArtifactDestroy(fqrn),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repository", repoName),
					resource.TestCheckResourceAttr(fqrn, "path", testData["path"]),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_md5"),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_sha1"),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_sha256"),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
					resource.TestCheckResourceAttrSet(fqrn, "created_by"),
					resource.TestCheckResourceAttrSet(fqrn, "download_uri"),
					resource.TestCheckResourceAttrSet(fqrn, "mime_type"),
					resource.TestCheckResourceAttr(fqrn, "size", "1034"),
					resource.TestCheckResourceAttrSet(fqrn, "uri"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repository", repoName),
					resource.TestCheckResourceAttr(fqrn, "path", updatedTestData["path"]),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_md5"),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_sha1"),
					resource.TestCheckResourceAttrSet(fqrn, "checksum_sha256"),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
					resource.TestCheckResourceAttrSet(fqrn, "created_by"),
					resource.TestCheckResourceAttrSet(fqrn, "download_uri"),
					resource.TestCheckResourceAttrSet(fqrn, "mime_type"),
					resource.TestCheckResourceAttr(fqrn, "size", "1034"),
					resource.TestCheckResourceAttrSet(fqrn, "uri"),
				),
			},
		},
	})
}

func TestAccArtifact_invalid_path(t *testing.T) {
	_, _, name := testutil.MkNames("test-artifact-", "artifactory_artifact")

	temp := `
	resource "artifactory_artifact" "{{ .name }}" {
		repository = "test-repo"
		path = "{{ .path }}"
		file_path = "{{ .filePath }}"
	}`
	testData := map[string]string{
		"name":     name,
		"path":     "foo/bar/multi1-3.7-20220310.233748-1.jar",
		"filePath": "../../../../samples/multi1-3.7-20220310.233748-1.jar",
	}

	config := util.ExecuteTemplate(name, temp, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("Path must start with '/'"),
			},
		},
	})
}

func TestAccArtifact_invalid_file_path(t *testing.T) {
	_, _, name := testutil.MkNames("test-artifact-", "artifactory_artifact")

	temp := `
	resource "artifactory_artifact" "{{ .name }}" {
		repository = "test-repo"
		path = "{{ .path }}"
		file_path = "{{ .filePath }}"
	}`
	testData := map[string]string{
		"name":     name,
		"path":     "/foo/bar/multi1-3.7-20220310.233748-1.jar",
		"filePath": "non-exist.jar",
	}

	config := util.ExecuteTemplate(name, temp, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*Invalid file path.*"),
			},
		},
	})
}

func TestAccArtifact_file_path_conflict(t *testing.T) {
	_, _, name := testutil.MkNames("test-artifact-", "artifactory_artifact")

	temp := `
	resource "artifactory_artifact" "{{ .name }}" {
		repository = "test-repo"
		path = "{{ .path }}"
		file_path = "{{ .filePath }}"
		content_base64 = "{{ .content }}"
	}`
	testData := map[string]string{
		"name":     name,
		"path":     "/foo/bar/multi1-3.7-20220310.233748-1.jar",
		"filePath": "../../../../samples/multi1-3.7-20220310.233748-1.jar",
		"content":  "foobar",
	}

	config := util.ExecuteTemplate(name, temp, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*Invalid Attribute Combination.*"),
			},
		},
	})
}

func TestAccArtifact_missing_file_path_and_content_base64(t *testing.T) {
	_, _, name := testutil.MkNames("test-artifact-", "artifactory_artifact")

	temp := `
	resource "artifactory_artifact" "{{ .name }}" {
		repository = "test-repo"
		path = "{{ .path }}"
	}`
	testData := map[string]string{
		"name": name,
		"path": "/foo/bar/multi1-3.7-20220310.233748-1.jar",
	}

	config := util.ExecuteTemplate(name, temp, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*Invalid Attribute Combination.*"),
			},
		},
	})
}

func testAccCheckArtifactDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		repo_path := path.Join(rs.Primary.Attributes["repository"], rs.Primary.Attributes["path"])
		response, err := client.R().
			SetRawPathParam("repo_path", repo_path).
			Get("/artifactory/api/storage/{repo_path}")
		if err != nil {
			return err
		}

		if response.StatusCode() == http.StatusOK {
			return fmt.Errorf("error: artifact %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
