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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

func ResourceArtifactoryVirtualGenericRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		return &RepositoryBaseParams{
			PackageType: packageType,
			Rclass:      Rclass,
		}, nil
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseVirtRepo(data, packageType)
		return repo, repo.Id(), nil
	}

	genericSchemas := GetSchemas(repository.RepoLayoutRefSDKv2Schema(Rclass, packageType))

	return repository.MkResourceSchema(
		genericSchemas,
		packer.Default(genericSchemas[CurrentSchemaVersion]),
		unpack,
		constructor,
	)
}

var RepoWithRetrivalCachePeriodSecsVirtualSchemas = func(packageType string) map[int16]map[string]*schema.Schema {
	var repoWithRetrivalCachePeriodSecsVirtualSchema = lo.Assign(
		RetrievalCachePeriodSecondsSchema,
		repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
	)

	return GetSchemas(repoWithRetrivalCachePeriodSecsVirtualSchema)
}

func ResourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(packageType string) *schema.Resource {
	repoWithRetrivalCachePeriodSecsVirtualSchemas := RepoWithRetrivalCachePeriodSecsVirtualSchemas(packageType)

	constructor := func() (interface{}, error) {
		return &RepositoryBaseParamsWithRetrievalCachePeriodSecs{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: packageType,
			},
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(data, packageType)
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(
		repoWithRetrivalCachePeriodSecsVirtualSchemas,
		packer.Default(repoWithRetrivalCachePeriodSecsVirtualSchemas[CurrentSchemaVersion]),
		unpack,
		constructor,
	)
}
