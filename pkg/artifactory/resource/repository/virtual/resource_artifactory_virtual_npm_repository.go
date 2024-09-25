package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var npmSchema = lo.Assign(
	RetrievalCachePeriodSecondsSchema,
	externalDependenciesSchema,
	repository.RepoLayoutRefSchema(Rclass, repository.NPMPackageType),
)

var NPMSchemas = GetSchemas(npmSchema)

type NpmVirtualRepositoryParams struct {
	ExternalDependenciesVirtualRepositoryParams
	VirtualRetrievalCachePeriodSecs int `hcl:"retrieval_cache_period_seconds" json:"virtualRetrievalCachePeriodSecs"`
}

func ResourceArtifactoryVirtualNpmRepository() *schema.Resource {
	var unpackNpmVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := NpmVirtualRepositoryParams{
			VirtualRetrievalCachePeriodSecs:             d.GetInt("retrieval_cache_period_seconds", false),
			ExternalDependenciesVirtualRepositoryParams: unpackExternalDependenciesVirtualRepository(s, repository.NPMPackageType),
		}
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &NpmVirtualRepositoryParams{
			ExternalDependenciesVirtualRepositoryParams: ExternalDependenciesVirtualRepositoryParams{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: repository.NPMPackageType,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(
		NPMSchemas,
		packer.Default(NPMSchemas[CurrentSchemaVersion]),
		unpackNpmVirtualRepository,
		constructor,
	)
}
