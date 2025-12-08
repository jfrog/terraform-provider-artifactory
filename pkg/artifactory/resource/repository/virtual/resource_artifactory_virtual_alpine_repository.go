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

var alpineSchema = lo.Assign(
	RetrievalCachePeriodSecondsSchema,
	repository.PrimaryKeyPairRefSDKv2,
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.AlpinePackageType),
)

var AlpineSchemas = GetSchemas(alpineSchema)

func ResourceArtifactoryVirtualAlpineRepository() *schema.Resource {
	type AlpineVirtualRepositoryParams struct {
		RepositoryBaseParamsWithRetrievalCachePeriodSecs
		repository.PrimaryKeyPairRefParam
	}

	var unpackAlpineVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := AlpineVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(s, repository.AlpinePackageType),
			PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
				PrimaryKeyPairRefSDKv2: d.GetString("primary_keypair_ref", false),
			},
		}
		repo.PackageType = repository.AlpinePackageType
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &AlpineVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: repository.AlpinePackageType,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(
		AlpineSchemas,
		packer.Default(AlpineSchemas[CurrentSchemaVersion]),
		unpackAlpineVirtualRepository,
		constructor,
	)
}
