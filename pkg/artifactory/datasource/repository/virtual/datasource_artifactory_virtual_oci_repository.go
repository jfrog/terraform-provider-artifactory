package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DatasourceArtifactoryVirtualOciRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(virtual.Rclass, resource_repository.OCIPackageType)()
		if err != nil {
			return nil, err
		}

		return &virtual.RepositoryBaseParams{
			PackageType:   resource_repository.OCIPackageType,
			Rclass:        virtual.Rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	ociSchema := virtual.OCISchemas[virtual.CurrentSchemaVersion]

	return &schema.Resource{
		Schema:      ociSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(ociSchema), constructor),
		Description: "Provides a data source for a virtual OCI repository",
	}
}
