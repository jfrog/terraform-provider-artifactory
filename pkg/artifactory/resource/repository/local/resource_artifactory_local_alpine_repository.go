package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const alpinePackageType = "alpine"

var AlpineLocalSchema = utilsdk.MergeMaps(
	BaseLocalRepoSchema,
	repository.RepoLayoutRefSchema(rclass, alpinePackageType),
	repository.AlpinePrimaryKeyPairRef,
	repository.CompressionFormats,
)

type AlpineLocalRepoParams struct {
	RepositoryBaseParams
	repository.PrimaryKeyPairRefParam
}

func UnpackLocalAlpineRepository(data *schema.ResourceData, rclass string) AlpineLocalRepoParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return AlpineLocalRepoParams{
		RepositoryBaseParams: UnpackBaseRepo(rclass, data, alpinePackageType),
		PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
			PrimaryKeyPairRef: d.GetString("primary_keypair_ref", false),
		},
	}
}

func ResourceArtifactoryLocalAlpineRepository() *schema.Resource {
	var unpackLocalAlpineRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalAlpineRepository(data, rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &AlpineLocalRepoParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: alpinePackageType,
				Rclass:      rclass,
			},
		}, nil
	}

	return repository.MkResourceSchema(AlpineLocalSchema, packer.Default(AlpineLocalSchema), unpackLocalAlpineRepo, constructor)
}
