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

package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

type TerraformRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	TerraformRegistryUrl  string `json:"terraformRegistryUrl"`
	TerraformProvidersUrl string `json:"terraformProvidersUrl"`
}

var TerraformSchema = lo.Assign(
	remote.BaseSchema,
	map[string]*schema.Schema{
		"terraform_registry_url": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Default:      "https://registry.terraform.io",
			Description: "The base URL of the registry API. When using Smart Remote Repositories, set the URL to" +
				" <base_Artifactory_URL>/api/terraform/repokey. Default value in UI is https://registry.terraform.io",
		},
		"terraform_providers_url": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Default:      "https://releases.hashicorp.com",
			Description: "The base URL of the Provider's storage API. When using Smart remote repositories, set " +
				"the URL to <base_Artifactory_URL>/api/terraform/repokey/providers. Default value in UI is https://releases.hashicorp.com",
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.TerraformPackageType),
)

var TerraformSchemas = remote.GetSchemas(TerraformSchema)

func DataSourceArtifactoryRemoteTerraformRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.TerraformPackageType)
		if err != nil {
			return nil, err
		}

		return &TerraformRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.TerraformPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	terraformSchema := getSchema(TerraformSchemas)

	return &schema.Resource{
		Schema:      terraformSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(terraformSchema), constructor),
		Description: "Provides a data source for a remote Terraform repository",
	}
}
