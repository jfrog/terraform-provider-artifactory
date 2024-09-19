package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DatasourceArtifactoryVirtualBowerRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(virtual.Rclass, resource_repository.BowerPackageType)()
		if err != nil {
			return nil, err
		}

		return &virtual.RepositoryBaseParams{
			PackageType:   resource_repository.BowerPackageType,
			Rclass:        virtual.Rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	bowerSchema := virtual.BowerSchemas[virtual.CurrentSchemaVersion]

	return &schema.Resource{
		Schema:      bowerSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(bowerSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", resource_repository.BowerPackageType),
	}
}
