package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func DataSourceArtifactoryFederatedHelmOciRepository() *schema.Resource {
	ociFederatedSchema := utilsdk.MergeMaps(
		local.HelmOciLocalSchema,
		federatedSchemaV4,
		resource_repository.RepoLayoutRefSchema(rclass, local.HelmOciPackageType),
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
					PackageType: local.HelmOciPackageType,
					Rclass:      rclass,
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
