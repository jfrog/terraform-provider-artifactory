package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalGenericRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.RepositoryBaseParams{
			PackageType: packageType,
			Rclass:      local.Rclass,
		}, nil
	}

	genericRepoSchemas := local.GetGenericSchemas(packageType)

	return &schema.Resource{
		Schema:      genericRepoSchemas[local.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(genericRepoSchemas[local.CurrentSchemaVersion]), constructor),
		Description: "Provides a data source for a local " + packageType + " repository",
	}
}
