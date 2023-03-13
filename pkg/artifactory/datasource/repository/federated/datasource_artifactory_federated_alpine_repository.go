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

func DataSourceArtifactoryFederatedAlpineRepository() *schema.Resource {
	packageType := "alpine"

	type AlpineFederatedRepositoryParams struct {
		local.AlpineLocalRepoParams
		Members []federated.Member `hcl:"member" json:"members"`
	}

	alpineFederatedSchema := util.MergeMaps(
		local.AlpineLocalSchema,
		MemberSchema,
		resource_repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var packAlpineMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*AlpineFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packAlpineMembers,
	)

	constructor := func() (interface{}, error) {
		return &AlpineFederatedRepositoryParams{
			AlpineLocalRepoParams: local.AlpineLocalRepoParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      alpineFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: fmt.Sprintf("Provides a data source for a federated alpine repository"),
	}
}
