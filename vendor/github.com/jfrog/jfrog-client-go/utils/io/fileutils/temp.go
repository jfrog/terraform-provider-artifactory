package fileutils

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	tempPrefix     = "jfrog.cli.temp."
)

// Max temp file age in hours
var maxFileAge = 24.0

// Path to the root temp dir
var tempDirBase string

func init() {
	tempDirBase = os.TempDir()
}

// Creates the temp dir at tempDirBase.
// Set tempDirPath to the created directory path.
func CreateTempDir() (string, error) {
	if tempDirBase == "" {
		return "", errorutils.CheckError(errors.New("Temp dir cannot be created in an empty base dir."))
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	path, err := ioutil.TempDir(tempDirBase, tempPrefix+"-"+timestamp+"-")
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	return path, nil
}

// Change the containing directory of temp dir.
func SetTempDirBase(dirPath string) {
	tempDirBase = dirPath
}

func RemoveTempDir(dirPath string) error {
	exists, err := IsDirExists(dirPath, false)
	if err != nil {
		return err
	}
	if exists {
		return os.RemoveAll(dirPath)
	}
	return nil
}

// Create a new temp file named "tempPrefix+timeStamp".
func CreateTempFile() (*os.File, error) {
	if tempDirBase == "" {
		return nil, errorutils.CheckError(errors.New("Temp File cannot be created in an empty base dir."))
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	fd, err := ioutil.TempFile(tempDirBase, tempPrefix+"-"+timestamp+"-")
	return fd, err
}

// Old runs/tests may leave junk at temp dir.
// Each temp file/Dir is named with prefix+timestamp, search for all temp files/dirs that match the common prefix and validate their timestamp.
func CleanOldDirs() error {
	// Get all files at temp dir
	files, err := ioutil.ReadDir(tempDirBase)
	if err != nil {
		log.Error(err)
		return errorutils.CheckError(err)
	}
	now := time.Now()
	// Search for files/dirs that match the template.
	for _, file := range files {
		if strings.HasPrefix(file.Name(), tempPrefix) {
			timeStamp, err := extractTimestamp(file.Name())
			if err != nil {
				return err
			}
			// Delete old file/dirs.
			if now.Sub(timeStamp).Hours() > maxFileAge {
				if err := os.Remove(path.Join(tempDirBase, file.Name())); err != nil {
					return errorutils.CheckError(err)
				}
			}
		}
	}
	return nil
}

func extractTimestamp(item string) (time.Time, error) {
	// Get timestamp from file/dir.
	endTimestampIdx := strings.LastIndex(item, "-")
	beginingTimestampIdx := strings.LastIndex(item[:endTimestampIdx], "-")
	timestampStr := item[beginingTimestampIdx+1:endTimestampIdx]
	// Convert to int.
	timeStampint, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}, errorutils.CheckError(err)
	}
	// Convert to time type.
	return time.Unix(timeStampint, 0), nil
}
