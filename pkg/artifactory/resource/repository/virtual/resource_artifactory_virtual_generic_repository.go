package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

func ResourceArtifactoryVirtualGenericRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		return &RepositoryBaseParams{
			PackageType: packageType,
			Rclass:      Rclass,
		}, nil
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseVirtRepo(data, packageType)
		return repo, repo.Id(), nil
	}

	genericSchemas := GetSchemas(repository.RepoLayoutRefSchema(Rclass, packageType))

	return repository.MkResourceSchema(
		genericSchemas,
		packer.Default(genericSchemas[CurrentSchemaVersion]),
		unpack,
		constructor,
	)
}

var RepoWithRetrivalCachePeriodSecsVirtualSchemas = func(packageType string) map[int16]map[string]*schema.Schema {
	var repoWithRetrivalCachePeriodSecsVirtualSchema = lo.Assign(
		RetrievalCachePeriodSecondsSchema,
		repository.RepoLayoutRefSchema(Rclass, packageType),
	)

	return GetSchemas(repoWithRetrivalCachePeriodSecsVirtualSchema)
}

func ResourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(packageType string) *schema.Resource {
	repoWithRetrivalCachePeriodSecsVirtualSchemas := RepoWithRetrivalCachePeriodSecsVirtualSchemas(packageType)

	constructor := func() (interface{}, error) {
		return &RepositoryBaseParamsWithRetrievalCachePeriodSecs{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: packageType,
			},
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(data, packageType)
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(
		repoWithRetrivalCachePeriodSecsVirtualSchemas,
		packer.Default(repoWithRetrivalCachePeriodSecsVirtualSchemas[CurrentSchemaVersion]),
		unpack,
		constructor,
	)
}
