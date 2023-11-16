package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type ConanRepositoryParams struct {
	local.ConanRepoParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedConanRepository() *schema.Resource {
	conanSchema := utilsdk.MergeMaps(
		local.ConanSchema,
		federatedSchema,
		repository.RepoLayoutRefSchema(rclass, repository.ConanPackageType),
	)

	var unpackConanRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}
		repo := ConanRepositoryParams{
			ConanRepoParams: local.ConanRepoParams{
				RepositoryBaseParams: local.UnpackBaseRepo(rclass, data, repository.ConanPackageType),
				ConanBaseParams: repository.ConanBaseParams{
					EnableConanSupport:       true,
					ForceConanAuthentication: d.GetBool("force_conan_authentication", false),
				},
			},
			Members:    unpackMembers(data),
			RepoParams: unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packConanMembers = func(repo interface{}, d *schema.ResourceData) error {
		repo.(*ConanRepositoryParams).EnableConanSupport = true
		members := repo.(*ConanRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type", "enable_conan_support"),
			),
		),
		packConanMembers,
	)

	constructor := func() (interface{}, error) {
		return &ConanRepositoryParams{
			ConanRepoParams: local.ConanRepoParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: repository.ConanPackageType,
					Rclass:      rclass,
				},
				ConanBaseParams: repository.ConanBaseParams{
					EnableConanSupport: true,
				},
			},
		}, nil
	}

	return mkResourceSchema(conanSchema, pkr, unpackConanRepository, constructor)
}
