package datasource_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/datasource"
	"github.com/stretchr/testify/assert"
)

func uploadTestFile(client *resty.Client, localPath, remotePath, contentType string) error {
	body, err := ioutil.ReadFile(localPath)
	if err != nil {
		return err
	}
	uri := "/artifactory/" + remotePath
	_, err = client.R().SetBody(body).SetHeader("Content-Type", contentType).Put(uri)
	return err
}

func downloadPreCheck(t *testing.T, downloadPath string, localFileModTime *time.Time) {
	const localFilePath = "../../../samples/crash.zip"
	client := acctest.GetTestResty(t)
	err := uploadTestFile(client, localFilePath, "example-repo-local/crash.zip", "application/zip")
	if err != nil {
		t.Fatal(err)
	}
	//copies the file at the same location where the file should be downloaded by DataSource. It will create the file exist scenario.
	err = copyFile(downloadPath, localFilePath)
	if err != nil {
		t.Fatal(err)
	}
	stat, _ := os.Stat(downloadPath)
	*localFileModTime = stat.ModTime()
}

func uploadTwoArtifacts(t *testing.T) {
	const localOlderFilePath = "../../../samples/multi1-3.7-20220310.233748-1.jar"
	const localNewerFilePath = "../../../samples/multi1-3.7-20220310.233859-2.jar"
	client := acctest.GetTestResty(t)
	err := uploadTestFile(client, localOlderFilePath, "my-maven-local/org/jfrog/test/multi1/3.7-SNAPSHOT/multi1-3.7-20220310.233748-1.jar", "application/java-archive")
	if err != nil {
		t.Fatal(err)
	}
	err = uploadTestFile(client, localNewerFilePath, "my-maven-local/org/jfrog/test/multi1/3.7-SNAPSHOT/multi1-3.7-20220310.233859-2.jar", "application/java-archive")
	if err != nil {
		t.Fatal(err)
	}
}

/*
Tests file downloads. Always downloads on force_overwrite = true
*/
func TestDownloadFile(t *testing.T) {
	downloadPath := fmt.Sprintf("%s/crash.zip", t.TempDir())
	localFileModTime := time.Time{}

	// every instance of RT has this repo and file out-of-the-box
	const script = `
		data "artifactory_file" "example" {
		  repository      = "example-repo-local"
		  path            = "crash.zip"
		  output_path     = "%s"
		  force_overwrite = true
		}
	`

	var downloadCheck = func(state *terraform.State) error {
		download := state.Modules[0].Resources["data.artifactory_file.example"].Primary.Attributes["output_path"]
		_, err := os.Stat(download)
		if err != nil {
			return err
		}
		verified, err := datasource.VerifySha256Checksum(download, "7a2489dd209d0acb72f7f11d171b418e65648b9cc96c6c351e00e22551fdd8f1")
		if !verified {
			return fmt.Errorf("%s checksum does not have expected checksum", download)
		}
		return err
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			downloadPreCheck(t, downloadPath, &localFileModTime)
		},
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(script, downloadPath),
				Check:  downloadCheck,
			},
		},
	})
}

/*
Tests the artifact download functionality using path_is_aliased.
For this test we create maven repository, upload 2 jars with different timestamps to the repo.
Then download the artifact using SNAPSHOT instead of the actual timestamp in the filename.
Compare sha256 to match with the file with the latest timestamp.
*/
func TestDownloadFileWithPath_is_aliased(t *testing.T) {
	downloadPath := fmt.Sprintf("%s/multi1-3.7-SNAPSHOT.jar", t.TempDir())

	const script = `
	data "artifactory_file" "example" {
		  repository      = "my-maven-local"
		  path            = "org/jfrog/test/multi1/3.7-SNAPSHOT/multi1-3.7-SNAPSHOT.jar"
          output_path     = "%s"
		  force_overwrite = true
          path_is_aliased = true
		}
	`

	var downloadCheck = func(state *terraform.State) error {
		download := state.Modules[0].Resources["data.artifactory_file.example"].Primary.Attributes["output_path"]
		_, err := os.Stat(download)
		if err != nil {
			return err
		}
		verified, err := datasource.VerifySha256Checksum(download, "fb59a2bb4698ed7ea025ea055e5dc1266ea2e669dd689765ebf26bcb7c94a230")
		if !verified {
			return fmt.Errorf("%s checksum does not match", download)
		}
		return err
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateRepo(t, "my-maven-local", "local",
				"maven", true, true)
			uploadTwoArtifacts(t)
		},
		CheckDestroy: func(_ *terraform.State) error {
			acctest.DeleteRepo(t, "my-maven-local")
			return nil
		},
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(script, downloadPath),
				Check:  downloadCheck,
			},
		},
	})
}

