package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	Development = "development"
	Agent       = "jfrog-client-go"
	Version     = "0.13.1"
)

// In order to limit the number of items loaded from a reader into the memory, we use a buffers with this size limit.
var MaxBufferSize = 50000

var userAgent = getDefaultUserAgent()

func getVersion() string {
	return Version
}

func GetUserAgent() string {
	return userAgent
}

func SetUserAgent(newUserAgent string) {
	userAgent = newUserAgent
}

func getDefaultUserAgent() string {
	return fmt.Sprintf("%s/%s", Agent, getVersion())
}

// Get the local root path, from which to start collecting artifacts to be used for:
// 1. Uploaded to Artifactory,
// 2. Adding to the local build-info, to be later published to Artifactory.
func GetRootPath(path string, useRegExp bool, parentheses ParenthesesSlice) string {
	// The first step is to split the local path pattern into sections, by the file separator.
	separator := "/"
	sections := strings.Split(path, separator)
	if len(sections) == 1 {
		separator = "\\"
		sections = strings.Split(path, separator)
	}

	// Now we start building the root path, making sure to leave out the sub-directory that includes the pattern.
	rootPath := ""
	for _, section := range sections {
		if section == "" {
			continue
		}
		if useRegExp {
			if strings.Index(section, "(") != -1 {
				break
			}
		} else {
			if strings.Index(section, "*") != -1 {
				break
			}
			if strings.Index(section, "(") != -1 {
				temp := rootPath + section
				if isWildcardParentheses(temp, parentheses) {
					break
				}
			}
		}
		if rootPath != "" {
			rootPath += separator
		}
		if section == "~" {
			rootPath += GetUserHomeDir()
		} else {
			rootPath += section
		}
	}
	if len(sections) > 0 && sections[0] == "" {
		rootPath = separator + rootPath
	}
	if rootPath == "" {
		return "."
	}
	return rootPath
}

// Return true if the ‘str’ argument contains open parentasis, that is related to a placeholder.
// The ‘parentheses’ argument contains all the indexes of placeholder parentheses.
func isWildcardParentheses(str string, parentheses ParenthesesSlice) bool {
	toFind := "("
	currStart := 0
	for {
		idx := strings.Index(str, toFind)
		if idx == -1 {
			break
		}
		if parentheses.IsPresent(idx) {
			return true
		}
		currStart += idx + len(toFind)
		str = str[idx+len(toFind):]
	}
	return false
}

func StringToBool(boolVal string, defaultValue bool) (bool, error) {
	if len(boolVal) > 0 {
		result, err := strconv.ParseBool(boolVal)
		errorutils.CheckError(err)
		return result, err
	}
	return defaultValue, nil
}

func AddTrailingSlashIfNeeded(url string) string {
	if url != "" && !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url
}

func IndentJson(jsonStr []byte) string {
	return doIndentJson(jsonStr, "", "  ")
}

func IndentJsonArray(jsonStr []byte) string {
	return doIndentJson(jsonStr, "  ", "  ")
}

func doIndentJson(jsonStr []byte, prefix, indent string) string {
	var content bytes.Buffer
	err := json.Indent(&content, jsonStr, prefix, indent)
	if err == nil {
		return content.String()
	}
	return string(jsonStr)
}

func MergeMaps(src map[string]string, dst map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}

func CopyMap(src map[string]string) (dst map[string]string) {
	dst = make(map[string]string)
	for k, v := range src {
		dst[k] = v
	}
	return
}

func PrepareLocalPathForUpload(localPath string, useRegExp bool) string {
	if localPath == "./" || localPath == ".\\" {
		return "^.*$"
	}
	if strings.HasPrefix(localPath, "./") {
		localPath = localPath[2:]
	} else if strings.HasPrefix(localPath, ".\\") {
		localPath = localPath[3:]
	}
	if !useRegExp {
		localPath = pathToRegExp(cleanPath(localPath))
	}
	return localPath
}

