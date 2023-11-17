package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func DataSourceArtifactoryFederatedCargoRepository() *schema.Resource {
	packageType := "cargo"

	cargoFederatedSchema := utilsdk.MergeMaps(
		local.CargoLocalSchema,
		federatedSchema,
		resource_repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var packCargoMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.CargoFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packCargoMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.CargoFederatedRepositoryParams{
			CargoLocalRepoParams: local.CargoLocalRepoParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      cargoFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: "Provides a data source for a federated cargo repository",
	}
}
