package utils

import (
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"regexp"
	"strings"
)

const CredentialsInUrlRegexp = `((http|https):\/\/.+:.*@)`

func GetRegExp(regex string) (*regexp.Regexp, error) {
	regExp, err := regexp.Compile(regex)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	return regExp, nil
}

// Mask the credentials information from the line, contained in credentialsPart.
// The credentials are built as user:password
// For example:
// line = 'This is a line http://user:password@127.0.0.1:8081/artifactory/path/to/repo'
// credentialsPart = 'http://user:password@'
// Returned value: 'This is a line http://***:***@127.0.0.1:8081/artifactory/path/to/repo'
func MaskCredentials(line, credentialsPart string) string {
	splitResult := strings.Split(credentialsPart, "//")
	return strings.Replace(line, credentialsPart, splitResult[0]+"//***.***@", 1)
}
