package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type TerraformFederatedRepositoryParams struct {
	local.RepositoryBaseParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedTerraformRepository(registryType string) *schema.Resource {
	packageType := "terraform_" + registryType

	terraformFederatedSchema := utilsdk.MergeMaps(
		local.GetTerraformLocalSchema(registryType),
		federatedSchema,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackFederatedTerraformRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := TerraformFederatedRepositoryParams{
			RepositoryBaseParams: local.UnpackLocalTerraformRepository(data, rclass, registryType),
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
				Rclass:      rclass,
			},
		}, nil
	}

	return mkResourceSchema(terraformFederatedSchema, pkr, unpackFederatedTerraformRepository, constructor)
}