// Clean /../ | /./ using filepath.Clean.
func cleanPath(path string) string {
	temp := path[len(path)-1:]
	path = filepath.Clean(path)
	if temp == `\` || temp == "/" {
		path += temp
	}
	// Since filepath.Clean replaces \\ with \, we revert this action.
	path = strings.Replace(path, `\`, `\\`, -1)
	return path
}

func pathToRegExp(localPath string) string {
	var SPECIAL_CHARS = []string{".", "^", "$", "+"}
	for _, char := range SPECIAL_CHARS {
		localPath = strings.Replace(localPath, char, "\\"+char, -1)
	}
	var wildcard = ".*"
	localPath = strings.Replace(localPath, "*", wildcard, -1)
	if strings.HasSuffix(localPath, "/") || strings.HasSuffix(localPath, "\\") {
		localPath += wildcard
	}
	return "^" + localPath + "$"
}

// Replaces matched regular expression from path to corresponding placeholder {i} at target.
// Example 1:
//      pattern = "repoA/1(.*)234" ; path = "repoA/1hello234" ; target = "{1}" ; ignoreRepo = false
//      returns "hello"
// Example 2:
//      pattern = "repoA/1(.*)234" ; path = "repoB/1hello234" ; target = "{1}" ; ignoreRepo = true
//      returns "hello"
func BuildTargetPath(pattern, path, target string, ignoreRepo bool) (string, error) {
	asteriskIndex := strings.Index(pattern, "*")
	slashIndex := strings.Index(pattern, "/")
	if shouldRemoveRepo(ignoreRepo, asteriskIndex, slashIndex) {
		// Removing the repository part of the path is required when working with virtual repositories, as the pattern
		// may contain the virtual-repository name, but the path contains the local-repository name.
		pattern = removeRepoFromPath(pattern)
		path = removeRepoFromPath(path)
	}
	pattern = addEscapingParentheses(pattern, target)
	pattern = pathToRegExp(pattern)
	if slashIndex < 0 {
		// If '/' doesn't exist, add an optional trailing-slash to support cases in which the provided pattern
		// is only the repository name.
		dollarIndex := strings.LastIndex(pattern, "$")
		pattern = pattern[:dollarIndex]
		pattern += "(/.*)?$"
	}

	r, err := regexp.Compile(pattern)
	err = errorutils.CheckError(err)
	if err != nil {
		return "", err
	}

	groups := r.FindStringSubmatch(path)
	size := len(groups)
	if size > 0 {
		for i := 1; i < size; i++ {
			group := strings.Replace(groups[i], "\\", "/", -1)
			target = strings.Replace(target, "{"+strconv.Itoa(i)+"}", group, -1)
		}
	}
	return target, nil
}

func GetLogMsgPrefix(threadId int, dryRun bool) string {
	var strDryRun string
	if dryRun {
		strDryRun = "[Dry run] "
	}
	return "[Thread " + strconv.Itoa(threadId) + "] " + strDryRun
}

func TrimPath(path string) string {
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.Replace(path, "../", "", -1)
	path = strings.Replace(path, "./", "", -1)
	return path
}

func Bool2Int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func ReplaceTildeWithUserHome(path string) string {
	if len(path) > 1 && path[0:1] == "~" {
		return GetUserHomeDir() + path[1:]
	}
	return path
}

func GetUserHomeDir() string {
	if IsWindows() {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return strings.Replace(home, "\\", "\\\\", -1)
	}
	return os.Getenv("HOME")
}

func GetBoolEnvValue(flagName string, defValue bool) (bool, error) {
	envVarValue := os.Getenv(flagName)
	if envVarValue == "" {
		return defValue, nil
	}
	val, err := strconv.ParseBool(envVarValue)
	err = CheckErrorWithMessage(err, "can't parse environment variable "+flagName)
	return val, err
}

func CheckErrorWithMessage(err error, message string) error {
	if err != nil {
		log.Error(message)
		err = errorutils.CheckError(err)
	}
	return err
}

func ConvertSliceToMap(slice []string) map[string]bool {
	mapFromSlice := make(map[string]bool)
	for _, value := range slice {
		mapFromSlice[value] = true
	}
	return mapFromSlice
}

func removeRepoFromPath(path string) string {
	if idx := strings.Index(path, "/"); idx != -1 {
		return path[idx:]
	}
	return path
}

func shouldRemoveRepo(ignoreRepo bool, asteriskIndex, slashIndex int) bool {
	if !ignoreRepo || slashIndex < 0 {
		return false
	}
	if asteriskIndex < 0 {
		return true
	}
	return IsSlashPrecedeAsterisk(asteriskIndex, slashIndex)
}

func IsSlashPrecedeAsterisk(asteriskIndex, slashIndex int) bool {
	return slashIndex < asteriskIndex && slashIndex >= 0
}

// Split str by the provided separator, escaping the separator if it is prefixed by a back-slash.
func SplitWithEscape(str string, separator rune) []string {
	var parts []string
	var current bytes.Buffer
	escaped := false
	for _, char := range str {
		if char == '\\' {
			if escaped {
				current.WriteRune(char)
			}
			escaped = true
		} else if char == separator && !escaped {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			escaped = false
			current.WriteRune(char)
		}
	}
	parts = append(parts, current.String())
	return parts
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

type Artifact struct {
	LocalPath  string
	TargetPath string
	Symlink    string
}
