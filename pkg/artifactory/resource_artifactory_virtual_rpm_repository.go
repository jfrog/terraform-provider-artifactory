package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var rpmVirtualSchema = mergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{
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
})

type CommonRpmDebianVirtualRepositoryParams struct {
	PrimaryKeyPairRef   string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
	SecondaryKeyPairRef string `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef"`
}

type RpmVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	CommonRpmDebianVirtualRepositoryParams
}

func resourceArtifactoryRpmVirtualRepository() *schema.Resource {
	return mkResourceSchema(rpmVirtualSchema, defaultPacker, unpackRpmVirtualRepository, func() interface{} {
		return &RpmVirtualRepositoryParams{
			VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: "rpm",
			},
		}
	})
}

func unpackRpmVirtualRepository(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}

	repo := RpmVirtualRepositoryParams{
		VirtualRepositoryBaseParams: unpackBaseVirtRepo(s, "rpm"),
		CommonRpmDebianVirtualRepositoryParams: CommonRpmDebianVirtualRepositoryParams{
			PrimaryKeyPairRef:   d.getString("primary_keypair_ref", false),
			SecondaryKeyPairRef: d.getString("secondary_keypair_ref", false),
		},
	}
	repo.PackageType = "rpm"

	return &repo, repo.Key, nil
}
