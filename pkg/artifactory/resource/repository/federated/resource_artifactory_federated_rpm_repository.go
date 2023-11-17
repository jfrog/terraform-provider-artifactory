package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type RpmFederatedRepositoryParams struct {
	local.RpmLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedRpmRepository() *schema.Resource {
	packageType := "rpm"

	rpmFederatedSchema := utilsdk.MergeMaps(
		local.RpmLocalSchema,
		federatedSchema,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackFederatedRpmRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := RpmFederatedRepositoryParams{
			RpmLocalRepositoryParams: local.UnpackLocalRpmRepository(data, rclass),
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
					PackageType: packageType,
					Rclass:      rclass,
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
