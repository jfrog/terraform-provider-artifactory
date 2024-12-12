package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/samber/lo"
)

type CargoFederatedRepositoryParams struct {
	local.CargoLocalRepoParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedCargoRepository() *schema.Resource {
	cargoFederatedSchema := lo.Assign(
		local.CargoSchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, repository.CargoPackageType),
	)

	var unpackFederatedCargoRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := CargoFederatedRepositoryParams{
			CargoLocalRepoParams: local.UnpackLocalCargoRepository(data, Rclass),
			Members:              unpackMembers(data),
			RepoParams:           unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packCargoMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*CargoFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packCargoMembers,
	)

	constructor := func() (interface{}, error) {
		return &CargoFederatedRepositoryParams{
			CargoLocalRepoParams: local.CargoLocalRepoParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: repository.CargoPackageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(cargoFederatedSchema, pkr, unpackFederatedCargoRepository, constructor)
}
