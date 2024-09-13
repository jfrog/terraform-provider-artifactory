package artifact_test

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccDataSourceFileList(t *testing.T) {
	_, _, genericRepoName := testutil.MkNames("generic-local", "artifactory_local_generic_repository")
	_, fqrn, name := testutil.MkNames("all-local", "data.artifactory_file_list")

	folderPath := "foo"

	repoConfig := util.ExecuteTemplate(
		"TestAccDataSourceFileList",
		`resource "artifactory_local_generic_repository" "{{ .repoKey }}" {
			key = "{{ .repoKey }}"
		}`,
		map[string]string{
			"repoKey": genericRepoName,
		},
	)

	config := util.ExecuteTemplate(
		"TestAccDataSourceFileList",
		`resource "artifactory_local_generic_repository" "{{ .repoKey }}" {
			key = "{{ .repoKey }}"
		}

		data "artifactory_file_list" "{{ .name }}" {
			repository_key = artifactory_local_generic_repository.{{ .repoKey }}.key
			folder_path    = "{{ .folderPath }}"
		}`,
		map[string]string{
			"repoKey":    genericRepoName,
			"name":       name,
			"folderPath": folderPath,
		},
	)

	filePath := fmt.Sprintf("%s/bar.txt", folderPath)
	artifactoryURL := acctest.GetArtifactoryUrl(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: repoConfig,
			},
			{
				Config: config,
				PreConfig: func() {
					uploadArtifact(t, artifactoryURL, genericRepoName, filePath)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "uri", fmt.Sprintf("%s/artifactory/api/storage/%s/%s", artifactoryURL, genericRepoName, folderPath)),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
					resource.TestCheckResourceAttr(fqrn, "files.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "files.0.uri", "/bar.txt"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.last_modified"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.size"),
					resource.TestCheckResourceAttr(fqrn, "files.0.folder", "false"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.sha1"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.sha2"),
				),
			},
		},
	})
}

func TestAccDataSourceFileList_deep_listing(t *testing.T) {
	_, _, genericRepoName := testutil.MkNames("generic-local", "artifactory_local_generic_repository")
	_, fqrn, name := testutil.MkNames("all-local", "data.artifactory_file_list")

	folderPath := "foo/bar"

	repoConfig := util.ExecuteTemplate(
		"TestAccDataSourceFileList",
		`resource "artifactory_local_generic_repository" "{{ .repoKey }}" {
			key = "{{ .repoKey }}"
		}`,
		map[string]string{
			"repoKey": genericRepoName,
		},
	)

	config := util.ExecuteTemplate(
		"TestAccDataSourceFileList",
		`resource "artifactory_local_generic_repository" "{{ .repoKey }}" {
			key = "{{ .repoKey }}"
		}

		data "artifactory_file_list" "{{ .name }}" {
			repository_key = artifactory_local_generic_repository.{{ .repoKey }}.key
			folder_path    = "/"
			deep_listing   = true
			depth          = 3
		}`,
		map[string]string{
			"repoKey":    genericRepoName,
			"name":       name,
			"folderPath": folderPath,
		},
	)

	filePath := fmt.Sprintf("%s/fizz.txt", folderPath)
	artifactoryURL := acctest.GetArtifactoryUrl(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: repoConfig,
			},
			{
				Config: config,
				PreConfig: func() {
					uploadArtifact(t, artifactoryURL, genericRepoName, filePath)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "uri", fmt.Sprintf("%s/artifactory/api/storage/%s", artifactoryURL, genericRepoName)),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
					resource.TestCheckResourceAttr(fqrn, "files.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "files.0.uri", "/foo/bar/fizz.txt"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.last_modified"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.size"),
					resource.TestCheckResourceAttr(fqrn, "files.0.folder", "false"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.sha1"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.sha2"),
				),
			},
		},
	})
}

