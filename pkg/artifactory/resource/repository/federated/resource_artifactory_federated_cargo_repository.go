package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type CargoFederatedRepositoryParams struct {
	local.CargoLocalRepoParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedCargoRepository() *schema.Resource {
	packageType := "cargo"

	cargoFederatedSchema := utilsdk.MergeMaps(
		local.CargoLocalSchema,
		federatedSchema,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackFederatedCargoRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := CargoFederatedRepositoryParams{
			CargoLocalRepoParams: local.UnpackLocalCargoRepository(data, rclass),
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
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(cargoFederatedSchema, pkr, unpackFederatedCargoRepository, constructor)
}
