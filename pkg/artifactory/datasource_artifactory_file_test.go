package artifactory

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/stretchr/testify/assert"
)

func downloadPreCheck(t *testing.T, downloadPath string, localFileModTime *time.Time) func() {
	return func() {
		const localFilePath = "../../samples/crash.zip"
		testAccPreCheck(t)
		client := getTestResty(t)
		err := uploadTestFile(client, localFilePath, "example-repo-local/crash.zip", "application/zip")
		if err != nil {
			panic(err)
		}
		//copies the file at the same location where the file should be downloaded by DataSource. It will create the file exist scenario.
		err = copyFile(downloadPath, localFilePath)
		if err != nil {
			panic(err)
		}
		stat, _ := os.Stat(downloadPath)
		*localFileModTime = stat.ModTime()
	}
}

/*
Tests file downloads. Always downloads on force_overwrite = true
*/
func TestDlFile(t *testing.T) {
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
		verified, err := VerifySha256Checksum(download, "7a2489dd209d0acb72f7f11d171b418e65648b9cc96c6c351e00e22551fdd8f1")
		if !verified {
			return fmt.Errorf("%s checksum does not have expected checksum", download)
		}
		return err
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          downloadPreCheck(t, downloadPath, &localFileModTime),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(script, downloadPath),
				Check:  downloadCheck,
			},
		},
	})
}

/*
Negative test case on file download skip
When file is present at output_path, checksum of files at output_path & repository path matches
artifactory_file datasource will skip the download.
*/
func TestFileDownloadSkipCheck(t *testing.T) {
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
		verified, err := VerifySha256Checksum(download, "7a2489dd209d0acb72f7f11d171b418e65648b9cc96c6c351e00e22551fdd8f1")
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
		PreCheck:          downloadPreCheck(t, downloadPath, &localFileModTime),
		ProviderFactories: testAccProviders,
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

func TestFileExists(t *testing.T) {
	tmpFile, err := CreateTempFile("test")

	assert.Nil(t, err)

	defer CloseAndRemove(tmpFile)

	existingPath, _ := filepath.Abs(tmpFile.Name())
	nonExistingPath := existingPath + "-doesnt-exist"

	assert.Equal(t, true, FileExists(existingPath))
	assert.Equal(t, false, FileExists(nonExistingPath))
}

func TestVerifySha256Checksum(t *testing.T) {
	const testString = "test content"
	const expectedSha256 = "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"

	file, err := CreateTempFile(testString)

	assert.Nil(t, err)

	defer CloseAndRemove(file)

	filePath, _ := filepath.Abs(file.Name())

	sha256Verified, err := VerifySha256Checksum(filePath, expectedSha256)

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
