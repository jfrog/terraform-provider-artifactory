package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteNugetRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.NugetPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.NugetRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   remote.NugetPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	nugetSchema := remote.NugetRemoteSchema(false)

	return &schema.Resource{
		Schema:      nugetSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(nugetSchema), constructor),
		Description: "Provides a data source for a remote NuGet repository",
	}
}
