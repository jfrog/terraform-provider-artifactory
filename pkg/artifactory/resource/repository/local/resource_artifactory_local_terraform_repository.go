package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func GetTerraformLocalSchema(registryType string) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseLocalRepoSchema,
		repository.RepoLayoutRefSchema(rclass, "terraform_"+registryType),
	)
}

func UnpackLocalTerraformRepository(data *schema.ResourceData, rclass string, registryType string) RepositoryBaseParams {
	repo := UnpackBaseRepo(rclass, data, "terraform_"+registryType)
	repo.TerraformType = registryType

	return repo
}

func ResourceArtifactoryLocalTerraformRepository(registryType string) *schema.Resource {

	var unpackLocalTerraformRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalTerraformRepository(data, rclass, registryType)
		return repo, repo.Id(), nil
	}

	terraformLocalSchema := GetTerraformLocalSchema(registryType)

	constructor := func() (interface{}, error) {
		return &RepositoryBaseParams{
			PackageType: "terraform_" + registryType,
			Rclass:      rclass,
		}, nil
	}

	return repository.MkResourceSchema(
		terraformLocalSchema,
		packer.Default(terraformLocalSchema),
		unpackLocalTerraformRepository,
		constructor,
	)
}
