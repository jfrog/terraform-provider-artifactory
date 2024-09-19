package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var rpmSchema = lo.Assign(
	repository.PrimaryKeyPairRef,
	repository.SecondaryKeyPairRef,
	repository.RepoLayoutRefSchema(Rclass, repository.RPMPackageType),
)

var RPMSchemas = GetSchemas(rpmSchema)

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
		repo.PackageType = repository.RPMPackageType

		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &RpmVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: repository.RPMPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		RPMSchemas,
		packer.Default(RPMSchemas[CurrentSchemaVersion]),
		unpackRpmVirtualRepository,
		constructor,
	)
}
