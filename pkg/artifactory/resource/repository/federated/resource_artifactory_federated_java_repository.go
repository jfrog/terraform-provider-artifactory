package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type JavaFederatedRepositoryParams struct {
	local.JavaLocalRepositoryParams
	Members []Member `hcl:"member" json:"members"`
	RepoParams
}

var unpackLocalJavaRepository = func(data *schema.ResourceData, Rclass string, packageType string) local.JavaLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return local.JavaLocalRepositoryParams{
		RepositoryBaseParams:         local.UnpackBaseRepo(Rclass, data, packageType),
		ChecksumPolicyType:           d.GetString("checksum_policy_type", false),
		SnapshotVersionBehavior:      d.GetString("snapshot_version_behavior", false),
		MaxUniqueSnapshots:           d.GetInt("max_unique_snapshots", false),
		HandleReleases:               d.GetBool("handle_releases", false),
		HandleSnapshots:              d.GetBool("handle_snapshots", false),
		SuppressPomConsistencyChecks: d.GetBool("suppress_pom_consistency_checks", false),
	}
}

func ResourceArtifactoryFederatedJavaRepository(packageType string, suppressPom bool) *schema.Resource {

	javaFederatedSchema := lo.Assign(
		local.GetJavaSchemas(packageType, suppressPom)[local.CurrentSchemaVersion],
		federatedSchemaV4,
		repository.RepoLayoutRefSDKv2Schema("federated", packageType),
	)

	var unpackFederatedJavaRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := JavaFederatedRepositoryParams{
			JavaLocalRepositoryParams: unpackLocalJavaRepository(data, Rclass, packageType),
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
					PackageType: packageType,
					Rclass:      Rclass,
				},
				SuppressPomConsistencyChecks: suppressPom,
			},
		}, nil
	}

	return mkResourceSchema(javaFederatedSchema, pkr, unpackFederatedJavaRepository, constructor)
}
