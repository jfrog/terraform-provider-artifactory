package artifactory

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
