package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var ansibleSchema = lo.Assign(
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.AnsiblePackageType),
	repository.AlpinePrimaryKeyPairRefSDKv2,
)

var AnsibleSchemas = GetSchemas(ansibleSchema)

type AnsibleLocalRepoParams struct {
	RepositoryBaseParams
	repository.PrimaryKeyPairRefParam
}

func UnpackLocalAnsibleRepository(data *schema.ResourceData, Rclass string) AnsibleLocalRepoParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return AnsibleLocalRepoParams{
		RepositoryBaseParams: UnpackBaseRepo(Rclass, data, repository.AnsiblePackageType),
		PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
			PrimaryKeyPairRefSDKv2: d.GetString("primary_keypair_ref", false),
		},
	}
}

func ResourceArtifactoryLocalAnsibleRepository() *schema.Resource {
	var unpackLocalAnsibleRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalAnsibleRepository(data, Rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &AnsibleLocalRepoParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: repository.AnsiblePackageType,
				Rclass:      Rclass,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		AnsibleSchemas,
		packer.Default(AnsibleSchemas[CurrentSchemaVersion]),
		unpackLocalAnsibleRepo,
		constructor,
	)
}
