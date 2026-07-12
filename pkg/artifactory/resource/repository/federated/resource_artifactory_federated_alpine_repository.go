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

package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type AlpineRepositoryParams struct {
	local.AlpineLocalRepoParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func unpackLocalAlpineRepository(data *schema.ResourceData, Rclass string) local.AlpineLocalRepoParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return local.AlpineLocalRepoParams{
		RepositoryBaseParams: local.UnpackBaseRepo(Rclass, data, repository.AlpinePackageType),
		PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
			PrimaryKeyPairRefSDKv2: d.GetString("primary_keypair_ref", false),
		},
	}
}

func ResourceArtifactoryFederatedAlpineRepository() *schema.Resource {
	alpineFederatedSchema := utilsdk.MergeMaps(
		local.AlpineLocalSchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, repository.AlpinePackageType),
	)

	var unpackFederatedAlpineRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := AlpineRepositoryParams{
			AlpineLocalRepoParams: unpackLocalAlpineRepository(data, Rclass),
			Members:               unpackMembers(data),
			RepoParams:            unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packAlpineMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*AlpineRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packAlpineMembers,
	)

	constructor := func() (interface{}, error) {
		return &AlpineRepositoryParams{
			AlpineLocalRepoParams: local.AlpineLocalRepoParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: repository.AlpinePackageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(alpineFederatedSchema, pkr, unpackFederatedAlpineRepository, constructor)
}
