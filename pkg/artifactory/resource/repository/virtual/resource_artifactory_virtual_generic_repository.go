package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
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

	genericSchema := utils.MergeSchema(BaseVirtualRepoSchema, repository.RepoLayoutRefSchema("virtual", pkt))

	return repository.MkResourceSchema(genericSchema, repository.DefaultPacker(genericSchema), unpack, constructor)
}

func ResourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(pkt string) *schema.Resource {
	var repoWithRetrivalCachePeriodSecsVirtualSchema = utils.MergeSchema(BaseVirtualRepoSchema, map[string]*schema.Schema{
		"retrieval_cache_period_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      7200,
			Description:  "This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.",
			ValidateFunc: validation.IntAtLeast(0),
		},
	}, repository.RepoLayoutRefSchema("virtual", pkt))
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
	return repository.MkResourceSchema(repoWithRetrivalCachePeriodSecsVirtualSchema, repository.DefaultPacker(repoWithRetrivalCachePeriodSecsVirtualSchema), unpack, constructor)
}
