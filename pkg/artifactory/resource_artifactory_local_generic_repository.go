package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func resourceArtifactoryLocalGenericRepository(pkt string) *schema.Resource {
	constructor := func() interface{} {
		return &LocalRepositoryBaseParams{
			PackageType: pkt,
			Rclass:      "local",
		}
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := unpackBaseRepo("local", data, pkt)
		return repo, repo.Id(), nil
	}
	mergedLocalRepoSchema := utils.MergeSchema(baseLocalRepoSchema, repoLayoutRefSchema("local", pkt))
	return mkResourceSchema(mergedLocalRepoSchema, defaultPacker(mergedLocalRepoSchema), unpack, constructor)
}
