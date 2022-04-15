package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func resourceArtifactoryVirtualGenericRepository(pkt string) *schema.Resource {
	constructor := func() interface{} {
		return &VirtualRepositoryBaseParams{
			PackageType: pkt,
			Rclass:      "virtual",
		}
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := unpackBaseVirtRepo(data, pkt)
		return repo, repo.Id(), nil
	}

	genericSchema := utils.MergeSchema(baseVirtualRepoSchema, repoLayoutRefSchema("virtual", pkt))

	return mkResourceSchema(genericSchema, defaultPacker(genericSchema), unpack, constructor)
}

func resourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(pkt string) *schema.Resource {
	var repoWithRetrivalCachePeriodSecsVirtualSchema = utils.MergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{
		"retrieval_cache_period_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      7200,
			Description:  "This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.",
			ValidateFunc: validation.IntAtLeast(0),
		},
	}, repoLayoutRefSchema("virtual", pkt))
	constructor := func() interface{} {
		return &VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs{
			VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: pkt,
			},
		}
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := unpackBaseVirtRepoWithRetrievalCachePeriodSecs(data, pkt)
		return repo, repo.Id(), nil
	}
	return mkResourceSchema(repoWithRetrivalCachePeriodSecsVirtualSchema, defaultPacker(repoWithRetrivalCachePeriodSecsVirtualSchema), unpack, constructor)
}
