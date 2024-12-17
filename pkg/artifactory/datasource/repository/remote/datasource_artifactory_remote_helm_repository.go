package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteHelmRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.HelmPackageType)
		if err != nil {
			return nil, err
		}

		return &remote.HelmRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.HelmPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	helmSchema := getSchema(remote.HelmSchemas)

	return &schema.Resource{
		Schema:      helmSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(helmSchema), constructor),
		Description: "Provides a data source for a remote Helm repository",
	}
}
