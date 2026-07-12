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
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

type VcsRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryVcsParams
	MaxUniqueSnapshots int `json:"maxUniqueSnapshots"`
}

var VCSSchema = lo.Assign(
	remote.BaseSchema,
	VcsRemoteRepoSchemaSDKv2,
	map[string]*schema.Schema{
		"max_unique_snapshots": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			Description: "The maximum number of unique snapshots of a single artifact to store. Once the number of " +
				"snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is " +
				"no limit, and unique snapshots are not cleaned up.",
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.VCSPackageType),
)

var VCSSchemas = remote.GetSchemas(VCSSchema)

func DataSourceArtifactoryRemoteVcsRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.VCSPackageType)
		if err != nil {
			return nil, err
		}

		return &VcsRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.VCSPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	vcsSchema := getSchema(VCSSchemas)

	return &schema.Resource{
		Schema:      vcsSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(vcsSchema), constructor),
		Description: "Provides a data source for a remote VCS repository",
	}
}
