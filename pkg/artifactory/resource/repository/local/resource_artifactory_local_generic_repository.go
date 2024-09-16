package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func GetGenericRepoSchema(packageType string) map[string]*schema.Schema {
	return utilsdk.MergeMaps(BaseLocalRepoSchema, repository.RepoLayoutRefSchema(rclass, packageType))
}

func ResourceArtifactoryLocalGenericRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		return &RepositoryBaseParams{
			PackageType: packageType,
			Rclass:      rclass,
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRepo(rclass, data, packageType)
		return repo, repo.Id(), nil
	}

	genericRepoSchema := GetGenericRepoSchema(packageType)

	return repository.MkResourceSchema(genericRepoSchema, packer.Default(genericRepoSchema), unpack, constructor)
}
