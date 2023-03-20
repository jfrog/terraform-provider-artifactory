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

func DataSourceArtifactoryFederatedDockerV2Repository() *schema.Resource {
	packageType := "docker"

	dockerV2FederatedSchema := util.MergeMaps(
		local.DockerV2LocalSchema,
		memberSchema,
		resource_repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var packDockerMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.DockerFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.All(
				predicate.NoClass,
				predicate.Ignore("member", "terraform_type"),
			),
		),
		packDockerMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.DockerFederatedRepositoryParams{
			DockerLocalRepositoryParams: local.DockerLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      dockerV2FederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: fmt.Sprintf("Provides a data source for a federated docker V2 repository"),
	}
}

func DataSourceArtifactoryFederatedDockerV1Repository() *schema.Resource {
	packageType := "docker"

	dockerFederatedSchema := util.MergeMaps(
		local.DockerV1LocalSchema,
		memberSchema,
		resource_repository.RepoLayoutRefSchema(rclass, packageType),
	)

	var packDockerMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*federated.DockerFederatedRepositoryParams).Members
		return federated.PackMembers(members, d)
	}

	pkr := packer.Compose(
		packer.Universal(
			predicate.Ignore("class", "rclass", "member", "terraform_type"),
		),
		packDockerMembers,
	)

	constructor := func() (interface{}, error) {
		return &federated.DockerFederatedRepositoryParams{
			DockerLocalRepositoryParams: local.DockerLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: packageType,
					Rclass:      rclass,
				},
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      dockerFederatedSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
		Description: "Provides a data source for a federated docker V1 repository",
	}
}
