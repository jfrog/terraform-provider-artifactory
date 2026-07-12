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

type CocoapodsRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryVcsParams
	PodsSpecsRepoUrl string `json:"podsSpecsRepoUrl"`
}

var cocoapodsSchema = lo.Assign(
	remote.BaseSchema,
	VcsRemoteRepoSchemaSDKv2,
	map[string]*schema.Schema{
		"pods_specs_repo_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://github.com/CocoaPods/Specs",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `Proxy remote CocoaPods Specs repositories. Default value is "https://github.com/CocoaPods/Specs".`,
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.CocoapodsPackageType),
)

var CocoapodsSchemas = remote.GetSchemas(cocoapodsSchema)

func DataSourceArtifactoryRemoteCoapodsRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.CocoapodsPackageType)
		if err != nil {
			return nil, err
		}

		return &CocoapodsRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.CocoapodsPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	cocoapodsSchema := getSchema(CocoapodsSchemas)

	return &schema.Resource{
		Schema:      cocoapodsSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(cocoapodsSchema), constructor),
		Description: "Provides a data source for a remote CocoaPods repository",
	}
}
