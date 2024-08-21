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

func DataSourceArtifactoryFederatedAnsibleRepository() *schema.Resource {
	packageType := "ansible"

	ansibleFederatedSchema := utilsdk.MergeMaps(
		local.AnsibleLocalSchema,
		federatedSchemaV3,
		resource_repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var packMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.AnsibleRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.AnsibleRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: packageType,
				Rclass:      rclass,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      ansibleFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: "Provides a data source for a federated Ansible repository",
	}
}
