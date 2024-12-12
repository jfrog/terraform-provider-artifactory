package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteGoRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.GoPackageType)
		if err != nil {
			return nil, err
		}

		return &remote.GoRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.GoPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	goSchema := getSchema(remote.GoSchemas)

	return &schema.Resource{
		Schema:      goSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(goSchema), constructor),
		Description: "Provides a data source for a remote Go repository",
	}
}
