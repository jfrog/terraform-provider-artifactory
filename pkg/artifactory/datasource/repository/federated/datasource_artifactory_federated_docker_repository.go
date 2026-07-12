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
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/samber/lo"
)

func DataSourceArtifactoryFederatedDockerV2Repository() *schema.Resource {
	dockerV2FederatedSchema := lo.Assign(
		local.DockerV2Schemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		resource_repository.RepoLayoutRefSDKv2Schema(federated.Rclass, resource_repository.DockerPackageType),
	)

	var packDockerMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.DockerFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
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
		return &federated.DockerFederatedRepositoryParams{
			DockerLocalRepositoryParams: local.DockerLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: resource_repository.DockerPackageType,
					Rclass:      federated.Rclass,
				},
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      dockerV2FederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: "Provides a data source for a federated docker V2 repository",
	}
}

func DataSourceArtifactoryFederatedDockerV1Repository() *schema.Resource {
	dockerFederatedSchema := lo.Assign(
		local.DockerV1Schemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		resource_repository.RepoLayoutRefSDKv2Schema(federated.Rclass, resource_repository.DockerPackageType),
	)

	var packDockerMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.DockerFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.Ignore("class", "rclass", "member", "terraform_type"),
		),
		packDockerMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.DockerFederatedRepositoryParams{
			DockerLocalRepositoryParams: local.DockerLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: resource_repository.DockerPackageType,
					Rclass:      federated.Rclass,
				},
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      dockerFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: "Provides a data source for a federated docker V1 repository",
	}
}
