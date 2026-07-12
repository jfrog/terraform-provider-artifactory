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

type ConanRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryCurationParams
	resource_repository.ConanBaseParams
}

var conanSchema = lo.Assign(
	remote.BaseSchema,
	remote.CurationRemoteRepoSchema,
	resource_repository.ConanBaseSchemaSDKv2,
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.ConanPackageType),
)

var ConanSchemas = remote.GetSchemas(conanSchema)

func DataSourceArtifactoryRemoteConanRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.ConanPackageType)
		if err != nil {
			return nil, err
		}

		return &ConanRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.ConanPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	conanSchema := getSchema(ConanSchemas)

	return &schema.Resource{
		Schema:      conanSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(conanSchema), constructor),
		Description: "Provides a data source for a remote Conan repository",
	}
}
