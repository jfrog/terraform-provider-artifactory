package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalDebianRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.DebianLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: resource_repository.DebianPackageType,
				Rclass:      local.Rclass,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      local.DebianSchemas[local.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.DebianSchemas[local.CurrentSchemaVersion]), constructor),
		Description: "Data source for local debian repository",
	}
}
