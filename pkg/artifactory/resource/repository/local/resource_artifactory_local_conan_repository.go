package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var conanSchema = lo.Assign(
	repository.ConanBaseSchema,
	repository.RepoLayoutRefSchema(Rclass, repository.ConanPackageType),
)

var ConanSchemas = GetSchemas(conanSchema)

type ConanRepoParams struct {
	RepositoryBaseParams
	repository.ConanBaseParams
}

func UnpackConnanRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &utilsdk.ResourceData{ResourceData: data}
	repo := ConanRepoParams{
		RepositoryBaseParams: UnpackBaseRepo(Rclass, data, repository.ConanPackageType),
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
				Rclass:      Rclass,
			},
			ConanBaseParams: repository.ConanBaseParams{
				EnableConanSupport: true,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		ConanSchemas,
		packer.Default(ConanSchemas[CurrentSchemaVersion]),
		UnpackConnanRepository,
		constructor,
	)
}
