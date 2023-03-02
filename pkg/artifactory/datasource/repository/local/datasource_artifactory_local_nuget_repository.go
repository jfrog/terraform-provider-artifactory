package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalNugetRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.NugetLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: "nuget",
				Rclass:      "local",
			},
			MaxUniqueSnapshots:       0,
			ForceNugetAuthentication: false,
		}, nil
	}

	return &schema.Resource{
		Schema:      local.NugetLocalSchema,
		ReadContext: MkRepoReadDataSource(packer.Default(local.NugetLocalSchema), constructor),
		Description: "Data Source for a local nuget repository",
	}
}
