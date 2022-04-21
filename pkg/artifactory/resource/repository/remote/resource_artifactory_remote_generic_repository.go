package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func ResourceArtifactoryRemoteGenericRepository(pkt string) *schema.Resource {
	constructor := func() interface{} {
		repoLayout, _ := utils.GetDefaultRepoLayoutRef("remote", pkt)()
		return &RemoteRepositoryBaseParams{
			PackageType:         pkt,
			Rclass:              "remote",
			RemoteRepoLayoutRef: repoLayout.(string),
		}
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRemoteRepo(data, pkt)
		return repo, repo.Id(), nil
	}

	mergedRemoteRepoSchema := utils.MergeSchema(BaseRemoteRepoSchema, repository.RepoLayoutRefSchema("remote", pkt))

	return repository.MkResourceSchema(mergedRemoteRepoSchema, repository.DefaultPacker(mergedRemoteRepoSchema), unpack, constructor)
}
