package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type HelmOciFederatedRepositoryParams struct {
	local.HelmOciLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedHelmOciRepository() *schema.Resource {
	packageType := "helmoci"

	helmociSchema := utilsdk.MergeMaps(
		local.HelmOciLocalSchema,
		federatedSchema,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := HelmOciFederatedRepositoryParams{
			HelmOciLocalRepositoryParams: local.UnpackLocalHelmOciRepository(data, rclass),
			Members:                      unpackMembers(data),
			RepoParams:                   unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*HelmOciFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packMembers,
	)

	constructor := func() (interface{}, error) {
		return &HelmOciFederatedRepositoryParams{
			HelmOciLocalRepositoryParams: local.HelmOciLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(helmociSchema, pkr, unpackRepository, constructor)
}
