package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

type NugetFederatedRepositoryParams struct {
	local.NugetLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
}

func ResourceArtifactoryFederatedNugetRepository() *schema.Resource {
	packageType := "nuget"

	nugetFederatedSchema := util.MergeMaps(
		local.NugetLocalSchema,
		memberSchema,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackFederatedNugetRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := NugetFederatedRepositoryParams{
			NugetLocalRepositoryParams: local.UnpackLocalNugetRepository(data, rclass),
			Members:                    unpackMembers(data),
		}
		return repo, repo.Id(), nil
	}

	var packNugetMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*NugetFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packNugetMembers,
	)

	constructor := func() (interface{}, error) {
		return &NugetFederatedRepositoryParams{
			NugetLocalRepositoryParams: local.NugetLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(nugetFederatedSchema, pkr, unpackFederatedNugetRepository, constructor)
}
