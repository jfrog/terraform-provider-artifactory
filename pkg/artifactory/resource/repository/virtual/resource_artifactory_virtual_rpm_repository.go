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

var rpmSchema = lo.Assign(
	repository.PrimaryKeyPairRefSDKv2,
	repository.SecondaryKeyPairRefSDKv2,
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.RPMPackageType),
)

var RPMSchemas = GetSchemas(rpmSchema)

func ResourceArtifactoryVirtualRpmRepository() *schema.Resource {
	type CommonRpmDebianVirtualRepositoryParams struct {
		repository.PrimaryKeyPairRefParam
		repository.SecondaryKeyPairRefParam
	}

	type RpmVirtualRepositoryParams struct {
		RepositoryBaseParams
		CommonRpmDebianVirtualRepositoryParams
	}

	var unpackRpmVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := RpmVirtualRepositoryParams{
			RepositoryBaseParams: UnpackBaseVirtRepo(s, "rpm"),
			CommonRpmDebianVirtualRepositoryParams: CommonRpmDebianVirtualRepositoryParams{
				PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
					PrimaryKeyPairRefSDKv2: d.GetString("primary_keypair_ref", false),
				},
				SecondaryKeyPairRefParam: repository.SecondaryKeyPairRefParam{
					SecondaryKeyPairRefSDKv2: d.GetString("secondary_keypair_ref", false),
				},
			},
		}
		repo.PackageType = repository.RPMPackageType

		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &RpmVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: repository.RPMPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		RPMSchemas,
		packer.Default(RPMSchemas[CurrentSchemaVersion]),
		unpackRpmVirtualRepository,
		constructor,
	)
}
