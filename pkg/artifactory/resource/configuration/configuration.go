package configuration

import (
	"github.com/jfrog/terraform-provider-shared/client"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

/*
	SendConfigurationPatch updates system configuration using YAML data.

See https://www.jfrog.com/confluence/display/JFROG/Artifactory+YAML+Configuration
*/
func SendConfigurationPatch(content []byte, m interface{}) error {
	_, err := m.(utilsdk.ProvderMetadata).Client.R().SetBody(content).
		SetHeader("Content-Type", "application/yaml").
		AddRetryCondition(client.RetryOnMergeError).
		Patch("artifactory/api/system/configuration")

	return err
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