func TestAccDataSourceFileList_list_folders(t *testing.T) {
	_, _, genericRepoName := testutil.MkNames("generic-local", "artifactory_local_generic_repository")
	_, fqrn, name := testutil.MkNames("all-local", "data.artifactory_file_list")

	folderPath := "foo/bar"

	repoConfig := util.ExecuteTemplate(
		"TestAccDataSourceFileList",
		`resource "artifactory_local_generic_repository" "{{ .repoKey }}" {
			key = "{{ .repoKey }}"
		}`,
		map[string]string{
			"repoKey": genericRepoName,
		},
	)

	config := util.ExecuteTemplate(
		"TestAccDataSourceFileList",
		`resource "artifactory_local_generic_repository" "{{ .repoKey }}" {
			key = "{{ .repoKey }}"
		}

		data "artifactory_file_list" "{{ .name }}" {
			repository_key = artifactory_local_generic_repository.{{ .repoKey }}.key
			folder_path    = "/"
			deep_listing   = true
			depth          = 3
			list_folders   = true
		}`,
		map[string]string{
			"repoKey":    genericRepoName,
			"name":       name,
			"folderPath": folderPath,
		},
	)

	filePath := fmt.Sprintf("%s/fizz.txt", folderPath)
	artifactoryURL := acctest.GetArtifactoryUrl(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: repoConfig,
			},
			{
				Config: config,
				PreConfig: func() {
					uploadArtifact(t, artifactoryURL, genericRepoName, filePath)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "uri", fmt.Sprintf("%s/artifactory/api/storage/%s", artifactoryURL, genericRepoName)),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
					resource.TestCheckResourceAttr(fqrn, "files.#", "3"),
					resource.TestCheckResourceAttr(fqrn, "files.0.uri", "/foo"),
					resource.TestCheckResourceAttr(fqrn, "files.0.size", "-1"),
					resource.TestCheckResourceAttr(fqrn, "files.0.folder", "true"),
					resource.TestCheckResourceAttr(fqrn, "files.1.uri", "/foo/bar"),
					resource.TestCheckResourceAttr(fqrn, "files.1.size", "-1"),
					resource.TestCheckResourceAttr(fqrn, "files.1.folder", "true"),
					resource.TestCheckResourceAttr(fqrn, "files.2.uri", "/foo/bar/fizz.txt"),
					resource.TestCheckResourceAttrSet(fqrn, "files.2.last_modified"),
					resource.TestCheckResourceAttrSet(fqrn, "files.2.size"),
					resource.TestCheckResourceAttr(fqrn, "files.2.folder", "false"),
					resource.TestCheckResourceAttrSet(fqrn, "files.2.sha1"),
					resource.TestCheckResourceAttrSet(fqrn, "files.2.sha2"),
				),
			},
		},
	})
}

func TestAccDataSourceFileList_include_root_path(t *testing.T) {
	_, _, genericRepoName := testutil.MkNames("generic-local", "artifactory_local_generic_repository")
	_, fqrn, name := testutil.MkNames("all-local", "data.artifactory_file_list")

	folderPath := "foo/bar"

	repoConfig := util.ExecuteTemplate(
		"TestAccDataSourceFileList",
		`resource "artifactory_local_generic_repository" "{{ .repoKey }}" {
			key = "{{ .repoKey }}"
		}`,
		map[string]string{
			"repoKey": genericRepoName,
		},
	)

	config := util.ExecuteTemplate(
		"TestAccDataSourceFileList",
		`resource "artifactory_local_generic_repository" "{{ .repoKey }}" {
			key = "{{ .repoKey }}"
		}

		data "artifactory_file_list" "{{ .name }}" {
			repository_key    = artifactory_local_generic_repository.{{ .repoKey }}.key
			folder_path       = "/"
			deep_listing      = true
			depth             = 3
			list_folders      = true
			include_root_path = true
		}`,
		map[string]string{
			"repoKey":    genericRepoName,
			"name":       name,
			"folderPath": folderPath,
		},
	)

	filePath := fmt.Sprintf("%s/fizz.txt", folderPath)
	artifactoryURL := acctest.GetArtifactoryUrl(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: repoConfig,
			},
			{
				Config: config,
				PreConfig: func() {
					uploadArtifact(t, artifactoryURL, genericRepoName, filePath)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "uri", fmt.Sprintf("%s/artifactory/api/storage/%s", artifactoryURL, genericRepoName)),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
					resource.TestCheckResourceAttr(fqrn, "files.#", "4"),
					resource.TestCheckResourceAttr(fqrn, "files.0.uri", "/"),
					resource.TestCheckResourceAttr(fqrn, "files.0.size", "-1"),
					resource.TestCheckResourceAttr(fqrn, "files.0.folder", "true"),
					resource.TestCheckResourceAttr(fqrn, "files.1.uri", "/foo"),
					resource.TestCheckResourceAttr(fqrn, "files.1.size", "-1"),
					resource.TestCheckResourceAttr(fqrn, "files.1.folder", "true"),
					resource.TestCheckResourceAttr(fqrn, "files.2.uri", "/foo/bar"),
					resource.TestCheckResourceAttr(fqrn, "files.2.size", "-1"),
					resource.TestCheckResourceAttr(fqrn, "files.2.folder", "true"),
					resource.TestCheckResourceAttr(fqrn, "files.3.uri", "/foo/bar/fizz.txt"),
					resource.TestCheckResourceAttrSet(fqrn, "files.3.last_modified"),
					resource.TestCheckResourceAttrSet(fqrn, "files.3.size"),
					resource.TestCheckResourceAttr(fqrn, "files.3.folder", "false"),
					resource.TestCheckResourceAttrSet(fqrn, "files.3.sha1"),
					resource.TestCheckResourceAttrSet(fqrn, "files.3.sha2"),
				),
			},
		},
	})
}

