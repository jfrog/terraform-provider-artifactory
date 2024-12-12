package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteVcsRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.VCSPackageType)
		if err != nil {
			return nil, err
		}

		return &remote.VcsRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.VCSPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	vcsSchema := getSchema(remote.VCSSchemas)

	return &schema.Resource{
		Schema:      vcsSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(vcsSchema), constructor),
		Description: "Provides a data source for a remote VCS repository",
	}
}
