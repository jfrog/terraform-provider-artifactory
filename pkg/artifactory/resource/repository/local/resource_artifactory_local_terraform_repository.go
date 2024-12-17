package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

func GetTerraformSchemas(registryType string) map[int16]map[string]*schema.Schema {
	return map[int16]map[string]*schema.Schema{
		0: lo.Assign(
			BaseSchemaV1,
			repository.RepoLayoutRefSDKv2Schema(Rclass, "terraform_"+registryType),
		),
		1: lo.Assign(
			BaseSchemaV1,
			repository.RepoLayoutRefSDKv2Schema(Rclass, "terraform_"+registryType),
		),
	}
}

func UnpackLocalTerraformRepository(data *schema.ResourceData, Rclass string, registryType string) RepositoryBaseParams {
	repo := UnpackBaseRepo(Rclass, data, "terraform_"+registryType)
	repo.TerraformType = registryType

	return repo
}

func ResourceArtifactoryLocalTerraformRepository(registryType string) *schema.Resource {

	var unpackLocalTerraformRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalTerraformRepository(data, Rclass, registryType)
		return repo, repo.Id(), nil
	}

	terraformSchemas := GetTerraformSchemas(registryType)

	constructor := func() (interface{}, error) {
		return &RepositoryBaseParams{
			PackageType: "terraform_" + registryType,
			Rclass:      Rclass,
		}, nil
	}

	return repository.MkResourceSchema(
		terraformSchemas,
		packer.Default(terraformSchemas[CurrentSchemaVersion]),
		unpackLocalTerraformRepository,
		constructor,
	)
}
