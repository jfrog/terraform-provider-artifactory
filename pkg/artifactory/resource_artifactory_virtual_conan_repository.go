package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	return mkResourceSchema(baseVirtualRepoSchema, defaultPacker, unpack, constructor)
}
