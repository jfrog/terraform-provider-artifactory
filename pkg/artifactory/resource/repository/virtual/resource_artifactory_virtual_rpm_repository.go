package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const RpmPackageType = "rpm"

var RpmVirtualSchema = utilsdk.MergeMaps(BaseVirtualRepoSchema, map[string]*schema.Schema{
	"primary_keypair_ref": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Primary keypair used to sign artifacts.",
	},
	"secondary_keypair_ref": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Secondary keypair used to sign artifacts.",
	},
}, repository.RepoLayoutRefSchema(Rclass, RpmPackageType))

func ResourceArtifactoryVirtualRpmRepository() *schema.Resource {
	type CommonRpmDebianVirtualRepositoryParams struct {
		PrimaryKeyPairRef   string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
		SecondaryKeyPairRef string `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef"`
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
				PrimaryKeyPairRef:   d.GetString("primary_keypair_ref", false),
				SecondaryKeyPairRef: d.GetString("secondary_keypair_ref", false),
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
