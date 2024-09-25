package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalHelmOciRepository() *schema.Resource {
	pkr := packer.Default(local.HelmOCISchemas[local.CurrentSchemaVersion])

	constructor := func() (interface{}, error) {
		return &local.HelmOciLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: resource_repository.HelmOCIPackageType,
				Rclass:      local.Rclass,
			},
			TagRetention:  1,
			MaxUniqueTags: 0, // no limit
		}, nil
	}

	return &schema.Resource{
		Schema:      local.HelmOCISchemas[local.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
	}
}
