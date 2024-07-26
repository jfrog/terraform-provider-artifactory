package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const AlpinePackageType = "alpine"

var AlpineVirtualSchema = utilsdk.MergeMaps(
	BaseVirtualRepoSchema,
	RetrievalCachePeriodSecondsSchema,
	repository.PrimaryKeyPairRef,
	repository.RepoLayoutRefSchema(Rclass, AlpinePackageType))

func ResourceArtifactoryVirtualAlpineRepository() *schema.Resource {
	type AlpineVirtualRepositoryParams struct {
		RepositoryBaseParamsWithRetrievalCachePeriodSecs
		repository.PrimaryKeyPairRefParam
	}

	var unpackAlpineVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := AlpineVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(s, AlpinePackageType),
			PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
				PrimaryKeyPairRef: d.GetString("primary_keypair_ref", false),
			},
		}
		repo.PackageType = AlpinePackageType
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &AlpineVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: AlpinePackageType,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(
		AlpineVirtualSchema,
		packer.Default(AlpineVirtualSchema),
		unpackAlpineVirtualRepository,
		constructor,
	)
}
