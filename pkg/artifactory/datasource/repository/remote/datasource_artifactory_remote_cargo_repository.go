package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteCargoRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.CargoPackageType)
		if err != nil {
			return nil, err
		}

		return &remote.CargoRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.CargoPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	cargoSchema := getSchema(remote.CargoSchemas)
	cargoSchema["git_registry_url"].Required = false
	cargoSchema["git_registry_url"].Optional = true

	return &schema.Resource{
		Schema:      cargoSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(cargoSchema), constructor),
		Description: "Provides a data source for a remote Cargo repository",
	}
}
