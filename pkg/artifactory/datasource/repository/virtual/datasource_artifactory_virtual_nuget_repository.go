package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DatasourceArtifactoryVirtualNugetRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, virtual.NugetPackageType)()
		if err != nil {
			return nil, err
		}

		return &virtual.RepositoryBaseParams{
			PackageType:   virtual.NugetPackageType,
			Rclass:        rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	nugetSchema := virtual.NugetVirtualSchema

	return &schema.Resource{
		Schema:      nugetSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(nugetSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", virtual.NugetPackageType),
	}
}
