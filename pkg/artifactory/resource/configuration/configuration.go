package configuration

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/jfrog/terraform-provider-shared/client"
)

const ConfigurationEndpoint = "artifactory/api/system/configuration"

/*
	SendConfigurationPatch updates system configuration using YAML data.

See https://www.jfrog.com/confluence/display/JFROG/Artifactory+YAML+Configuration
*/
func SendConfigurationPatch(content []byte, restyClient *resty.Client) error {
	resp, err := restyClient.R().SetBody(content).
		SetHeader("Content-Type", "application/yaml").
		AddRetryCondition(client.RetryOnMergeError).
		Patch(ConfigurationEndpoint)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("%s", resp.String())
	}

	return nil
}

type Configuration interface {
	Id() string
}

func FindConfigurationById[C Configuration](configurations []C, id string) *C {
	for _, configuration := range configurations {
		if configuration.Id() == id {
			return &configuration
		}
	}
	return nil
}
