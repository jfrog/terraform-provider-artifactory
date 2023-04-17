package virtual

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DatasourceArtifactoryVirtualDebianRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, virtual.DebianPackageType)()
		if err != nil {
			return nil, err
		}

		return &virtual.RepositoryBaseParams{
			PackageType:   virtual.DebianPackageType,
			Rclass:        rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	debianSchema := virtual.DebianVirtualSchema

	return &schema.Resource{
		Schema:      debianSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(debianSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a virtual %s repository", virtual.DebianPackageType),
	}
}
