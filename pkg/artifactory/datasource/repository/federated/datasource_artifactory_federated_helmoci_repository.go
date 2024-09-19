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

func DataSourceArtifactoryFederatedHelmOciRepository() *schema.Resource {
	ociFederatedSchema := lo.Assign(
		local.HelmOCISchemas[local.CurrentSchemaVersion],
		federatedSchemaV4,
		resource_repository.RepoLayoutRefSchema(federated.Rclass, resource_repository.HelmOCIPackageType),
	)

	var packOciMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.HelmOciFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type", "docker_api_version"),
			),
		),
		packOciMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.HelmOciFederatedRepositoryParams{
			HelmOciLocalRepositoryParams: local.HelmOciLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: resource_repository.HelmOCIPackageType,
					Rclass:      federated.Rclass,
				},
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      ociFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: "Provides a data source for a federated Helm OCI repository",
	}
}
