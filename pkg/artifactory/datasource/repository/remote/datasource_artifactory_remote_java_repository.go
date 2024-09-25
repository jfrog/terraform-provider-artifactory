package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteJavaRepository(packageType string, suppressPom bool) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, packageType)()
		if err != nil {
			return nil, err
		}

		return &remote.JavaRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   packageType,
				RepoLayoutRef: repoLayout.(string),
			},
			SuppressPomConsistencyChecks: suppressPom,
		}, nil
	}

	javaSchema := getSchema(remote.GetSchemas(remote.JavaSchema(packageType, suppressPom)))

	return &schema.Resource{
		Schema:      javaSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(javaSchema), constructor),
		Description: "Data source for a local Java repository of type: " + packageType,
	}
}
