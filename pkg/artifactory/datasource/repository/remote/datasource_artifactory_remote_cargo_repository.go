package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteCargoRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.CargoPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.CargoRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   remote.CargoPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	cargoSchema := remote.CargoRemoteSchema(false)

	return &schema.Resource{
		Schema:      cargoSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(cargoSchema), constructor),
		Description: "Provides a data source for a remote Cargo repository",
	}
}
