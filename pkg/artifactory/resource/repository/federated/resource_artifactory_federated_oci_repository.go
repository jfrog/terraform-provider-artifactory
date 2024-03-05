package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type OciFederatedRepositoryParams struct {
	local.OciLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedOciRepository() *schema.Resource {
	packageType := "oci"

	ociFederatedSchema := utilsdk.MergeMaps(
		local.OciLocalSchema,
		federatedSchema,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackFederatedOciRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := OciFederatedRepositoryParams{
			OciLocalRepositoryParams: local.UnpackLocalOciRepository(data, rclass),
			Members:                  unpackMembers(data),
			RepoParams:               unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packOciMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*OciFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type", "docker_api_version"),
			),
		),
		packOciMembers,
	)

	constructor := func() (interface{}, error) {
		return &OciFederatedRepositoryParams{
			OciLocalRepositoryParams: local.OciLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(ociFederatedSchema, pkr, unpackFederatedOciRepository, constructor)
}
