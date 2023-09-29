package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DatasourceArtifactoryVirtualNpmRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, virtual.NpmPackageType)()
		if err != nil {
			return nil, err
		}

		return &virtual.RepositoryBaseParams{
			PackageType:   virtual.NpmPackageType,
			Rclass:        rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	npmSchema := virtual.NpmVirtualSchema

	return &schema.Resource{
		Schema:      npmSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(npmSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", virtual.NpmPackageType),
	}
}
