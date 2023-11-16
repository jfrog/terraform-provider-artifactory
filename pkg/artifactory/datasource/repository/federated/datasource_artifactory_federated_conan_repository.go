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

func DataSourceArtifactoryFederatedConanRepository() *schema.Resource {
	conanSchema := utilsdk.MergeMaps(
		local.ConanSchema,
		federatedSchema,
		resource_repository.RepoLayoutRefSchema(rclass, resource_repository.ConanPackageType),
	)

	var packConanMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.ConanRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type", "enable_conan_support"),
			),
		),
		packConanMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.ConanRepositoryParams{
			ConanRepoParams: local.ConanRepoParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: resource_repository.ConanPackageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      conanSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: "Provides a data source for a federated Conan repository",
	}
}
