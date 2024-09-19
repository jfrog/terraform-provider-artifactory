package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type AlpineRepositoryParams struct {
	local.AlpineLocalRepoParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedAlpineRepository() *schema.Resource {
	alpineFederatedSchema := utilsdk.MergeMaps(
		local.AlpineLocalSchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSchema(Rclass, repository.AlpinePackageType),
	)

	var unpackFederatedAlpineRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := AlpineRepositoryParams{
			AlpineLocalRepoParams: local.UnpackLocalAlpineRepository(data, Rclass),
			Members:               unpackMembers(data),
			RepoParams:            unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packAlpineMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*AlpineRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packAlpineMembers,
	)

	constructor := func() (interface{}, error) {
		return &AlpineRepositoryParams{
			AlpineLocalRepoParams: local.AlpineLocalRepoParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: repository.AlpinePackageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(alpineFederatedSchema, pkr, unpackFederatedAlpineRepository, constructor)
}
