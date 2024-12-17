package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteCoapodsRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.CocoapodsPackageType)
		if err != nil {
			return nil, err
		}

		return &remote.CocoapodsRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.CocoapodsPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	cocoapodsSchema := getSchema(remote.CocoapodsSchemas)

	return &schema.Resource{
		Schema:      cocoapodsSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(cocoapodsSchema), constructor),
		Description: "Provides a data source for a remote CocoaPods repository",
	}
}
