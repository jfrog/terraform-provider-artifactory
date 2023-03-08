package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemotecoDockerRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.DockerPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.DockerRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   remote.DockerPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	dockerSchema := remote.DockerRemoteSchema(false)

	return &schema.Resource{
		Schema:      dockerSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(dockerSchema), constructor),
		Description: "Provides a data source for a remote Docker repository",
	}
}
