package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

type CargoFederatedRepositoryParams struct {
	local.CargoLocalRepoParams
	Members []Member `hcl:"member" json:"members"`
}

func ResourceArtifactoryFederatedCargoRepository() *schema.Resource {
	packageType := "cargo"

	cargoFederatedSchema := util.MergeMaps(
		local.CargoLocalSchema,
		memberSchema,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackFederatedCargoRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := CargoFederatedRepositoryParams{
			CargoLocalRepoParams: local.UnpackLocalCargoRepository(data, rclass),
			Members:              unpackMembers(data),
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
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(cargoFederatedSchema, pkr, unpackFederatedCargoRepository, constructor)
}
