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

type OciFederatedRepositoryParams struct {
	local.OciLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func unpackLocalOciRepository(data *schema.ResourceData, Rclass string) local.OciLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return local.OciLocalRepositoryParams{
		RepositoryBaseParams: local.UnpackBaseRepo(Rclass, data, repository.OCIPackageType),
		MaxUniqueTags:        d.GetInt("max_unique_tags", false),
		DockerApiVersion:     "V2",
		TagRetention:         d.GetInt("tag_retention", false),
	}
}

func ResourceArtifactoryFederatedOciRepository() *schema.Resource {
	ociFederatedSchema := lo.Assign(
		local.OCILocalSchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, repository.OCIPackageType),
	)

	var unpackFederatedOciRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := OciFederatedRepositoryParams{
			OciLocalRepositoryParams: unpackLocalOciRepository(data, Rclass),
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
					PackageType: repository.OCIPackageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(ociFederatedSchema, pkr, unpackFederatedOciRepository, constructor)
}
