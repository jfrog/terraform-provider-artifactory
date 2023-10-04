package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalConanRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.ConanRepoParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: resource_repository.ConanPackageType,
				Rclass:      rclass,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      local.ConanSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.ConanSchema), constructor),
		Description: "Data source for local Conan repository",
	}
}
