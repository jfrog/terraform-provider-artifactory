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

type AnsibleRepositoryParams struct {
	local.RepositoryBaseParams
	RepoParams
	Members []Member `hcl:"member" json:"members"`
	repository.PrimaryKeyPairRefParam
}

func ResourceArtifactoryFederatedAnsibleRepository() *schema.Resource {
	var ansibleSchema = lo.Assign(
		local.GetGenericSchemas(repository.AnsiblePackageType)[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.AlpinePrimaryKeyPairRefSDKv2,
		repository.RepoLayoutRefSDKv2Schema(Rclass, repository.AnsiblePackageType),
	)

	var unpackFederatedRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}

		repo := AnsibleRepositoryParams{
			RepositoryBaseParams: local.UnpackBaseRepo(Rclass, data, repository.AnsiblePackageType),
			RepoParams:           unpackRepoParams(data),
			Members:              unpackMembers(data),
			PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
				PrimaryKeyPairRefSDKv2: d.GetString("primary_keypair_ref", false),
			},
		}
		return repo, repo.Id(), nil
	}

	var packMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*AnsibleRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packMembers,
	)

	constructor := func() (interface{}, error) {
		return &AnsibleRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: repository.AnsiblePackageType,
				Rclass:      Rclass,
			},
		}, nil
	}

	return mkResourceSchema(ansibleSchema, pkr, unpackFederatedRepository, constructor)
}
