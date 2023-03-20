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

func DataSourceArtifactoryFederatedRpmRepository() *schema.Resource {
	packageType := "rpm"

	rpmFederatedSchema := util.MergeMaps(
		local.RpmLocalSchema,
		memberSchema,
		resource_repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var packRpmMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.RpmFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packRpmMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.RpmFederatedRepositoryParams{
			RpmLocalRepositoryParams: local.RpmLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
				RootDepth:               0,
				CalculateYumMetadata:    false,
				EnableFileListsIndexing: false,
				GroupFileNames:          "",
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      rpmFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: "Provides a data source for a federated RPM repository",
	}
}
