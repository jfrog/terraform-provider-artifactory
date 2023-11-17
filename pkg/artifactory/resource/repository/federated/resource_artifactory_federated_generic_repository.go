package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type GenericRepositoryParams struct {
	local.RepositoryBaseParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedGenericRepository(repoType string) *schema.Resource {
	var genericSchema = utilsdk.MergeMaps(
		local.GetGenericRepoSchema(repoType),
		federatedSchema,
		repository.RepoLayoutRefSchema(rclass, repoType),
	)

	var unpackFederatedRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := GenericRepositoryParams{
			RepositoryBaseParams: local.UnpackBaseRepo(rclass, data, repoType),
			Members:              unpackMembers(data),
			RepoParams:           unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packGenericMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*GenericRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packGenericMembers,
	)

	constructor := func() (interface{}, error) {
		return &GenericRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: local.GetPackageType(repoType),
				Rclass:      rclass,
			},
		}, nil
	}

	return mkResourceSchema(genericSchema, pkr, unpackFederatedRepository, constructor)
}
