package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteBowerRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.BowerPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.BowerRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   remote.BowerPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	bowerSchema := remote.BowerRemoteSchema(false)

	return &schema.Resource{
		Schema:      bowerSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(bowerSchema), constructor),
		Description: "Provides a data source for a remote Bower repository",
	}
}
