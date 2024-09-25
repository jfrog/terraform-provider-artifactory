package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var alpineSchema = lo.Assign(
	repository.RepoLayoutRefSchema(Rclass, repository.AlpinePackageType),
	repository.AlpinePrimaryKeyPairRef,
	repository.CompressionFormats,
)

var AlpineLocalSchemas = GetSchemas(alpineSchema)

type AlpineLocalRepoParams struct {
	RepositoryBaseParams
	repository.PrimaryKeyPairRefParam
}

func UnpackLocalAlpineRepository(data *schema.ResourceData, Rclass string) AlpineLocalRepoParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return AlpineLocalRepoParams{
		RepositoryBaseParams: UnpackBaseRepo(Rclass, data, repository.AlpinePackageType),
		PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
			PrimaryKeyPairRef: d.GetString("primary_keypair_ref", false),
		},
	}
}

func ResourceArtifactoryLocalAlpineRepository() *schema.Resource {
	var unpackLocalAlpineRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalAlpineRepository(data, Rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &AlpineLocalRepoParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: repository.AlpinePackageType,
				Rclass:      Rclass,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		AlpineLocalSchemas,
		packer.Default(AlpineLocalSchemas[CurrentSchemaVersion]),
		unpackLocalAlpineRepo,
		constructor,
	)
}