/*
Tests the artifact download functionality using path_is_aliased.
For this test we create maven repository, upload 2 jars with different timestamps to the repo.
Then download the artifact using SNAPSHOT instead of the actual timestamp in the filename.
path_is_aliased parameter is set to `false`, so provider will send try to find exact filename with `SNAPSHOT`
in the filename and will fail.
*/
func TestDownloadFileWithPath_is_aliasedNegative(t *testing.T) {
	downloadPath := fmt.Sprintf("%s/multi1-3.7-SNAPSHOT.jar", t.TempDir())

	const script = `
	data "artifactory_file" "example" {
		  repository      = "my-maven-local"
		  path            = "org/jfrog/test/multi1/3.7-SNAPSHOT/multi1-3.7-SNAPSHOT.jar"
          output_path     = "%s"
		  force_overwrite = true
          path_is_aliased = false
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateRepo(t, "my-maven-local", "local",
				"maven", true, true)
			uploadTwoArtifacts(t)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			acctest.DeleteRepo(t, "my-maven-local")
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(script, downloadPath),
				ExpectError: regexp.MustCompile(".*Unable to find item.*"),
			},
		},
	})
}

/*
Negative test case on file download skip
When file is present at output_path, checksum of files at output_path & repository path matches
artifactory_file datasource will skip the download.
*/
func TestDownloadFileSkipCheck(t *testing.T) {
	downloadPath := fmt.Sprintf("%s/crash.zip", t.TempDir())
	localFileModTime := time.Time{}

	// every instance of RT has this repo and file out-of-the-box
	const noOverWriteForcedScript = `
		data "artifactory_file" "example" {
		  repository      = "example-repo-local"
		  path            = "crash.zip"
		  output_path     = "%s"
		  force_overwrite = false
		}
	`
	const forceOverWriteScript = `
		data "artifactory_file" "example" {
		  repository      = "example-repo-local"
		  path            = "crash.zip"
		  output_path     = "%s"
		  force_overwrite = true
		}
	`

	var skipDownloadCheck = func(state *terraform.State) error {
		download := state.Modules[0].Resources["data.artifactory_file.example"].Primary.Attributes["output_path"]
		downloadedFileStat, err := os.Stat(download)
		if err != nil {
			return err
		}
		downloadedFileModTime := downloadedFileStat.ModTime()
		if downloadedFileModTime.After(localFileModTime) { //determines fresh download occurred during the test step
			return fmt.Errorf("fresh download observed. Existing file modification time: %v. Fresh downloaded file modification time: %v", localFileModTime, downloadedFileModTime)
		}
		return nil
	}

	var downloadCheck = func(state *terraform.State) error {
		download := state.Modules[0].Resources["data.artifactory_file.example"].Primary.Attributes["output_path"]
		downloadedFileStat, err := os.Stat(download)
		if err != nil {
			return err
		}
		downloadedFileModTime := downloadedFileStat.ModTime()
		verified, err := datasource.VerifySha256Checksum(download, "7a2489dd209d0acb72f7f11d171b418e65648b9cc96c6c351e00e22551fdd8f1")
		if !verified {
			return fmt.Errorf("%s checksum does not have expected checksum", download)
		}
		//makes sure download occurred.
		if !downloadedFileModTime.After(localFileModTime) { //determines fresh download occurred during the test step
			return fmt.Errorf("fresh download observed. Existing file modification time: %v. Fresh downloaded file modification time: %v", localFileModTime, downloadedFileModTime)
		}
		return err
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			downloadPreCheck(t, downloadPath, &localFileModTime)
		},
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(noOverWriteForcedScript, downloadPath),
				Check:  skipDownloadCheck,
			},
			{
				Config: fmt.Sprintf(forceOverWriteScript, downloadPath),
				Check:  downloadCheck,
			},
		},
	})
}

//Copies file from source path to destination path
func copyFile(destPath string, srcPath string) error {
	destDir := filepath.Dir(destPath)
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return err
	}
	fin, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer fin.Close()

	fout, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer fout.Close()

	_, err = io.Copy(fout, fin)
	if err != nil {
		return err
	}
	return nil
}

func TestDownloadFileExists(t *testing.T) {
	tmpFile, err := CreateTempFile("test")

	assert.Nil(t, err)

	defer CloseAndRemove(tmpFile)

	existingPath, _ := filepath.Abs(tmpFile.Name())
	nonExistingPath := existingPath + "-doesnt-exist"

	assert.Equal(t, true, datasource.FileExists(existingPath))
	assert.Equal(t, false, datasource.FileExists(nonExistingPath))
}

func TestDownloadFileVerifySha256Checksum(t *testing.T) {
	const testString = "test content"
	const expectedSha256 = "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"

	file, err := CreateTempFile(testString)

	assert.Nil(t, err)

	defer CloseAndRemove(file)

	filePath, _ := filepath.Abs(file.Name())

	sha256Verified, err := datasource.VerifySha256Checksum(filePath, expectedSha256)

	assert.Nil(t, err)
	assert.Equal(t, true, sha256Verified)
}

func CreateTempFile(content string) (f *os.File, err error) {
	file, err := ioutil.TempFile(os.TempDir(), "terraform-provider-artifactory-")

	if err != nil {
		return nil, err
	}

	if content != "" {
		_, err := file.WriteString(content)
		if err != nil {
			return nil, err
		}
	}

	return file, err
}

func CloseAndRemove(f *os.File) {
	_ = f.Close()
	_ = os.Remove(f.Name())
}
