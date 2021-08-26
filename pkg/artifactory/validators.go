package artifactory

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"net/mail"
	"regexp"
	"strings"
)

func validateLowerCase(value interface{}, key string) (ws []string, es []error) {
	m := value.(string)
	low := strings.ToLower(m)

	if m != low {
		es = append(es, fmt.Errorf("%s should be lowercase", key))
	}
	return
}

func validateCron(value interface{}, key string) (ws []string, es []error) {
	_, err := cronexpr.Parse(value.(string))
	if err != nil {
		return nil, []error{err}
	}
	return nil,nil
}

var repoTypesSupported = []string{
	"alpine",
	"bower",
	//"cargo", // not supported
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"conda",
	"cran",
	"debian",
	"docker",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"gradle",
	"helm",
	"ivy",
	"maven",
	"npm",
	"nuget",
	"opkg",
	"p2",
	"puppet",
	"pypi",
	"rpm",
	"sbt",
	"vagrant",
	"vcs",
}
var repoTypeValidator = validation.StringInSlice(repoTypesSupported, false)

func toString(i interface{}, key string) (result string, err error) {
	result, ok := i.(string)
	if !ok {
		return "", fmt.Errorf("expected type of %q to be string", key)
	}
	return result, nil
}
func validateIsEmail(i interface{}, k string) ([]string, []error) {
	addr, e := toString(i, k)
	if e == nil {
		_, err := mail.ParseAddress(addr)
		if err != nil {
			return nil, []error{fmt.Errorf("%s is not a valid address: %s", addr, err)}
		}
		return nil, nil
	}
	return nil, []error{e}
}
func match(regex, str, msg string) []error {
	matches, _ := regexp.MatchString(regex, str)
	if !matches {
		return []error{fmt.Errorf(msg)}
	}
	return nil
}
func containsLower(i interface{}, k string) ([]string, []error) {
	str, e := toString(i, k)
	if e == nil {
		return nil, match("[a-z]+", str,
			fmt.Sprintf("password must contain at least 1 lower case char. It was: %s", str),
		)
	}
	return nil, []error{e}
}

var repoKeyValidator = validation.All(
	validation.StringDoesNotMatch(regexp.MustCompile("^[0-9].*"), "repo key cannot start with a number"),
	validation.StringDoesNotContainAny(" !@#$%^&*()_+={}[]:;<>,./?~`|\\"),
)

func containsUpper(i interface{}, k string) ([]string, []error) {
	str, e := toString(i, k)
	if e == nil {
		return nil, match("[A-Z]+", str,
			fmt.Sprintf("password must contain at least 1 upper case char. It was: %s", str),
		)
	}
	return nil, []error{e}
}
func containsDigit(i interface{}, k string) ([]string, []error) {
	str, e := toString(i, k)
	if e == nil {
		return nil, match("[0-9]+", str,
			fmt.Sprintf("password must contain at least 1 digit case char. It was: %s", str),
		)
	}
	return nil, []error{e}
}
func minLength(i interface{}, k string) ([]string, []error) {
	str, e := toString(i, k)
	if e == nil {
		if len(str) < 8 {
			return nil, []error{fmt.Errorf("password must be atleast 8 characters long")}
		}
	}
	return nil, []error{e}
}

