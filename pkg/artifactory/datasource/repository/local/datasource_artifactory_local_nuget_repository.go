package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalNugetRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.NugetLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: "nuget",
				Rclass:      rclass,
			},
			MaxUniqueSnapshots:       0,
			ForceNugetAuthentication: false,
		}, nil
	}

	return &schema.Resource{
		Schema:      local.NugetLocalSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.NugetLocalSchema), constructor),
		Description: "Data Source for a local nuget repository",
	}
}
