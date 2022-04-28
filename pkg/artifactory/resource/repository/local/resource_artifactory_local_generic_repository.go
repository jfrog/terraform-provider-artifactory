package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryLocalGenericRepository(pkt string) *schema.Resource {
	constructor := func() interface{} {
		return &LocalRepositoryBaseParams{
			PackageType: pkt,
			Rclass:      "local",
		}
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRepo("local", data, pkt)
		return repo, repo.Id(), nil
	}
	mergedLocalRepoSchema := util.MergeSchema(BaseLocalRepoSchema, repository.RepoLayoutRefSchema("local", pkt))
	return repository.MkResourceSchema(mergedLocalRepoSchema, repository.DefaultPacker(mergedLocalRepoSchema), unpack, constructor)
}
