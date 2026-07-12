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

type DockerFederatedRepositoryParams struct {
	local.DockerLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func unpackLocalDockerV2Repository(data *schema.ResourceData, Rclass string) local.DockerLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return local.DockerLocalRepositoryParams{
		RepositoryBaseParams: local.UnpackBaseRepo(Rclass, data, repository.DockerPackageType),
		MaxUniqueTags:        d.GetInt("max_unique_tags", false),
		DockerApiVersion:     "V2",
		TagRetention:         d.GetInt("tag_retention", false),
		BlockPushingSchema1:  d.GetBool("block_pushing_schema1", false),
	}
}

func ResourceArtifactoryFederatedDockerV2Repository() *schema.Resource {
	dockerV2FederatedSchema := lo.Assign(
		local.DockerV2Schemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, repository.DockerPackageType),
	)

	var unpackFederatedDockerRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := DockerFederatedRepositoryParams{
			DockerLocalRepositoryParams: unpackLocalDockerV2Repository(data, Rclass),
			Members:                     unpackMembers(data),
			RepoParams:                  unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packDockerMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*DockerFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packDockerMembers,
	)

	constructor := func() (interface{}, error) {
		return &DockerFederatedRepositoryParams{
			DockerLocalRepositoryParams: local.DockerLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: repository.DockerPackageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(dockerV2FederatedSchema, pkr, unpackFederatedDockerRepository, constructor)
}

func unpackLocalDockerV1Repository(data *schema.ResourceData, Rclass string) local.DockerLocalRepositoryParams {
	return local.DockerLocalRepositoryParams{
		RepositoryBaseParams: local.UnpackBaseRepo(Rclass, data, repository.DockerPackageType),
		DockerApiVersion:     "V1",
		MaxUniqueTags:        0,
		TagRetention:         1,
		BlockPushingSchema1:  false,
	}
}

func ResourceArtifactoryFederatedDockerV1Repository() *schema.Resource {
	dockerFederatedSchema := lo.Assign(
		local.DockerV1Schemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, repository.DockerPackageType),
	)

	var unpackFederatedDockerRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := DockerFederatedRepositoryParams{
			DockerLocalRepositoryParams: unpackLocalDockerV1Repository(data, Rclass),
			Members:                     unpackMembers(data),
			RepoParams:                  unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packDockerMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*DockerFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.Ignore("class", "rclass", "member", "terraform_type"),
		),
		packDockerMembers,
	)

	constructor := func() (interface{}, error) {
		return &DockerFederatedRepositoryParams{
			DockerLocalRepositoryParams: local.DockerLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: repository.DockerPackageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(dockerFederatedSchema, pkr, unpackFederatedDockerRepository, constructor)
}
