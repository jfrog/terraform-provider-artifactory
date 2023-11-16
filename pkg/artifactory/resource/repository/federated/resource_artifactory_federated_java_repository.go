package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type JavaFederatedRepositoryParams struct {
	local.JavaLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

func ResourceArtifactoryFederatedJavaRepository(repoType string, suppressPom bool) *schema.Resource {

	javaFederatedSchema := utilsdk.MergeMaps(
		local.GetJavaRepoSchema(repoType, suppressPom),
		federatedSchema,
		repository.RepoLayoutRefSchema("federated", repoType),
	)

	var unpackFederatedJavaRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := JavaFederatedRepositoryParams{
			JavaLocalRepositoryParams: local.UnpackLocalJavaRepository(data, rclass, repoType),
			Members:                   unpackMembers(data),
			RepoParams:                unpackRepoParams(data),
		}

		return repo, repo.Id(), nil
	}

	var packJavaMembers = func(repo interface{}, d *schema.ResourceData) error {
		members := repo.(*JavaFederatedRepositoryParams).Members
		return PackMembers(members, d)
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
		return &JavaFederatedRepositoryParams{
			JavaLocalRepositoryParams: local.JavaLocalRepositoryParams{
				RepositoryBaseParams: local.RepositoryBaseParams{
					PackageType: repoType,
					Rclass:      rclass,
				},
				SuppressPomConsistencyChecks: suppressPom,
			},
		}, nil
	}

	return mkResourceSchema(javaFederatedSchema, pkr, unpackFederatedJavaRepository, constructor)
}
