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

type BowerRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryVcsParams
	BowerRegistryUrl string `json:"bowerRegistryUrl"`
}

var bowerSchema = lo.Assign(
	remote.BaseSchema,
	VcsRemoteRepoSchemaSDKv2,
	map[string]*schema.Schema{
		"bower_registry_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://registry.bower.io",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `Proxy remote Bower repository. Default value is "https://registry.bower.io".`,
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.BowerPackageType),
)

var BowerSchemas = remote.GetSchemas(bowerSchema)

func DataSourceArtifactoryRemoteBowerRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.BowerPackageType)
		if err != nil {
			return nil, err
		}

		return &BowerRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.BowerPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	bowerSchema := getSchema(BowerSchemas)

	return &schema.Resource{
		Schema:      bowerSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(bowerSchema), constructor),
		Description: "Provides a data source for a remote Bower repository",
	}
}
