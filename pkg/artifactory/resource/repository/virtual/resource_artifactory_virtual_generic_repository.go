package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func ResourceArtifactoryVirtualGenericRepository(pkt string) *schema.Resource {
	constructor := func() (interface{}, error) {
		return &RepositoryBaseParams{
			PackageType: pkt,
			Rclass:      Rclass,
		}, nil
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseVirtRepo(data, pkt)
		return repo, repo.Id(), nil
	}

	genericSchema := utilsdk.MergeMaps(BaseVirtualRepoSchema,
		repository.RepoLayoutRefSchema(Rclass, pkt))

	return repository.MkResourceSchema(genericSchema, packer.Default(genericSchema), unpack, constructor)
}

func ResourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(pkt string) *schema.Resource {
	var repoWithRetrivalCachePeriodSecsVirtualSchema = utilsdk.MergeMaps(
		BaseVirtualRepoSchema,
		RetrievalCachePeriodSecondsSchema,
		repository.RepoLayoutRefSchema(Rclass, pkt),
	)

	constructor := func() (interface{}, error) {
		return &RepositoryBaseParamsWithRetrievalCachePeriodSecs{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: pkt,
			},
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(data, pkt)
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(
		repoWithRetrivalCachePeriodSecsVirtualSchema,
		packer.Default(repoWithRetrivalCachePeriodSecsVirtualSchema),
		unpack,
		constructor,
	)
}
