package artifactory

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/mail"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
	return nil, nil
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


func validateIsEmail(address interface{}, _ string) ([]string, []error) {
	_, err := mail.ParseAddress(address.(string))
	if err != nil {
		return nil, []error{fmt.Errorf("%s is not a valid address: %s", address, err)}
	}
	return nil, nil
}

var repoKeyValidator = validation.All(
	validation.StringDoesNotMatch(regexp.MustCompile("^[0-9].*"), "repo key cannot start with a number"),
	validation.StringDoesNotContainAny(" !@#$%^&*()_+={}[]:;<>,./?~`|\\"),
)

func fileExist(value interface{}, _ string) ([]string, []error) {
	if _, err := os.Stat(value.(string)); err != nil {
		return nil, []error{err}
	}
	return nil, nil
}

var defaultPassValidation = validation.All(
	validation.StringMatch(regexp.MustCompile("[0-9]+"), "password must contain at least 1 digit case char"),
	validation.StringMatch(regexp.MustCompile("[a-z]+"), "password must contain at least 1 lower case char"),
	validation.StringMatch(regexp.MustCompile("[A-Z]+"), "password must contain at least 1 upper case char"),
	minLength(8),
)

var sliceIs = func (slice ... interface{}) schema.SchemaValidateFunc{
	return func (value interface{}, _ string) ([]string, []error){
		for _, e := range slice {
			if e == value {
				return nil, nil
			}
		}
		return nil, []error{fmt.Errorf("value %s not found in %q",value, slice)}
	}
}

func minLength(length int) func(i interface{}, k string) ([]string, []error) {
	return func(value interface{}, k string) ([]string, []error) {
		if len(value.(string)) < length {
			return nil, []error{fmt.Errorf("password must be atleast %d characters long",length)}
		}
		return nil, nil
	}
}
