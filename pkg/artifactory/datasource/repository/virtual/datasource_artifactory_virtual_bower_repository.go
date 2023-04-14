package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DatasourceArtifactoryVirtualBowerRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, virtual.BowerPackageType)()
		if err != nil {
			return nil, err
		}

		return &virtual.RepositoryBaseParams{
			PackageType:   virtual.BowerPackageType,
			Rclass:        rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	bowerSchema := virtual.BowerVirtualSchema

	return &schema.Resource{
		Schema:      bowerSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(bowerSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", virtual.BowerPackageType),
	}
}
