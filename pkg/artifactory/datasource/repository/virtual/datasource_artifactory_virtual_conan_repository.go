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

package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DatasourceArtifactoryVirtualConanRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(virtual.Rclass, resource_repository.ConanPackageType)
		if err != nil {
			return nil, err
		}

		return &virtual.ConanRepoParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: virtual.RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: virtual.RepositoryBaseParams{
					Rclass:        virtual.Rclass,
					PackageType:   resource_repository.ConanPackageType,
					RepoLayoutRef: repoLayout,
				},
			},
			ConanBaseParams: resource_repository.ConanBaseParams{
				EnableConanSupport: true,
			},
		}, nil
	}

	conanSchema := virtual.ConanSchemas[virtual.CurrentSchemaVersion]

	return &schema.Resource{
		Schema:      conanSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(conanSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", resource_repository.ConanPackageType),
	}
}
