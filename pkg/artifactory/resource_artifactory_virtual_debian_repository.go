package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
)

func resourceArtifactoryDebianVirtualRepository() *schema.Resource {

	const packageType = "debian"

	var debianVirtualSchema = mergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{
		"primary_keypair_ref": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "(Optional) Primary keypair used to sign artifacts. Default is empty.",
		},
		"secondary_keypair_ref": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "(Optional) Secondary keypair used to sign artifacts. Default is empty.",
		},
		"optional_index_compression_formats": {
			Type:     schema.TypeSet,
			Optional: true,
			MinItems: 0,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"bz2",
					"lzma",
					"xz",
				}, false),
			},
			Description: `(Optional) Index file formats you would like to create in addition to the default Gzip (.gzip extension). Supported values are 'bz2','lzma' and 'xz'. Default value is 'bz2'.`,
		},
		"debian_default_architectures": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "amd64,i386",
			ValidateDiagFunc: validation.ToDiagFunc(validation.All(validation.StringIsNotEmpty, validation.StringMatch(regexp.MustCompile(`.+(?:,.+)*`), "must be comma separated string"))),
			StateFunc:        formatCommaSeparatedString,
			Description:      `(Optional) Specifying  architectures will speed up Artifactory's initial metadata indexing process. The default architecture values are amd64 and i386.`,
		},
	}, repoLayoutRefSchema("virtual", packageType))

	type DebianVirtualRepositoryParams struct {
		VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs
		OptionalIndexCompressionFormats []string `json:"optionalIndexCompressionFormats"`
		PrimaryKeyPairRef               string   `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
		SecondaryKeyPairRef             string   `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef"`
		DebianDefaultArchitectures      string   `json:"debianDefaultArchitectures"`
	}

	var unpackDebianVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}

		repo := DebianVirtualRepositoryParams{
			VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs: unpackBaseVirtRepoWithRetrievalCachePeriodSecs(s, packageType),
			OptionalIndexCompressionFormats:                         d.getSet("optional_index_compression_formats"),
			PrimaryKeyPairRef:                                       d.getString("primary_keypair_ref", false),
			SecondaryKeyPairRef:                                     d.getString("secondary_keypair_ref", false),
			DebianDefaultArchitectures:                              d.getString("debian_default_architectures", false),
		}
		repo.PackageType = packageType
		return &repo, repo.Key, nil
	}

	return mkResourceSchema(debianVirtualSchema, defaultPacker, unpackDebianVirtualRepository, func() interface{} {
		return &DebianVirtualRepositoryParams{
			VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs: VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs{
				VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
					Rclass:      "virtual",
					PackageType: packageType,
				},
			},
		}
	})
}
