package federated

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/federated"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func DataSourceArtifactoryFederatedJavaRepository(packageType string, suppressPom bool) *schema.Resource {

	javaFederatedSchema := utilsdk.MergeMaps(
		local.GetJavaRepoSchema(packageType, suppressPom),
		federatedSchema,
		resource_repository.RepoLayoutRefSchema("federated", packageType),
	)

	var packJavaMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.JavaFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packJavaMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.JavaFederatedRepositoryParams{
			JavaLocalRepositoryParams: local.JavaLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
				SuppressPomConsistencyChecks: suppressPom,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      javaFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: fmt.Sprintf("Provides a data source for a federated %s repository", packageType),
	}
}
