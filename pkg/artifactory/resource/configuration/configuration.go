// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
