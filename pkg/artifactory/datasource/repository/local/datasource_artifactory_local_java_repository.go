package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalJavaRepository(repoType string, suppressPom bool) *schema.Resource {
	javaLocalSchema := local.GetJavaRepoSchema(repoType, suppressPom)

	constructor := func() (interface{}, error) {
		return &local.JavaLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: repoType,
				Rclass:      rclass,
			},
			SuppressPomConsistencyChecks: suppressPom,
		}, nil
	}

	return &schema.Resource{
		Schema:      javaLocalSchema,
		ReadContext: MkRepoReadDataSource(packer.Default(javaLocalSchema), constructor),
		Description: "Data source for a local java repository of type: " + repoType,
	}
}
