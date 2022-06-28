package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func getTerraformLocalSchema(registryType string) map[string]*schema.Schema {
	return util.MergeMaps(
		BaseLocalRepoSchema,
		repository.RepoLayoutRefSchema("local", "terraform_"+registryType),
	)
}

func ResourceArtifactoryLocalTerraformRepository(registryType string) *schema.Resource {

	var unPackLocalTerraformRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRepo("local", data, "terraform_"+registryType)
		repo.TerraformType = registryType

		return repo, repo.Id(), nil
	}

	terraformLocalSchema := getTerraformLocalSchema(registryType)

	return repository.MkResourceSchema(
		terraformLocalSchema,
		packer.Default(terraformLocalSchema),
		unPackLocalTerraformRepository,
		func() interface{} {
			return &RepositoryBaseParams{
				PackageType: "terraform_" + registryType,
				Rclass:      "local",
			}
		},
	)
}
