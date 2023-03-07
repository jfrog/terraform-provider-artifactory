package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalRpmRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.RpmLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: "rpm",
				Rclass:      rclass,
			},
			RootDepth:               0,
			CalculateYumMetadata:    false,
			EnableFileListsIndexing: false,
			GroupFileNames:          "",
		}, nil
	}

	return &schema.Resource{
		Schema:      local.RpmLocalSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.RpmLocalSchema), constructor),
		Description: "Data source for a local rpm repository",
	}
}
