package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	mergedRemoteRepoSchema := mergeSchema(baseRemoteRepoSchema, repoLayoutRefSchema("remote", pkt))

	genericRepoPacker := universalPack(
		allHclPredicate(
			noPassword,
			schemaHasKey(mergedRemoteRepoSchema),
		),
	)

	return mkResourceSchema(mergedRemoteRepoSchema, genericRepoPacker, unpack, constructor)
}
