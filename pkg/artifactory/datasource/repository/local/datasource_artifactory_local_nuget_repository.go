package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalNugetRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.NugetLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: resource_repository.NugetPackageType,
				Rclass:      local.Rclass,
			},
			MaxUniqueSnapshots:       0,
			ForceNugetAuthentication: false,
		}, nil
	}

	return &schema.Resource{
		Schema:      local.NugetSchemas[local.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.NugetSchemas[local.CurrentSchemaVersion]), constructor),
		Description: "Data Source for a local nuget repository",
	}
}
