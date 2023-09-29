package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalCargoRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.CargoLocalRepoParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: "cargo",
				Rclass:      rclass,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      local.CargoLocalSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.CargoLocalSchema), constructor),
		Description: "Data source for local cargo repository",
	}
}
