package fileutils

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils/checksum"
)

const (
	SYMLINK_FILE_CONTENT          = ""
	File                 ItemType = "file"
	Dir                  ItemType = "dir"
)

func GetFileSeparator() string {
	return string(os.PathSeparator)
}

// Check if path exists.
// If path points at a symlink and `preserveSymLink == true`,
// function will return `true` regardless of the symlink target
func IsPathExists(path string, preserveSymLink bool) bool {
	_, err := GetFileInfo(path, preserveSymLink)
	return !os.IsNotExist(err)
}

// Check if path points at a file.
// If path points at a symlink and `preserveSymLink == true`,
// function will return `true` regardless of the symlink target
func IsFileExists(path string, preserveSymLink bool) (bool, error) {
	fileInfo, err := GetFileInfo(path, preserveSymLink)
	if err != nil {
		if os.IsNotExist(err) { // If doesn't exist, don't omit an error
			return false, nil
		}
		return false, errorutils.CheckError(err)
	}
	return !fileInfo.IsDir(), nil
}

// Check if path points at a directory.
// If path points at a symlink and `preserveSymLink == true`,
// function will return `false` regardless of the symlink target
func IsDirExists(path string, preserveSymLink bool) (bool, error) {
	fileInfo, err := GetFileInfo(path, preserveSymLink)
	if err != nil {
		if os.IsNotExist(err) { // If doesn't exist, don't omit an error
			return false, nil
		}
		return false, errorutils.CheckError(err)
	}
	return fileInfo.IsDir(), nil
}

// Get the file info of the file in path.
// If path points at a symlink and `preserveSymLink == true`, return the file info of the symlink instead
func GetFileInfo(path string, preserveSymLink bool) (fileInfo os.FileInfo, err error) {
	if preserveSymLink {
		fileInfo, err = os.Lstat(path)
	} else {
		fileInfo, err = os.Stat(path)
	}
	// We should not do CheckError here, because the error is checked by the calling functions.
	return fileInfo, err
}

func IsPathSymlink(path string) bool {
	f, _ := os.Lstat(path)
	return f != nil && IsFileSymlink(f)
}

func IsFileSymlink(file os.FileInfo) bool {
	return file.Mode()&os.ModeSymlink != 0
}

func GetFileAndDirFromPath(path string) (fileName, dir string) {
	index1 := strings.LastIndex(path, "/")
	index2 := strings.LastIndex(path, "\\")
	var index int
	if index1 >= index2 {
		index = index1
	} else {
		index = index2
	}
	if index != -1 {
		fileName = path[index+1:]
		dir = path[:index]
		return
	}
	fileName = path
	dir = ""
	return
}

// Get the local path and filename from original file name and path according to targetPath
func GetLocalPathAndFile(originalFileName, relativePath, targetPath string, flat bool) (localTargetPath, fileName string) {
	targetFileName, targetDirPath := GetFileAndDirFromPath(targetPath)
	localTargetPath = targetDirPath
	if !flat {
		localTargetPath = filepath.Join(targetDirPath, relativePath)
	}

	fileName = originalFileName
	if targetFileName != "" {
		fileName = targetFileName
	}
	return
}

// Return the recursive list of files and directories in the specified path
func ListFilesRecursiveWalkIntoDirSymlink(path string, walkIntoDirSymlink bool) (fileList []string, err error) {
	fileList = []string{}
	err = Walk(path, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	}, walkIntoDirSymlink)
	err = errorutils.CheckError(err)
	return
}

// Return all files with the specified extension in the specified path. Not recursive.
func ListFilesWithExtension(path, ext string) ([]string, error) {
	sep := GetFileSeparator()
	if !strings.HasSuffix(path, sep) {
		path += sep
	}
	fileList := []string{}
	files, _ := ioutil.ReadDir(path)
	path = strings.TrimPrefix(path, "."+sep)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ext) {
			continue
		}
		filePath := path + f.Name()
		exists, err := IsFileExists(filePath, false)
		if err != nil {
			return nil, err
		}
		if exists {
			fileList = append(fileList, filePath)
			continue
		}

		// Checks if the filepath is a symlink.
		if IsPathSymlink(filePath) {
			// Gets the file info of the symlink.
			file, err := GetFileInfo(filePath, false)
			if errorutils.CheckError(err) != nil {
				return nil, err
			}
			// Checks if the symlink is a file.
			if !file.IsDir() {
				fileList = append(fileList, filePath)
			}
		}
	}
	return fileList, nil
}

