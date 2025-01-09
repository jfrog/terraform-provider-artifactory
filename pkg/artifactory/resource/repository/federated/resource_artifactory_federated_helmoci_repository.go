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

type HelmOciFederatedRepositoryParams struct {
	local.HelmOciLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func unpackLocalHelmOciRepository(data *schema.ResourceData, Rclass string) local.HelmOciLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return local.HelmOciLocalRepositoryParams{
		RepositoryBaseParams: local.UnpackBaseRepo(Rclass, data, repository.HelmOCIPackageType),
		MaxUniqueTags:        d.GetInt("max_unique_tags", false),
		TagRetention:         d.GetInt("tag_retention", false),
	}
}

func ResourceArtifactoryFederatedHelmOciRepository() *schema.Resource {
	helmociSchema := lo.Assign(
		local.HelmOCISchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, repository.HelmOCIPackageType),
	)

	var unpackRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := HelmOciFederatedRepositoryParams{
			HelmOciLocalRepositoryParams: unpackLocalHelmOciRepository(data, Rclass),
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
					PackageType: repository.HelmOCIPackageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(helmociSchema, pkr, unpackRepository, constructor)
}
