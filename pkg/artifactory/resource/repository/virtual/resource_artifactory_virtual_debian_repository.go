package virtual

import (
	"github.com/jfrog/terraform-provider-shared/packer"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryVirtualDebianRepository() *schema.Resource {

	const packageType = "debian"

	var debianVirtualSchema = util.MergeSchema(BaseVirtualRepoSchema, map[string]*schema.Schema{
		"primary_keypair_ref": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Primary keypair used to sign artifacts. Default is empty.",
		},
		"secondary_keypair_ref": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Secondary keypair used to sign artifacts. Default is empty.",
		},
		"optional_index_compression_formats": {
			Type:     schema.TypeSet,
			Optional: true,
			MinItems: 0,
			Computed: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"bz2", "lzma", "xz"}, false),
			},
			Description: `Index file formats you would like to create in addition to the default Gzip (.gzip extension). Supported values are 'bz2','lzma' and 'xz'. Default value is 'bz2'.`,
		},
		"debian_default_architectures": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "amd64,i386",
			ValidateDiagFunc: validation.ToDiagFunc(validation.All(validation.StringIsNotEmpty, validation.StringMatch(regexp.MustCompile(`.+(?:,.+)*`), "must be comma separated string"))),
			StateFunc:        util.FormatCommaSeparatedString,
			Description:      `Specifying  architectures will speed up Artifactory's initial metadata indexing process. The default architecture values are amd64 and i386.`,
		},
	}, repository.RepoLayoutRefSchema("virtual", packageType))

	type DebianVirtualRepositoryParams struct {
		RepositoryBaseParamsWithRetrievalCachePeriodSecs
		OptionalIndexCompressionFormats []string `json:"optionalIndexCompressionFormats"`
		PrimaryKeyPairRef               string   `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
		SecondaryKeyPairRef             string   `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef"`
		DebianDefaultArchitectures      string   `json:"debianDefaultArchitectures"`
	}

	var unpackDebianVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}

		repo := DebianVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(s, packageType),
			OptionalIndexCompressionFormats:                  d.GetSet("optional_index_compression_formats"),
			PrimaryKeyPairRef:                                d.GetString("primary_keypair_ref", false),
			SecondaryKeyPairRef:                              d.GetString("secondary_keypair_ref", false),
			DebianDefaultArchitectures:                       d.GetString("debian_default_architectures", false),
		}
		repo.PackageType = packageType
		return &repo, repo.Key, nil
	}

	return repository.MkResourceSchema(debianVirtualSchema, packer.Default(debianVirtualSchema), unpackDebianVirtualRepository, func() interface{} {
		return &DebianVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      "virtual",
					PackageType: packageType,
				},
			},
		}
	})
}
