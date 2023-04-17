package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func DataSourceArtifactoryVirtualGenericRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, packageType)()
		if err != nil {
			return nil, err
		}

		return &virtual.RepositoryBaseParams{
			PackageType:   packageType,
			Rclass:        rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	genericSchema := virtual.BaseVirtualRepoSchema

	return &schema.Resource{
		Schema:      genericSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(genericSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", packageType),
	}
}

func DataSourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, packageType)()
		if err != nil {
			return nil, err
		}
		return &virtual.RepositoryBaseParamsWithRetrievalCachePeriodSecs{
			RepositoryBaseParams: virtual.RepositoryBaseParams{
				Rclass:        rclass,
				PackageType:   packageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	var repoWithRetrievalCachePeriodSecsVirtualSchema = util.MergeMaps(
		virtual.BaseVirtualRepoSchema,
		virtual.RetrievalCachePeriodSecondsSchema,
	)

	return &schema.Resource{
		Schema:      repoWithRetrievalCachePeriodSecsVirtualSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(repoWithRetrievalCachePeriodSecsVirtualSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", packageType),
	}
}
