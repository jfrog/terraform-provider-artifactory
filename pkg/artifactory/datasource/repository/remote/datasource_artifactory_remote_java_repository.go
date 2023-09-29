package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteJavaRepository(packageType string, suppressPom bool) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, packageType)()
		if err != nil {
			return nil, err
		}

		return &remote.JavaRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   packageType,
				RepoLayoutRef: repoLayout.(string),
			},
			SuppressPomConsistencyChecks: suppressPom,
		}, nil
	}

	javaSchema := remote.JavaRemoteSchema(false, packageType, suppressPom)

	return &schema.Resource{
		Schema:      javaSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(javaSchema), constructor),
		Description: "Data source for a local Java repository of type: " + packageType,
	}
}
