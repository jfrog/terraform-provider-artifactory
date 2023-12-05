package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteMavenRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.MavenPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.MavenRemoteRepo{
			JavaRemoteRepo: remote.JavaRemoteRepo{
				RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
					Rclass:        rclass,
					PackageType:   remote.MavenPackageType,
					RepoLayoutRef: repoLayout.(string),
				},
				SuppressPomConsistencyChecks: false,
			},
		}, nil
	}

	mavenSchema := remote.MavenRemoteSchema(false)

	return &schema.Resource{
		Schema:      mavenSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(mavenSchema), constructor),
		Description: "Provides a data source for a remote Maven repository",
	}
}
