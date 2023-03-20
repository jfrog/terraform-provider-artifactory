package federated

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

func DataSourceArtifactoryFederatedGenericRepository(packageType string) *schema.Resource {
	localRepoSchema := local.GetGenericRepoSchema(packageType)

	var federatedSchema = util.MergeMaps(localRepoSchema, memberSchema, resource_repository.RepoLayoutRefSchema(rclass, packageType))

	var packGenericMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.FederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packGenericMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.FederatedRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: local.GetPackageType(packageType),
				Rclass:      rclass,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      federatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: fmt.Sprintf("Provides a data source for a federated %s repository", packageType),
	}
}
