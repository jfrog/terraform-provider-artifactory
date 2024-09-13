package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const ansiblePackageType = "ansible"

var AnsibleLocalSchema = utilsdk.MergeMaps(
	BaseLocalRepoSchema,
	repository.RepoLayoutRefSchema(rclass, ansiblePackageType),
	repository.AlpinePrimaryKeyPairRef,
)

type AnsibleLocalRepoParams struct {
	RepositoryBaseParams
	repository.PrimaryKeyPairRefParam
}

func UnpackLocalAnsibleRepository(data *schema.ResourceData, rclass string) AnsibleLocalRepoParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return AnsibleLocalRepoParams{
		RepositoryBaseParams: UnpackBaseRepo(rclass, data, ansiblePackageType),
		PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
			PrimaryKeyPairRef: d.GetString("primary_keypair_ref", false),
		},
	}
}

func ResourceArtifactoryLocalAnsibleRepository() *schema.Resource {
	var unpackLocalAnsibleRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalAnsibleRepository(data, rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &AnsibleLocalRepoParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: alpinePackageType,
				Rclass:      rclass,
			},
		}, nil
	}

	return repository.MkResourceSchema(AnsibleLocalSchema, packer.Default(AnsibleLocalSchema), unpackLocalAnsibleRepo, constructor)
}
