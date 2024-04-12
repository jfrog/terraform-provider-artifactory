package configuration

import (
	"fmt"

	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
)

/*
	SendConfigurationPatch updates system configuration using YAML data.

See https://www.jfrog.com/confluence/display/JFROG/Artifactory+YAML+Configuration
*/
func SendConfigurationPatch(content []byte, m interface{}) error {
	resp, err := m.(util.ProviderMetadata).Client.R().SetBody(content).
		SetHeader("Content-Type", "application/yaml").
		AddRetryCondition(client.RetryOnMergeError).
		Patch("artifactory/api/system/configuration")

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
