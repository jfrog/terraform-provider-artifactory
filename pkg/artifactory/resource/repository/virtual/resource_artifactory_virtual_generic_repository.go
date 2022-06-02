package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryVirtualGenericRepository(pkt string) *schema.Resource {
	constructor := func() interface{} {
		return &VirtualRepositoryBaseParams{
			PackageType: pkt,
			Rclass:      "virtual",
		}
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseVirtRepo(data, pkt)
		return repo, repo.Id(), nil
	}

	genericSchema := util.MergeSchema(BaseVirtualRepoSchema, repository.RepoLayoutRefSchema("virtual", pkt))

	return repository.MkResourceSchema(genericSchema, repository.DefaultPacker(genericSchema), unpack, constructor)
}

func ResourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(pkt string) *schema.Resource {
	var repoWithRetrivalCachePeriodSecsVirtualSchema = util.MergeSchema(
		BaseVirtualRepoSchema,
		retrievalCachePeriodSecondsSchema,
		repository.RepoLayoutRefSchema("virtual", pkt),
	)

	constructor := func() interface{} {
		return &VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs{
			VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: pkt,
			},
		}
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(data, pkt)
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(
		repoWithRetrivalCachePeriodSecsVirtualSchema,
		repository.DefaultPacker(repoWithRetrivalCachePeriodSecsVirtualSchema),
		unpack,
		constructor,
	)
}
