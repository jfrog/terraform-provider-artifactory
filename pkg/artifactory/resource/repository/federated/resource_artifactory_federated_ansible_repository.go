package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type AnsibleRepositoryParams struct {
	local.RepositoryBaseParams
	RepoParams
	Members []Member `hcl:"member" json:"members"`
	repository.PrimaryKeyPairRefParam
}

func ResourceArtifactoryFederatedAnsibleRepository() *schema.Resource {
	packageType := "ansible"

	var ansibleSchema = utilsdk.MergeMaps(
		local.GetGenericRepoSchema(packageType),
		federatedSchemaV4,
		repository.AlpinePrimaryKeyPairRef,
		repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var unpackFederatedRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}

		repo := AnsibleRepositoryParams{
			RepositoryBaseParams: local.UnpackBaseRepo(rclass, data, packageType),
			RepoParams:           unpackRepoParams(data),
			Members:              unpackMembers(data),
			PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
				PrimaryKeyPairRef: d.GetString("primary_keypair_ref", false),
			},
		}
		return repo, repo.Id(), nil
	}

	var packMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*AnsibleRepositoryParams).Members
		return PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packMembers,
	)

	constructor := func() (interface{}, error) {
		return &AnsibleRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: local.GetPackageType(packageType),
				Rclass:      rclass,
			},
		}, nil
	}

	return mkResourceSchema(ansibleSchema, pkr, unpackFederatedRepository, constructor)
}
