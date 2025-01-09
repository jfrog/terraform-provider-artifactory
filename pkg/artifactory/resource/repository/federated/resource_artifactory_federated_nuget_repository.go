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

type NugetFederatedRepositoryParams struct {
	local.NugetLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func unpackLocalNugetRepository(data *schema.ResourceData, Rclass string) local.NugetLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return local.NugetLocalRepositoryParams{
		RepositoryBaseParams:     local.UnpackBaseRepo(Rclass, data, repository.NugetPackageType),
		MaxUniqueSnapshots:       d.GetInt("max_unique_snapshots", false),
		ForceNugetAuthentication: d.GetBool("force_nuget_authentication", false),
	}
}

func ResourceArtifactoryFederatedNugetRepository() *schema.Resource {
	nugetFederatedSchema := lo.Assign(
		local.NugetSchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, repository.NugetPackageType),
	)

	var unpackFederatedNugetRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := NugetFederatedRepositoryParams{
			NugetLocalRepositoryParams: unpackLocalNugetRepository(data, Rclass),
			Members:                    unpackMembers(data),
			RepoParams:                 unpackRepoParams(data),
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
					PackageType: repository.NugetPackageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(nugetFederatedSchema, pkr, unpackFederatedNugetRepository, constructor)
}
