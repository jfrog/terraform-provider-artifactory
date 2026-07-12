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

type NpmRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryCurationParams
}

var NPMSchema = lo.Assign(
	remote.BaseSchema,
	remote.CurationRemoteRepoSchema,
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.NPMPackageType),
)

var NPMSchemas = remote.GetSchemas(NPMSchema)

func DataSourceArtifactoryRemoteNpmRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.NPMPackageType)
		if err != nil {
			return nil, err
		}

		return &NpmRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.NPMPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	npmSchema := getSchema(NPMSchemas)

	return &schema.Resource{
		Schema:      npmSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(npmSchema), constructor),
		Description: "Provides a data source for a remote NPM repository",
	}
}
