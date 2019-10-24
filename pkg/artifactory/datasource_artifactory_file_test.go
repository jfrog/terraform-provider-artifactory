package artifactory

import (
	"testing"
    "io/ioutil"
    "os"
	"github.com/stretchr/testify/assert"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"path/filepath"
)

func TestSkipDownload(t *testing.T) {
	const testString = "test content"
	const expectedSha256 = "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"

	file, err := CreateTempFile(testString)

	assert.Nil(t, err)

	defer CloseAndRemove(file)

	existingPath, _ := filepath.Abs(file.Name())
	nonExistingPath := existingPath + "-doesnt-exist"

	sha256 := expectedSha256
	fileInfo := new(v1.FileInfo)
	fileInfo.Checksums = new(v1.Checksums)
	fileInfo.Checksums.Sha256 = &sha256

	skip, err := SkipDownload(fileInfo, existingPath)
	assert.Equal(t, true, skip) // file exists, checksum matches => skip
	assert.Nil(t, err)

	skip, err = SkipDownload(fileInfo, nonExistingPath)
	assert.Equal(t, false, skip) // file doesn't exist => dont skip
	assert.Nil(t, err)

	sha256 = "6666666666666666666666666666666666666666666666666666666666666666"
	fileInfo.Checksums.Sha256 = &sha256

	skip, err = SkipDownload(fileInfo, existingPath)
	assert.Equal(t, false, skip) // file exists, checksum doesnt match => dont skip & err
	assert.NotNil(t, err)
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

	if content != "" {
		file.WriteString(content)
	}

	return file, err
}

func CloseAndRemove(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}