// Return the list of files and directories in the specified path
func ListFiles(path string, includeDirs bool) ([]string, error) {
	sep := GetFileSeparator()
	if !strings.HasSuffix(path, sep) {
		path += sep
	}
	fileList := []string{}
	files, _ := ioutil.ReadDir(path)
	path = strings.TrimPrefix(path, "."+sep)

	for _, f := range files {
		filePath := path + f.Name()
		exists, err := IsFileExists(filePath, false)
		if err != nil {
			return nil, err
		}
		if exists || IsPathSymlink(filePath) {
			fileList = append(fileList, filePath)
		} else if includeDirs {
			isDir, err := IsDirExists(filePath, false)
			if err != nil {
				return nil, err
			}
			if isDir {
				fileList = append(fileList, filePath)
			}
		}
	}
	return fileList, nil
}

func GetUploadRequestContent(file *os.File) io.Reader {
	if file == nil {
		return bytes.NewBuffer([]byte(SYMLINK_FILE_CONTENT))
	}
	return bufio.NewReader(file)
}

func GetFileSize(file *os.File) (int64, error) {
	size := int64(0)
	if file != nil {
		fileInfo, err := file.Stat()
		if errorutils.CheckError(err) != nil {
			return size, err
		}
		size = fileInfo.Size()
	}
	return size, nil
}

func CreateFilePath(localPath, fileName string) (string, error) {
	if localPath != "" {
		err := os.MkdirAll(localPath, 0777)
		if errorutils.CheckError(err) != nil {
			return "", err
		}
		fileName = filepath.Join(localPath, fileName)
	}
	return fileName, nil
}

func CreateDirIfNotExist(path string) error {
	exist, err := IsDirExists(path, false)
	if exist || err != nil {
		return err
	}
	_, err = CreateFilePath(path, "")
	return err
}

// Reads the content of the file in the source path and appends it to
// the file in the destination path.
func AppendFile(srcPath string, destFile *os.File) error {
	srcFile, err := os.Open(srcPath)
	err = errorutils.CheckError(err)
	if err != nil {
		return err
	}

	defer func() error {
		err := srcFile.Close()
		return errorutils.CheckError(err)
	}()

	reader := bufio.NewReader(srcFile)

	writer := bufio.NewWriter(destFile)
	buf := make([]byte, 1024000)
	for {
		n, err := reader.Read(buf)
		if err != io.EOF {
			err = errorutils.CheckError(err)
			if err != nil {
				return err
			}
		}
		if n == 0 {
			break
		}
		_, err = writer.Write(buf[:n])
		err = errorutils.CheckError(err)
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	return errorutils.CheckError(err)
}

func GetHomeDir() string {
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}
	home = os.Getenv("USERPROFILE")
	if home != "" {
		return home
	}
	user, err := user.Current()
	if err == nil {
		return user.HomeDir
	}
	return ""
}

func IsSshUrl(urlPath string) bool {
	u, err := url.Parse(urlPath)
	if err != nil {
		return false
	}
	return strings.ToLower(u.Scheme) == "ssh"
}

func ReadFile(filePath string) ([]byte, error) {
	content, err := ioutil.ReadFile(filePath)
	err = errorutils.CheckError(err)
	return content, err
}

func GetFileDetails(filePath string) (*FileDetails, error) {
	var err error
	details := new(FileDetails)
	details.Checksum, err = calcChecksumDetails(filePath)

	file, err := os.Open(filePath)
	defer file.Close()
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	fileInfo, err := file.Stat()
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	details.Size = fileInfo.Size()
	return details, nil
}

func calcChecksumDetails(filePath string) (ChecksumDetails, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if errorutils.CheckError(err) != nil {
		return ChecksumDetails{}, err
	}
	checksumInfo, err := checksum.Calc(file)
	if err != nil {
		return ChecksumDetails{}, err
	}
	return ChecksumDetails{Md5: checksumInfo[checksum.MD5], Sha1: checksumInfo[checksum.SHA1], Sha256: checksumInfo[checksum.SHA256]}, nil
}

