package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteHelmOciRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.HelmOciPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.HelmOciRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   remote.HelmOciPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	ociSchema := remote.HelmOciRemoteSchema(false)

	return &schema.Resource{
		Schema:      ociSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(ociSchema), constructor),
		Description: "Provides a data source for a remote Helm OCI repository",
	}
}
