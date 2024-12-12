package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryVirtualGenericRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(virtual.Rclass, packageType)
		if err != nil {
			return nil, err
		}

		return &virtual.RepositoryBaseParams{
			PackageType:   packageType,
			Rclass:        virtual.Rclass,
			RepoLayoutRef: repoLayout,
		}, nil
	}

	genericSchemas := virtual.GetSchemas(resource_repository.RepoLayoutRefSDKv2Schema(virtual.Rclass, packageType))

	return &schema.Resource{
		Schema:      genericSchemas[virtual.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(genericSchemas[virtual.CurrentSchemaVersion]), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", packageType),
	}
}

func DataSourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(virtual.Rclass, packageType)
		if err != nil {
			return nil, err
		}
		return &virtual.RepositoryBaseParamsWithRetrievalCachePeriodSecs{
			RepositoryBaseParams: virtual.RepositoryBaseParams{
				Rclass:        virtual.Rclass,
				PackageType:   packageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	repoWithRetrivalCachePeriodSecsVirtualSchemas := virtual.RepoWithRetrivalCachePeriodSecsVirtualSchemas(packageType)

	return &schema.Resource{
		Schema:      repoWithRetrivalCachePeriodSecsVirtualSchemas[virtual.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(repoWithRetrivalCachePeriodSecsVirtualSchemas[virtual.CurrentSchemaVersion]), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", packageType),
	}
}
