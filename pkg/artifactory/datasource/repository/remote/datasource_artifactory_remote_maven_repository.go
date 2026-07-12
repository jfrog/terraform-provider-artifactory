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
)

type MavenRemoteRepo struct {
	remote.JavaRemoteRepo
	remote.RepositoryCurationParams
}

func DataSourceArtifactoryRemoteMavenRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.MavenPackageType)
		if err != nil {
			return nil, err
		}

		return &MavenRemoteRepo{
			JavaRemoteRepo: remote.JavaRemoteRepo{
				RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
					Rclass:        remote.Rclass,
					PackageType:   resource_repository.MavenPackageType,
					RepoLayoutRef: repoLayout,
				},
				SuppressPomConsistencyChecks: false,
			},
		}, nil
	}

	mavenSchema := remote.MavenSchemas[remote.MavenCurrentSchemaVersion]
	mavenSchema["url"].Required = false
	mavenSchema["url"].Optional = true

	return &schema.Resource{
		Schema:      mavenSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(mavenSchema), constructor),
		Description: "Provides a data source for a remote Maven repository",
	}
}
