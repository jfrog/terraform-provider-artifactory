package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceArtifactoryAlpineVirtualRepository() *schema.Resource {

	const packageType = "alpine"

	var alpineVirtualSchema = mergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{
		"primary_keypair_ref": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "(Optional) Primary keypair used to sign artifacts. Default value is empty.",
		},
	}, repoLayoutRefSchema("virtual", packageType))

	type AlpineVirtualRepositoryParams struct {
		VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs
		PrimaryKeyPairRef string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
	}

	var unpackAlpineVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}

		repo := AlpineVirtualRepositoryParams{
			VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs: unpackBaseVirtRepoWithRetrievalCachePeriodSecs(s, packageType),
			PrimaryKeyPairRef: d.getString("primary_keypair_ref", false),
		}
		repo.PackageType = packageType
		return &repo, repo.Key, nil
	}

	return mkResourceSchema(alpineVirtualSchema, defaultPacker, unpackAlpineVirtualRepository, func() interface{} {
		return &AlpineVirtualRepositoryParams{
			VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs: VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs{
				VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
					Rclass:      "virtual",
					PackageType: packageType,
				},
			},
		}
	})
}
