package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

func DataSourceArtifactoryVirtualJavaRepository(packageType string) *schema.Resource {
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
	var mavenSchema = lo.Assign(
		virtual.JavaSchema,
		resource_repository.RepoLayoutRefSDKv2Schema(virtual.Rclass, packageType),
	)

	var mavenSchemas = virtual.GetSchemas(mavenSchema)

	return &schema.Resource{
		Schema:      mavenSchemas[virtual.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(mavenSchemas[virtual.CurrentSchemaVersion]), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", packageType),
	}
}