func TestAccDataSourceFileList_metadata_timestamps(t *testing.T) {
	_, _, genericRepoName := testutil.MkNames("generic-local", "artifactory_local_generic_repository")
	_, fqrn, name := testutil.MkNames("all-local", "data.artifactory_file_list")

	folderPath := "foo"

	repoConfig := util.ExecuteTemplate(
		"TestAccDataSourceFileList",
		`resource "artifactory_local_generic_repository" "{{ .repoKey }}" {
			key = "{{ .repoKey }}"
		}`,
		map[string]string{
			"repoKey": genericRepoName,
		},
	)

	config := util.ExecuteTemplate(
		"TestAccDataSourceFileList",
		`resource "artifactory_local_generic_repository" "{{ .repoKey }}" {
			key = "{{ .repoKey }}"
		}

		data "artifactory_file_list" "{{ .name }}" {
			repository_key = artifactory_local_generic_repository.{{ .repoKey }}.key
			folder_path         = "{{ .folderPath }}"
			metadata_timestamps = true
		}`,
		map[string]string{
			"repoKey":    genericRepoName,
			"name":       name,
			"folderPath": folderPath,
		},
	)

	filePath := fmt.Sprintf("%s/bar.txt", folderPath)
	artifactoryURL := acctest.GetArtifactoryUrl(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: repoConfig,
			},
			{
				Config: config,
				PreConfig: func() {
					uploadArtifact(t, artifactoryURL, genericRepoName, filePath)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "uri", fmt.Sprintf("%s/artifactory/api/storage/%s/%s", artifactoryURL, genericRepoName, folderPath)),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
					resource.TestCheckResourceAttr(fqrn, "files.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "files.0.uri", "/bar.txt"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.last_modified"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.size"),
					resource.TestCheckResourceAttr(fqrn, "files.0.folder", "false"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.sha1"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.sha2"),
					resource.TestCheckResourceAttrSet(fqrn, "files.0.metadata_timestamps.properties"),
				),
			},
		},
	})
}

type Artifact struct {
	Uri         string            `json:"uri"`
	DownloadUri string            `json:"downloadUri"`
	Repo        string            `json:"repo"`
	Path        string            `json:"path"`
	Created     time.Time         `json:"created"`
	CreatedBy   string            `json:"createdBy"`
	Size        string            `json:"size"`
	MimeType    string            `json:"mimeType"`
	Checksums   map[string]string `json:"checksums"`
}

func uploadArtifact(t *testing.T, artifactoryUrl, repoKey, filePath string) {
	uri, err := url.JoinPath(artifactoryUrl, repoKey, filePath)
	if err != nil {
		t.Fatal(err)
	}

	artifact := Artifact{
		Uri:         uri,
		DownloadUri: uri,
		Repo:        repoKey,
		Path:        filePath,
	}

	restyClient := acctest.GetTestResty(t)
	_, err = restyClient.R().
		SetPathParam("repoKey", repoKey).
		SetBody(&artifact).
		Put("/artifactory/{repoKey}/" + filePath)

	if err != nil {
		t.Fatal(err)
	}

	_, err = restyClient.R().
		SetPathParam("repoKey", repoKey).
		SetQueryParams(map[string]string{
			"properties": "test=1",
		}).
		Put("/artifactory/api/storage/{repoKey}/" + filePath)

	if err != nil {
		t.Fatal(err)
	}
}
