package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

func GetGenericSchemas(packageType string) map[int16]map[string]*schema.Schema {
	return map[int16]map[string]*schema.Schema{
		0: lo.Assign(
			BaseSchemaV1,
			repository.RepoLayoutRefSchema(Rclass, packageType),
		),
		1: lo.Assign(
			BaseSchemaV1,
			repository.RepoLayoutRefSchema(Rclass, packageType),
		),
	}
}

func ResourceArtifactoryLocalGenericRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		return &RepositoryBaseParams{
			PackageType: packageType,
			Rclass:      Rclass,
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRepo(Rclass, data, packageType)
		return repo, repo.Id(), nil
	}

	genericRepoSchemas := GetGenericSchemas(packageType)

	return repository.MkResourceSchema(
		genericRepoSchemas,
		packer.Default(genericRepoSchemas[CurrentSchemaVersion]),
		unpack,
		constructor,
	)
}
