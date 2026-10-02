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
	"github.com/samber/lo"
)

type TerraformFederatedRepositoryParams struct {
	local.TerraformLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func unpackLocalTerraformRepository(data *schema.ResourceData, Rclass string, registryType string) local.TerraformLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	repo := local.UnpackBaseRepo(Rclass, data, "terraform_"+registryType)
	repo.TerraformType = registryType

	return local.TerraformLocalRepositoryParams{
		RepositoryBaseParams: repo,
		PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
			PrimaryKeyPairRefSDKv2: d.GetString("primary_keypair_ref", false),
		},
		SecondaryKeyPairRefParam: repository.SecondaryKeyPairRefParam{
			SecondaryKeyPairRefSDKv2: d.GetString("secondary_keypair_ref", false),
		},
	}
}

func ResourceArtifactoryFederatedTerraformRepository(registryType string) *schema.Resource {
	packageType := "terraform_" + registryType

	terraformFederatedSchema := lo.Assign(
		local.GetTerraformSchemas(registryType)[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
	)

	var unpackFederatedTerraformRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := TerraformFederatedRepositoryParams{
			TerraformLocalRepositoryParams: unpackLocalTerraformRepository(data, Rclass, registryType),
			Members:                        unpackMembers(data),
			RepoParams:                     unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packTerraformMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*TerraformFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packTerraformMembers,
	)

	constructor := func() (interface{}, error) {
		return &TerraformFederatedRepositoryParams{
			TerraformLocalRepositoryParams: local.TerraformLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(terraformFederatedSchema, pkr, unpackFederatedTerraformRepository, constructor)
}
