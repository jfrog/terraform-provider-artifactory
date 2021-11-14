package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArtifactoryLocalGenericRepository(pkt string) *schema.Resource {
	constructor := func() interface{} {
		return &LocalRepositoryBaseParams{
			PackageType: pkt,
			Rclass:      "local",
		}
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := unpackBaseLocalRepo(data, pkt)
		return repo, repo.Id(), nil
	}
	return mkResourceSchema(baseLocalRepoSchema, universalPack(schemaHasKey(baseRemoteSchema)), unpack, constructor)
}
