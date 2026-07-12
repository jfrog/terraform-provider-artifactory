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
	"github.com/samber/lo"
)

type ReleasebundlesRepositoryParams struct {
	local.RepositoryBaseParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedReleasebundlesRepository(packageType string) *schema.Resource {
	var genericSchema = lo.Assign(
		local.GetGenericSchemas(packageType)[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
	)

	var unpackFederatedRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := ReleasebundlesRepositoryParams{
			RepositoryBaseParams: local.UnpackBaseRepo(Rclass, data, packageType),
			Members:              unpackMembers(data),
			RepoParams:           unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packGenericMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*ReleasebundlesRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packGenericMembers,
	)

	constructor := func() (interface{}, error) {
		return &ReleasebundlesRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: local.GetPackageType(packageType),
				Rclass:      Rclass,
			},
		}, nil
	}

	return mkResourceSchema(genericSchema, pkr, unpackFederatedRepository, constructor)
}
