package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteHelmRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.HelmPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.HelmRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   remote.HelmPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	helmSchema := remote.HelmRemoteSchema(false)

	return &schema.Resource{
		Schema:      helmSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(helmSchema), constructor),
		Description: "Provides a data source for a remote Helm repository",
	}
}
