package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalOciRepository() *schema.Resource {
	pkr := packer.Default(local.OciLocalSchema)

	constructor := func() (interface{}, error) {
		return &local.OciLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: local.OciPackageType,
				Rclass:      rclass,
			},
			DockerApiVersion: "V2",
			TagRetention:     1,
			MaxUniqueTags:    0, // no limit
		}, nil
	}

	return &schema.Resource{
		Schema:      local.OciLocalSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
	}
}
