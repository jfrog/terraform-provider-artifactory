package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/samber/lo"
)

func DataSourceArtifactoryFederatedNugetRepository() *schema.Resource {
	nugetFederatedSchema := lo.Assign(
		local.NugetSchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		resource_repository.RepoLayoutRefSDKv2Schema(federated.Rclass, resource_repository.NugetPackageType),
	)

	var packNugetMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.NugetFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packNugetMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.NugetFederatedRepositoryParams{
			NugetLocalRepositoryParams: local.NugetLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: resource_repository.NugetPackageType,
					Rclass:      federated.Rclass,
				},
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      nugetFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: "Provides a data source for a federated nuget repository",
	}
}
