package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalJavaRepository(packageType string, suppressPom bool) *schema.Resource {
	javaLocalSchemas := local.GetJavaSchemas(packageType, suppressPom)

	constructor := func() (interface{}, error) {
		return &local.JavaLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: packageType,
				Rclass:      local.Rclass,
			},
			SuppressPomConsistencyChecks: suppressPom,
		}, nil
	}

	return &schema.Resource{
		Schema:      javaLocalSchemas[local.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(javaLocalSchemas[local.CurrentSchemaVersion]), constructor),
		Description: "Data source for a local Java repository of type: " + packageType,
	}
}
