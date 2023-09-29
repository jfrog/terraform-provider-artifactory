package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

var ConanSchema = utilsdk.MergeMaps(
	BaseLocalRepoSchema,
	repository.ConanBaseSchema,
	repository.RepoLayoutRefSchema(rclass, repository.ConanPackageType),
)

type ConanRepoParams struct {
	RepositoryBaseParams
	repository.ConanBaseParams
}

func UnpackConnanRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &utilsdk.ResourceData{ResourceData: data}
	repo := ConanRepoParams{
		RepositoryBaseParams: UnpackBaseRepo(rclass, data, repository.ConanPackageType),
		ConanBaseParams: repository.ConanBaseParams{
			EnableConanSupport:       true,
			ForceConanAuthentication: d.GetBool("force_conan_authentication", false),
		},
	}
	return repo, repo.Id(), nil
}

func ResourceArtifactoryLocalConanRepository() *schema.Resource {

	constructor := func() (interface{}, error) {
		return &ConanRepoParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: repository.ConanPackageType,
				Rclass:      rclass,
			},
			ConanBaseParams: repository.ConanBaseParams{
				EnableConanSupport: true,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		ConanSchema,
		packer.Default(ConanSchema),
		UnpackConnanRepository,
		constructor,
	)
}
