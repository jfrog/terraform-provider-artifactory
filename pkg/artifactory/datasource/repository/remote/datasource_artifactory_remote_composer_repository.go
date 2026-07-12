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

type ComposerRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryVcsParams
	ComposerRegistryUrl string `json:"composerRegistryUrl"`
}

var composerSchema = lo.Assign(
	remote.BaseSchema,
	VcsRemoteRepoSchemaSDKv2,
	map[string]*schema.Schema{
		"composer_registry_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://packagist.org",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `Proxy remote Composer repository. Default value is "https://packagist.org".`,
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.ComposerPackageType),
)

var ComposerSchemas = remote.GetSchemas(composerSchema)

func DataSourceArtifactoryRemoteComposerRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.ComposerPackageType)
		if err != nil {
			return nil, err
		}

		return &ComposerRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.ComposerPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	composerSchema := getSchema(ComposerSchemas)

	return &schema.Resource{
		Schema:      composerSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(composerSchema), constructor),
		Description: "Provides a data source for a remote Composer repository",
	}
}
