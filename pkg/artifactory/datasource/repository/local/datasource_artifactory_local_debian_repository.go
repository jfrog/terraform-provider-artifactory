package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalDebianRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.DebianLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: "debian",
				Rclass:      rclass,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      local.DebianLocalSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.DebianLocalSchema), constructor),
		Description: "Data source for local debian repository",
	}
}
