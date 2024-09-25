package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/samber/lo"
)

type DockerFederatedRepositoryParams struct {
	local.DockerLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedDockerV2Repository() *schema.Resource {
	dockerV2FederatedSchema := lo.Assign(
		local.DockerV2Schemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSchema(Rclass, repository.DockerPackageType),
	)

	var unpackFederatedDockerRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := DockerFederatedRepositoryParams{
			DockerLocalRepositoryParams: local.UnpackLocalDockerV2Repository(data, Rclass),
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

func ResourceArtifactoryFederatedDockerV1Repository() *schema.Resource {
	dockerFederatedSchema := lo.Assign(
		local.DockerV1Schemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSchema(Rclass, repository.DockerPackageType),
	)

	var unpackFederatedDockerRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := DockerFederatedRepositoryParams{
			DockerLocalRepositoryParams: local.UnpackLocalDockerV1Repository(data, Rclass),
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
