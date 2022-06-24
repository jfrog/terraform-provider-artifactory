package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryVirtualNpmRepository() *schema.Resource {

	const packageType = "npm"

	var npmVirtualSchema = util.MergeSchema(
		BaseVirtualRepoSchema,
		retrievalCachePeriodSecondsSchema,
		externalDependenciesSchema,
		repository.RepoLayoutRefSchema("virtual", packageType),
	)

	type NpmVirtualRepositoryParams struct {
		ExternalDependenciesVirtualRepositoryParams
		VirtualRetrievalCachePeriodSecs int `json:"virtualRetrievalCachePeriodSecs"`
	}

	var unpackNpmVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{s}

		repo := NpmVirtualRepositoryParams{
			VirtualRetrievalCachePeriodSecs:             d.GetInt("retrieval_cache_period_seconds", false),
			ExternalDependenciesVirtualRepositoryParams: unpackExternalDependenciesVirtualRepository(s, packageType),
		}
		return &repo, repo.Key, nil
	}

	return repository.MkResourceSchema(
		npmVirtualSchema,
		packer.Default(npmVirtualSchema),
		unpackNpmVirtualRepository,
		func() interface{} {
			return &NpmVirtualRepositoryParams{
				ExternalDependenciesVirtualRepositoryParams: ExternalDependenciesVirtualRepositoryParams{
					VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
						Rclass:      "virtual",
						PackageType: packageType,
					},
				},
			}
		},
	)
}
