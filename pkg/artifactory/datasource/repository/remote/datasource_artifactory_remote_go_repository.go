package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteGoRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.GoPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.GoRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   remote.GoPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	goSchema := remote.GoRemoteSchema(false)

	return &schema.Resource{
		Schema:      goSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(goSchema), constructor),
		Description: "Provides a data source for a remote Go repository",
	}
}
