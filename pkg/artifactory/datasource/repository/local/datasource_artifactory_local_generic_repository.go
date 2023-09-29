package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalGenericRepository(repoType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.RepositoryBaseParams{
			PackageType: repoType,
			Rclass:      rclass,
		}, nil
	}

	genericRepoSchema := local.GetGenericRepoSchema(repoType)

	return &schema.Resource{
		Schema:      genericRepoSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(genericRepoSchema), constructor),
		Description: "Provides a data source for a local " + repoType + " repository",
	}
}
