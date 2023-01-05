package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryRemoteGenericRepository(pkt string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef("remote", pkt)()
		if err != nil {
			return nil, err
		}

		return &RepositoryRemoteBaseParams{
			PackageType:   pkt,
			Rclass:        "remote",
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRemoteRepo(data, pkt)
		return repo, repo.Id(), nil
	}

	mergedRemoteRepoSchema := util.MergeMaps(BaseRemoteRepoSchema, repository.RepoLayoutRefSchema("remote", pkt))

	return repository.MkResourceSchema(mergedRemoteRepoSchema, packer.Default(mergedRemoteRepoSchema), unpack, constructor)
}