type FileDetails struct {
	Checksum ChecksumDetails
	Size     int64
}

type ChecksumDetails struct {
	Md5    string
	Sha1   string
	Sha256 string
}

func CopyFile(dst, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	fileName, _ := GetFileAndDirFromPath(src)
	dstPath, err := CreateFilePath(dst, fileName)
	if err != nil {
		return err
	}
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	io.Copy(dstFile, srcFile)
	return nil
}

// Copy directory content from one path to another.
// includeDirs means to copy also the dirs if presented in the src folder.
// excludeNames - Skip files/dirs in the src folder that match names in provided slice. ONLY excludes first layer (only in src folder).
func CopyDir(fromPath, toPath string, includeDirs bool, excludeNames []string) error {
	err := CreateDirIfNotExist(toPath)
	if err != nil {
		return err
	}

	files, err := ListFiles(fromPath, includeDirs)
	if err != nil {
		return err
	}

	for _, v := range files {
		// Skip if excluded
		if IsStringInSlice(filepath.Base(v), excludeNames) {
			continue
		}

		dir, err := IsDirExists(v, false)
		if err != nil {
			return err
		}

		if dir {
			toPath := toPath + GetFileSeparator() + filepath.Base(v)
			err := CopyDir(v, toPath, true, nil)
			if err != nil {
				return err
			}
			continue
		}
		err = CopyFile(toPath, v)
		if err != nil {
			return err
		}
	}
	return err
}

func IsStringInSlice(string string, strings []string) bool {
	for _, v := range strings {
		if v == string {
			return true
		}
	}
	return false
}

// Removing the provided path from the filesystem
func RemovePath(testPath string) error {
	if _, err := os.Stat(testPath); err == nil {
		// Delete the path
		err = os.RemoveAll(testPath)
		if err != nil {
			return errors.New("Cannot remove path: " + testPath + " due to: " + err.Error())
		}
	}
	return nil
}

// Renaming from old path to new path.
func RenamePath(oldPath, newPath string) error {
	err := CopyDir(oldPath, newPath, true, nil)
	if err != nil {
		return errors.New("Error copying directory: " + oldPath + "to" + newPath + err.Error())
	}
	RemovePath(oldPath)
	return nil
}

// Returns the path to the directory in which itemToFind is located.
// Traversing through directories from current work-dir to root.
// itemType determines whether looking for a file or dir.
func FindUpstream(itemToFInd string, itemType ItemType) (wd string, exists bool, err error) {
	// Create a map to store all paths visited, to avoid running in circles.
	visitedPaths := make(map[string]bool)
	// Get the current directory.
	wd, err = os.Getwd()
	if err != nil {
		return
	}
	defer os.Chdir(wd)

	// Get the OS root.
	osRoot := os.Getenv("SYSTEMDRIVE")
	if osRoot != "" {
		// If this is a Windows machine:
		osRoot += "\\"
	} else {
		// Unix:
		osRoot = "/"
	}

	// Check if the current directory includes itemToFind. If not, check the parent directory
	// and so on.
	exists = false
	for {
		// If itemToFind is found in the current directory, return the path.
		if itemType == File {
			exists, err = IsFileExists(filepath.Join(wd, itemToFInd), false)
		} else {
			exists, err = IsDirExists(filepath.Join(wd, itemToFInd), false)
		}
		if err != nil || exists {
			return
		}

		// If this the OS root, we can stop.
		if wd == osRoot {
			break
		}

		// Save this path.
		visitedPaths[wd] = true
		// CD to the parent directory.
		wd = filepath.Dir(wd)
		os.Chdir(wd)

		// If we already visited this directory, it means that there's a loop and we can stop.
		if visitedPaths[wd] {
			return "", false, nil
		}
	}

	return "", false, nil
}

type ItemType string

// Returns true if the two files have the same MD5 checksum.
func FilesIdentical(file1 string, file2 string) (bool, error) {
	srcDetails, err := GetFileDetails(file1)
	if err != nil {
		return false, err
	}
	toCompareDetails, err := GetFileDetails(file2)
	if err != nil {
		return false, err
	}
	return srcDetails.Checksum.Md5 == toCompareDetails.Checksum.Md5, nil
}
