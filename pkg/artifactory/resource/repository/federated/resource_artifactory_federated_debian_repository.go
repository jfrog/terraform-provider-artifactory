package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type DebianFederatedRepositoryParams struct {
	local.DebianLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func unpackLocalDebianRepository(data *schema.ResourceData, Rclass string) local.DebianLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return local.DebianLocalRepositoryParams{
		PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
			PrimaryKeyPairRefSDKv2: d.GetString("primary_keypair_ref", false),
		},
		SecondaryKeyPairRefParam: repository.SecondaryKeyPairRefParam{
			SecondaryKeyPairRefSDKv2: d.GetString("secondary_keypair_ref", false),
		},
		RepositoryBaseParams:    local.UnpackBaseRepo(Rclass, data, repository.DebianPackageType),
		TrivialLayout:           d.GetBool("trivial_layout", false),
		IndexCompressionFormats: d.GetSet("index_compression_formats"),
	}
}

func ResourceArtifactoryFederatedDebianRepository() *schema.Resource {
	debianFederatedSchema := utilsdk.MergeMaps(
		local.DebianSchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema(Rclass, repository.DebianPackageType),
	)

	var unpackFederatedDebianRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := DebianFederatedRepositoryParams{
			DebianLocalRepositoryParams: unpackLocalDebianRepository(data, Rclass),
			Members:                     unpackMembers(data),
			RepoParams:                  unpackRepoParams(data),
		}
		return repo, repo.Id(), nil
	}

	var packDebianMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*DebianFederatedRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packDebianMembers,
	)

	constructor := func() (interface{}, error) {
		return &DebianFederatedRepositoryParams{
			DebianLocalRepositoryParams: local.DebianLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: repository.DebianPackageType,
					Rclass:      Rclass,
				},
			},
		}, nil
	}

	return mkResourceSchema(debianFederatedSchema, pkr, unpackFederatedDebianRepository, constructor)
}
