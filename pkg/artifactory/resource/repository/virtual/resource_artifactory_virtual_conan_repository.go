package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var conanSchema = lo.Assign(
	RetrievalCachePeriodSecondsSchema,
	repository.ConanBaseSchemaSDKv2,
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.ConanPackageType),
)

var ConanSchemas = GetSchemas(conanSchema)

type ConanRepoParams struct {
	RepositoryBaseParamsWithRetrievalCachePeriodSecs
	repository.ConanBaseParams
}

func ResourceArtifactoryVirtualConanRepository() *schema.Resource {
	var unpackConanRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := ConanRepoParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(s, repository.ConanPackageType),
			ConanBaseParams: repository.ConanBaseParams{
				EnableConanSupport:       true,
				ForceConanAuthentication: d.GetBool("force_conan_authentication", false),
			},
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &ConanRepoParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: repository.ConanPackageType,
				},
			},
			ConanBaseParams: repository.ConanBaseParams{
				EnableConanSupport: true,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		ConanSchemas,
		packer.Default(ConanSchemas[CurrentSchemaVersion]),
		unpackConanRepository,
		constructor,
	)
}
