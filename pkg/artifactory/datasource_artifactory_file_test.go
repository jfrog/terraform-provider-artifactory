package artifactory

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/stretchr/testify/assert"
)

func TestDlFile(t *testing.T) {
	// every instance of RT has this repo and file out-of-the-box
	script := `
		data "artifactory_file" "example" {
		  repository      = "example-repo-local"
		  path            = "crash.zip"
		  output_path     = "${path.cwd}/crash.zip"
		  force_overwrite = true
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			client := getTestResty(t)
			err := uploadTestFile(client, "../../samples/crash.zip", "example-repo-local/crash.zip", "application/zip")
			if err != nil {
				panic(err)
			}
			if err := copyFile("../../samples/crash.zip", "../../temp/crash.zip"); err != nil {
				panic(err)
			}
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: script,
				Check: func(state *terraform.State) error {
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
				},
			},
		},
	})
}

//Creates new directory tree if not exist
func createNewDir(srcPath string) error {
	if _, err := os.Stat(srcPath); errors.Is(err, os.ErrNotExist) {
		errMkDirAll := os.MkdirAll(srcPath, os.ModePerm)
		if errMkDirAll != nil {
			return errMkDirAll
		}
	}
	return nil
}

//Copies file from source path to destination path
func copyFile(srcPath string, destPath string) error {
	destDir := filepath.Dir(destPath)
	err := createNewDir(destDir)
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
