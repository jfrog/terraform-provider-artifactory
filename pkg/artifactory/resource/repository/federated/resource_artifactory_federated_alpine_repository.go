package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

type AlpineFederatedRepositoryParams struct {
	local.AlpineLocalRepoParams
	Members []Member `hcl:"member" json:"members"`
}

func ResourceArtifactoryFederatedAlpineRepository() *schema.Resource {
	packageType := "alpine"

	alpineFederatedSchema := util.MergeMaps(
		local.AlpineLocalSchema,
		memberSchema,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackFederatedAlpineRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := AlpineFederatedRepositoryParams{
			AlpineLocalRepoParams: local.UnpackLocalAlpineRepository(data, rclass),
			Members:               unpackMembers(data),
		}
		return repo, repo.Id(), nil
	}

	var packAlpineMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*AlpineFederatedRepositoryParams).Members
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
		return &AlpineFederatedRepositoryParams{
			AlpineLocalRepoParams: local.AlpineLocalRepoParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(alpineFederatedSchema, pkr, unpackFederatedAlpineRepository, constructor)
}
