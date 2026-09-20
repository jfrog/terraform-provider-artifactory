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

type ComposerFederatedRepositoryParams struct {
	local.ComposerLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func unpackLocalComposerRepository(data *schema.ResourceData, Rclass string) local.ComposerLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return local.ComposerLocalRepositoryParams{
		RepositoryBaseParams:     local.UnpackBaseRepo(Rclass, data, repository.ComposerPackageType),
		EnableComposerV1Indexing: d.GetBool("enable_composer_v1_indexing", false),
	}
}

func ResourceArtifactoryFederatedComposerRepository() *schema.Resource {
	composerFederatedSchema := utilsdk.MergeMaps(
		local.ComposerSchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, repository.ComposerPackageType),
	)

	var unpackFederatedComposerRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := ComposerFederatedRepositoryParams{
			ComposerLocalRepositoryParams: unpackLocalComposerRepository(data, Rclass),
			Members:                       unpackMembers(data),
			RepoParams:                    unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packComposerMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*ComposerFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packComposerMembers,
	)

	constructor := func() (interface{}, error) {
		return &ComposerFederatedRepositoryParams{
			ComposerLocalRepositoryParams: local.ComposerLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: repository.ComposerPackageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(composerFederatedSchema, pkr, unpackFederatedComposerRepository, constructor)
}
