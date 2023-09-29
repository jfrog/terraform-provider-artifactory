package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const NpmPackageType = "npm"

var NpmVirtualSchema = utilsdk.MergeMaps(
	BaseVirtualRepoSchema,
	RetrievalCachePeriodSecondsSchema,
	externalDependenciesSchema,
	repository.RepoLayoutRefSchema(Rclass, NpmPackageType),
)

func ResourceArtifactoryVirtualNpmRepository() *schema.Resource {

	type NpmVirtualRepositoryParams struct {
		ExternalDependenciesVirtualRepositoryParams
		VirtualRetrievalCachePeriodSecs int `hcl:"retrieval_cache_period_seconds" json:"virtualRetrievalCachePeriodSecs"`
	}

	var unpackNpmVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := NpmVirtualRepositoryParams{
			VirtualRetrievalCachePeriodSecs:             d.GetInt("retrieval_cache_period_seconds", false),
			ExternalDependenciesVirtualRepositoryParams: unpackExternalDependenciesVirtualRepository(s, NpmPackageType),
		}
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &NpmVirtualRepositoryParams{
			ExternalDependenciesVirtualRepositoryParams: ExternalDependenciesVirtualRepositoryParams{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: NpmPackageType,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(
		NpmVirtualSchema,
		packer.Default(NpmVirtualSchema),
		unpackNpmVirtualRepository,
		constructor,
	)
}
