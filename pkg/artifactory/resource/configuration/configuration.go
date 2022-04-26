package configuration

import (
	"github.com/go-resty/resty/v2"
	"github.com/jfrog/terraform-provider-shared/util"
	"gopkg.in/ldap.v2"
)

func SendConfigurationPatch(content []byte, m interface{}) error {
	_, err := m.(*resty.Client).R().SetBody(content).
		SetHeader("Content-Type", "application/yaml").
		AddRetryCondition(util.RetryOnMergeError).
		Patch("artifactory/api/system/configuration")

	return err
}

func ValidateLdapDn(value interface{}, _ string) ([]string, []error) {
	_, err := ldap.ParseDN(value.(string))
	if err != nil {
		return nil, []error{err}
	}
	return nil, nil
}

func ValidateLdapFilter(value interface{}, _ string) ([]string, []error) {
	_, err := ldap.CompileFilter(value.(string))
	if err != nil {
		return nil, []error{err}
	}
	return nil, nil
}
