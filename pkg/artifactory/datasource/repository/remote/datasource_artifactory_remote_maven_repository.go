package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteMavenRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.MavenPackageType)
		if err != nil {
			return nil, err
		}

		return &remote.MavenRemoteRepo{
			JavaRemoteRepo: remote.JavaRemoteRepo{
				RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
					Rclass:        remote.Rclass,
					PackageType:   resource_repository.MavenPackageType,
					RepoLayoutRef: repoLayout,
				},
				SuppressPomConsistencyChecks: false,
			},
		}, nil
	}

	mavenSchema := remote.MavenSchemas[remote.MavenCurrentSchemaVersion]
	mavenSchema["url"].Required = false
	mavenSchema["url"].Optional = true

	return &schema.Resource{
		Schema:      mavenSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(mavenSchema), constructor),
		Description: "Provides a data source for a remote Maven repository",
	}
}
