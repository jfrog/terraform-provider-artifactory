package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type ConanRepo struct {
	RepositoryRemoteBaseParams
	RepositoryCurationParams
	repository.ConanBaseParams
}

var conanSchema = lo.Assign(
	baseSchema,
	CurationRemoteRepoSchema,
	repository.ConanBaseSchema,
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.ConanPackageType),
)

var ConanSchemas = GetSchemas(conanSchema)

func ResourceArtifactoryRemoteConanRepository() *schema.Resource {
	var unpackConanRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := ConanRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.ConanPackageType),
			RepositoryCurationParams: RepositoryCurationParams{
				Curated: d.GetBool("curated", false),
			},
			ConanBaseParams: repository.ConanBaseParams{
				EnableConanSupport:       true,
				ForceConanAuthentication: d.GetBool("force_conan_authentication", false),
			},
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.ConanPackageType)
		if err != nil {
			return nil, err
		}

		return &ConanRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.ConanPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	return mkResourceSchema(
		ConanSchemas,
		packer.Default(ConanSchemas[CurrentSchemaVersion]),
		unpackConanRepo,
		constructor,
	)
}
