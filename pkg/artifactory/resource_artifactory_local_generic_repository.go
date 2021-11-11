package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArtifactoryLocalGenericRepository(pkt string) *schema.Resource {
	return mkResourceSchema(baseLocalRepoSchema, universalPack, func(data *schema.ResourceData) (interface{}, string, error) {
		repo := unpackBaseLocalRepo(data, pkt)
		return repo, repo.Id(), nil
	}, func() interface{} {
		return &LocalRepositoryBaseParams{
			PackageType: pkt,
			Rclass:      "local",
		}
	})
}
