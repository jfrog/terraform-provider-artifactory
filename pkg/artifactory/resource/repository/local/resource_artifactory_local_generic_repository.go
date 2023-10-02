package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func GetGenericRepoSchema(repoType string) map[string]*schema.Schema {
	return utilsdk.MergeMaps(BaseLocalRepoSchema, repository.RepoLayoutRefSchema(rclass, repoType))
}

func ResourceArtifactoryLocalGenericRepository(repoType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		return &RepositoryBaseParams{
			PackageType: repoType,
			Rclass:      rclass,
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRepo(rclass, data, repoType)
		return repo, repo.Id(), nil
	}

	genericRepoSchema := GetGenericRepoSchema(repoType)

	return repository.MkResourceSchema(genericRepoSchema, packer.Default(genericRepoSchema), unpack, constructor)
}
