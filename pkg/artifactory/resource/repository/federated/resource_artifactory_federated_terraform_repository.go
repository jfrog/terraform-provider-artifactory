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
	local.RepositoryBaseParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
	repository.PrimaryKeyPairRefParam
}

func unpackLocalTerraformRepository(data *schema.ResourceData, Rclass string, registryType string) local.RepositoryBaseParams {
	repo := local.UnpackBaseRepo(Rclass, data, "terraform_"+registryType)
	repo.TerraformType = registryType

	return repo
}

func ResourceArtifactoryFederatedTerraformRepository(registryType string) *schema.Resource {
	packageType := "terraform_" + registryType
	isProvider := registryType == "provider"

	terraformFederatedSchema := lo.Assign(
		local.GetTerraformSchemas(registryType)[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
	)
	if isProvider {
		terraformFederatedSchema = lo.Assign(
			terraformFederatedSchema,
			repository.PrimaryKeyPairRefSDKv2,
		)
	}

	var unpackFederatedTerraformRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}
		repo := TerraformFederatedRepositoryParams{
			RepositoryBaseParams: unpackLocalTerraformRepository(data, Rclass, registryType),
			Members:              unpackMembers(data),
			RepoParams:           unpackRepoParams(data),
		}
		if isProvider {
			repo.PrimaryKeyPairRefParam = repository.PrimaryKeyPairRefParam{
				PrimaryKeyPairRefSDKv2: d.GetString("primary_keypair_ref", false),
			}
		}
		return repo, repo.Id(), nil
	}

	var packTerraformMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*TerraformFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	ignoreFields := []string{"member", "terraform_type"}
	if !isProvider {
		ignoreFields = append(ignoreFields, "primary_keypair_ref")
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore(ignoreFields...),
			),
		),
		packTerraformMembers,
	)

	constructor := func() (interface{}, error) {
		return &TerraformFederatedRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: packageType,
				Rclass:      Rclass,
			},
		}, nil
	}

	return mkResourceSchema(terraformFederatedSchema, pkr, unpackFederatedTerraformRepository, constructor)
}
