package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalCargoRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.CargoLocalRepoParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: resource_repository.CargoPackageType,
				Rclass:      local.Rclass,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      local.CargoSchemas[local.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.CargoSchemas[local.CurrentSchemaVersion]), constructor),
		Description: "Data source for local cargo repository",
	}
}
