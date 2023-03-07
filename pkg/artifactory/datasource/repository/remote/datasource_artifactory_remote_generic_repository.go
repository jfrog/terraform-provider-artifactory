package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteGenericRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.GenericPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.GenericRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   remote.GenericPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	genericSchema := remote.GenericRemoteSchema(false)

	return &schema.Resource{
		Schema:      genericSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(genericSchema), constructor),
		Description: "Provides a data source for a remote Generic repository",
	}
}
