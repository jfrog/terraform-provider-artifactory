package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryFederatedDebianRepository() *schema.Resource {
	packageType := "debian"

	type DebianFederatedRepositoryParams struct {
		local.DebianLocalRepositoryParams
		Members []Member `hcl:"member" json:"members"`
	}

	debianFederatedSchema := util.MergeMaps(
		local.DebianLocalSchema,
		memberSchema,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackFederatedDebianRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := DebianFederatedRepositoryParams{
			DebianLocalRepositoryParams: local.UnpackLocalDebianRepository(data, rclass),
			Members:                     unpackMembers(data),
		}
		return repo, repo.Id(), nil
	}

	var packDebianMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*DebianFederatedRepositoryParams).Members
		return packMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.Ignore("class", "rclass", "member", "terraform_type"),
		),
		packDebianMembers,
	)

	constructor := func() (interface{}, error) {
		return &DebianFederatedRepositoryParams{
			DebianLocalRepositoryParams: local.DebianLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(debianFederatedSchema, pkr, unpackFederatedDebianRepository, constructor)
}
