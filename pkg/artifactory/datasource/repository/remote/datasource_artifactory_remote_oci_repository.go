package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteOciRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.OCIPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.OciRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.OCIPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	ociSchema := getSchema(remote.OCISchemas)

	return &schema.Resource{
		Schema:      ociSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(ociSchema), constructor),
		Description: "Provides a data source for a remote OCI repository",
	}
}
