package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/samber/lo"
)

type RpmFederatedRepositoryParams struct {
	local.RpmLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedRpmRepository() *schema.Resource {
	rpmFederatedSchema := lo.Assign(
		local.RPMSchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSchema(Rclass, repository.RPMPackageType),
	)

	var unpackFederatedRpmRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := RpmFederatedRepositoryParams{
			RpmLocalRepositoryParams: local.UnpackLocalRpmRepository(data, Rclass),
			Members:                  unpackMembers(data),
			RepoParams:               unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packRpmMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*RpmFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packRpmMembers,
	)

	constructor := func() (interface{}, error) {
		return &RpmFederatedRepositoryParams{
			RpmLocalRepositoryParams: local.RpmLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: repository.RPMPackageType,
					Rclass:      Rclass,
				},
				RootDepth:               0,
				CalculateYumMetadata:    false,
				EnableFileListsIndexing: false,
				GroupFileNames:          "",
			},
		}, nil
	}

	return mkResourceSchema(rpmFederatedSchema, pkr, unpackFederatedRpmRepository, constructor)
}
