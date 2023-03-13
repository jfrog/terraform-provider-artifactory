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

type TerraformFederatedRepositoryParams struct {
	local.RepositoryBaseParams
	Members []federated.Member `hcl:"member" json:"members"`
}

func DataSourceArtifactoryFederatedTerraformRepository(registryType string) *schema.Resource {
	packageType := "terraform_" + registryType

	terraformFederatedSchema := util.MergeMaps(
		local.GetTerraformLocalSchema(registryType),
		MemberSchema,
		resource_repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var packTerraformMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*TerraformFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packTerraformMembers,
	)

	constructor := func() (interface{}, error) {
		return &TerraformFederatedRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: packageType,
				Rclass:      rclass,
			},
		}, nil
	}
	return &schema.Resource{
		Schema:      terraformFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: fmt.Sprintf("Provides a data source for a federated terraform-%s repository", registryType),
	}
}
