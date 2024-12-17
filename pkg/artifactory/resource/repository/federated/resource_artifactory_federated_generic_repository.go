package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/samber/lo"
)

type GenericRepositoryParams struct {
	local.RepositoryBaseParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedGenericRepository(packageType string) *schema.Resource {
	var genericSchema = lo.Assign(
		local.GetGenericSchemas(packageType)[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
	)

	var unpackFederatedRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := GenericRepositoryParams{
			RepositoryBaseParams: local.UnpackBaseRepo(Rclass, data, packageType),
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
				PackageType: local.GetPackageType(packageType),
				Rclass:      Rclass,
			},
		}, nil
	}

	return mkResourceSchema(genericSchema, pkr, unpackFederatedRepository, constructor)
}
