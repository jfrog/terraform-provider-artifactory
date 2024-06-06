package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalHelmOciRepository() *schema.Resource {
	pkr := packer.Default(local.OciLocalSchema)

	constructor := func() (interface{}, error) {
		return &local.HelmOciLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: local.HelmOciPackageType,
				Rclass:      rclass,
			},
			TagRetention:  1,
			MaxUniqueTags: 0, // no limit
		}, nil
	}

	return &schema.Resource{
		Schema:      local.HelmOciLocalSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
	}
}
