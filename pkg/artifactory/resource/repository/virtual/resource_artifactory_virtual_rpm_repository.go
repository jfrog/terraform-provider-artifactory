package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const RpmPackageType = "rpm"

var RpmVirtualSchema = utilsdk.MergeMaps(
	BaseVirtualRepoSchema,
	repository.PrimaryKeyPairRef,
	repository.SecondaryKeyPairRef,
	repository.RepoLayoutRefSchema(Rclass, RpmPackageType),
)

func ResourceArtifactoryVirtualRpmRepository() *schema.Resource {
	type CommonRpmDebianVirtualRepositoryParams struct {
		repository.PrimaryKeyPairRefParam
		repository.SecondaryKeyPairRefParam
	}

	type RpmVirtualRepositoryParams struct {
		RepositoryBaseParams
		CommonRpmDebianVirtualRepositoryParams
	}

	var unpackRpmVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := RpmVirtualRepositoryParams{
			RepositoryBaseParams: UnpackBaseVirtRepo(s, "rpm"),
			CommonRpmDebianVirtualRepositoryParams: CommonRpmDebianVirtualRepositoryParams{
				PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
					PrimaryKeyPairRef: d.GetString("primary_keypair_ref", false),
				},
				SecondaryKeyPairRefParam: repository.SecondaryKeyPairRefParam{
					SecondaryKeyPairRef: d.GetString("secondary_keypair_ref", false),
				},
			},
		}
		repo.PackageType = "rpm"

		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &RpmVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: RpmPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		RpmVirtualSchema,
		packer.Default(RpmVirtualSchema),
		unpackRpmVirtualRepository,
		constructor,
	)
}
