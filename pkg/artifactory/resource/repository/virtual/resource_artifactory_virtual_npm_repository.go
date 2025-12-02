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
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var npmSchema = lo.Assign(
	RetrievalCachePeriodSecondsSchema,
	externalDependenciesSchema,
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.NPMPackageType),
)

var NPMSchemas = GetSchemas(npmSchema)

type NpmVirtualRepositoryParams struct {
	ExternalDependenciesVirtualRepositoryParams
	VirtualRetrievalCachePeriodSecs int `hcl:"retrieval_cache_period_seconds" json:"virtualRetrievalCachePeriodSecs"`
}

func ResourceArtifactoryVirtualNpmRepository() *schema.Resource {
	var unpackNpmVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := NpmVirtualRepositoryParams{
			VirtualRetrievalCachePeriodSecs:             d.GetInt("retrieval_cache_period_seconds", false),
			ExternalDependenciesVirtualRepositoryParams: unpackExternalDependenciesVirtualRepository(s, repository.NPMPackageType),
		}
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &NpmVirtualRepositoryParams{
			ExternalDependenciesVirtualRepositoryParams: ExternalDependenciesVirtualRepositoryParams{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: repository.NPMPackageType,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(
		NPMSchemas,
		packer.Default(NPMSchemas[CurrentSchemaVersion]),
		unpackNpmVirtualRepository,
		constructor,
	)
}
