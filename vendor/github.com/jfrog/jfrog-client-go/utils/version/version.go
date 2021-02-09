package version

import (
	"github.com/jfrog/jfrog-client-go/utils"
	"strconv"
	"strings"
)

type Version struct {
	version string
}

func NewVersion(version string) *Version {
	return &Version{version: version}
}

func (version *Version) SetVersion(artifactoryVersion string) {
	version.version = artifactoryVersion
}

// If ver1 == version returns 0
// If ver1 > version returns 1
// If ver1 < version returns -1
func (version *Version) Compare(ver1 string) int {
	if ver1 == version.version {
		return 0
	}

	ver1Tokens := strings.Split(ver1, ".")
	ver2Tokens := strings.Split(version.version, ".")

	maxIndex := len(ver1Tokens)
	if len(ver2Tokens) > maxIndex {
		maxIndex = len(ver2Tokens)
	}

	for tokenIndex := 0; tokenIndex < maxIndex; tokenIndex++ {
		ver1Token := "0"
		if len(ver1Tokens) >= tokenIndex+1 {
			ver1Token = strings.TrimSpace(ver1Tokens[tokenIndex])
		}
		ver2Token := "0"
		if len(ver2Tokens) >= tokenIndex+1 {
			ver2Token = strings.TrimSpace(ver2Tokens[tokenIndex])
		}
		compare := compareTokens(ver1Token, ver2Token)
		if compare != 0 {
			return compare
		}
	}

	return 0
}

// Returns true if this version is larger or equals from the version sent as an argument.
func (version *Version) AtLeast(minVersion string) bool {
	if version.Compare(minVersion) > 0 && version.version != utils.Development {
		return false
	}
	return true
}

func compareTokens(ver1Token, ver2Token string) int {
	// Ignoring error because we strip all the non numeric values in advance.
	ver1TokenInt, _ := strconv.Atoi(getFirstNumeral(ver1Token))
	ver2TokenInt, _ := strconv.Atoi(getFirstNumeral(ver2Token))

	switch {
	case ver1TokenInt > ver2TokenInt:
		return 1
	case ver1TokenInt < ver2TokenInt:
		return -1
	default:
		return strings.Compare(ver1Token, ver2Token)
	}
}

func getFirstNumeral(token string) string {
	numeric := ""
	for i := 0; i < len(token); i++ {
		n := token[i : i+1]
		if _, err := strconv.Atoi(n); err != nil {
			return "999999"
		}
		numeric += n
	}
	return numeric
}
