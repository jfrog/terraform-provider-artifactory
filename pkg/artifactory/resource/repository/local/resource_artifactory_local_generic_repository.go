package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func GetGenericRepoSchema(repoType string) map[string]*schema.Schema {
	return util.MergeMaps(BaseLocalRepoSchema, repository.RepoLayoutRefSchema("local", repoType))
}

func ResourceArtifactoryLocalGenericRepository(repoType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		return &RepositoryBaseParams{
			PackageType: repoType,
			Rclass:      "local",
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRepo("local", data, repoType)
		return repo, repo.Id(), nil
	}

	genericRepoSchema := GetGenericRepoSchema(repoType)

	return repository.MkResourceSchema(genericRepoSchema, packer.Default(genericRepoSchema), unpack, constructor)
}
