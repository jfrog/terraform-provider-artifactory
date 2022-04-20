package configuration

import (
	"github.com/go-resty/resty/v2"

	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func SendConfigurationPatch(content []byte, m interface{}) error {
	_, err := m.(*resty.Client).R().SetBody(content).
		SetHeader("Content-Type", "application/yaml").
		AddRetryCondition(utils.RetryOnMergeError).
		Patch("artifactory/api/system/configuration")

	return err
}
