package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func resourceArtifactoryRemoteGenericRepository(pkt string) *schema.Resource {
	constructor := func() interface{} {
		repoLayout, _ := getDefaultRepoLayoutRef("remote", pkt)()
		return &RemoteRepositoryBaseParams{
			PackageType:         pkt,
			Rclass:              "remote",
			RemoteRepoLayoutRef: repoLayout.(string),
		}
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := unpackBaseRemoteRepo(data, pkt)
		return repo, repo.Id(), nil
	}

	mergedRemoteRepoSchema := utils.MergeSchema(baseRemoteRepoSchema, repoLayoutRefSchema("remote", pkt))

	return mkResourceSchema(mergedRemoteRepoSchema, defaultPacker(mergedRemoteRepoSchema), unpack, constructor)
}
