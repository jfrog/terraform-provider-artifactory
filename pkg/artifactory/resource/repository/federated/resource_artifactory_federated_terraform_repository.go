package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/samber/lo"
)

type TerraformFederatedRepositoryParams struct {
	local.RepositoryBaseParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func unpackLocalTerraformRepository(data *schema.ResourceData, Rclass string, registryType string) local.RepositoryBaseParams {
	repo := local.UnpackBaseRepo(Rclass, data, "terraform_"+registryType)
	repo.TerraformType = registryType

	return repo
}

func ResourceArtifactoryFederatedTerraformRepository(registryType string) *schema.Resource {
	packageType := "terraform_" + registryType

	terraformFederatedSchema := lo.Assign(
		local.GetTerraformSchemas(registryType)[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
	)

	var unpackFederatedTerraformRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := TerraformFederatedRepositoryParams{
			RepositoryBaseParams: unpackLocalTerraformRepository(data, Rclass, registryType),
			Members:              unpackMembers(data),
			RepoParams:           unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packTerraformMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*TerraformFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packTerraformMembers,
	)

	constructor := func() (interface{}, error) {
		return &TerraformFederatedRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: packageType,
				Rclass:      Rclass,
			},
		}, nil
	}

	return mkResourceSchema(terraformFederatedSchema, pkr, unpackFederatedTerraformRepository, constructor)
}
