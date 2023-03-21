package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

func DataSourceArtifactoryFederatedDebianRepository() *schema.Resource {
	packageType := "debian"

	debianFederatedSchema := util.MergeMaps(
		local.DebianLocalSchema,
		memberSchema,
		resource_repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var packDebianMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.DebianFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packDebianMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.DebianFederatedRepositoryParams{
			DebianLocalRepositoryParams: local.DebianLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      debianFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: "Provides a data source for a federated debian repository",
	}
}
