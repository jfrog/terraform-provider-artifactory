package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryVirtualNpmRepository() *schema.Resource {

	const packageType = "npm"

	var npmVirtualSchema = util.MergeMaps(
		BaseVirtualRepoSchema,
		retrievalCachePeriodSecondsSchema,
		externalDependenciesSchema,
		repository.RepoLayoutRefSchema("virtual", packageType),
	)

	type NpmVirtualRepositoryParams struct {
		ExternalDependenciesVirtualRepositoryParams
		VirtualRetrievalCachePeriodSecs int `hcl:"retrieval_cache_period_seconds" json:"virtualRetrievalCachePeriodSecs,omitempty"`
	}

	var unpackNpmVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}

		repo := NpmVirtualRepositoryParams{
			VirtualRetrievalCachePeriodSecs:             d.GetInt("retrieval_cache_period_seconds", false),
			ExternalDependenciesVirtualRepositoryParams: unpackExternalDependenciesVirtualRepository(s, packageType),
		}
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &NpmVirtualRepositoryParams{
			ExternalDependenciesVirtualRepositoryParams: ExternalDependenciesVirtualRepositoryParams{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      "virtual",
					PackageType: packageType,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(
		npmVirtualSchema,
		packer.Default(npmVirtualSchema),
		unpackNpmVirtualRepository,
		constructor,
	)
}
