package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteBowerRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.BowerPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.BowerRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.BowerPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	bowerSchema := getSchema(remote.BowerSchemas)

	return &schema.Resource{
		Schema:      bowerSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(bowerSchema), constructor),
		Description: "Provides a data source for a remote Bower repository",
	}
}
