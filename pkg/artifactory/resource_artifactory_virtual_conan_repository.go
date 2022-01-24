package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type ConanVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	VirtualRetrievalCachePeriodSecs int `hcl:"virtual_retrieval_cache_period_seconds" json:"virtualRetrievalCachePeriodSecs,omitempty"`
}

func resourceArtifactoryConanVirtualRepository() *schema.Resource {
	var conanVirtualSchema = mergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{
		"virtual_retrieval_cache_period_seconds": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.",
			DefaultFunc: func() (interface{}, error) {
				return 7200, nil
			},
			ValidateFunc: validation.IntAtLeast(0),
		},
	})

	return mkResourceSchema(conanVirtualSchema, defaultPacker, unpackConanVirtualRepository, func() interface{} {
		return &ConanVirtualRepositoryParams{
			VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: "conan",
			},
		}
	})

}

func unpackConanVirtualRepository(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	repo := ConanVirtualRepositoryParams{
		VirtualRepositoryBaseParams:     unpackBaseVirtRepo(s),
		VirtualRetrievalCachePeriodSecs: d.getInt("virtual_retrieval_cache_period_seconds", true),
	}
	repo.PackageType = "conan"
	return &repo, repo.Key, nil
}
