package virtual

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var debianSchema = lo.Assign(
	RetrievalCachePeriodSecondsSchema,
	repository.PrimaryKeyPairRef,
	repository.SecondaryKeyPairRef,
	map[string]*schema.Schema{
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
			StateFunc:        utilsdk.FormatCommaSeparatedString,
			Description:      `Specifying  architectures will speed up Artifactory's initial metadata indexing process. The default architecture values are amd64 and i386.`,
		},
	}, repository.RepoLayoutRefSchema(Rclass, repository.DebianPackageType),
)

var DebianSchemas = GetSchemas(debianSchema)

func ResourceArtifactoryVirtualDebianRepository() *schema.Resource {

	type DebianVirtualRepositoryParams struct {
		RepositoryBaseParamsWithRetrievalCachePeriodSecs
		repository.PrimaryKeyPairRefParam
		repository.SecondaryKeyPairRefParam
		OptionalIndexCompressionFormats []string `json:"optionalIndexCompressionFormats"`
		DebianDefaultArchitectures      string   `json:"debianDefaultArchitectures"`
	}

	var unpackDebianVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}

		repo := DebianVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(s, repository.DebianPackageType),
			PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
				PrimaryKeyPairRef: d.GetString("primary_keypair_ref", false),
			},
			SecondaryKeyPairRefParam: repository.SecondaryKeyPairRefParam{
				SecondaryKeyPairRef: d.GetString("secondary_keypair_ref", false),
			},
			OptionalIndexCompressionFormats: d.GetSet("optional_index_compression_formats"),
			DebianDefaultArchitectures:      d.GetString("debian_default_architectures", false),
		}
		repo.PackageType = repository.DebianPackageType
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &DebianVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: repository.DebianPackageType,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(
		DebianSchemas,
		packer.Default(DebianSchemas[CurrentSchemaVersion]),
		unpackDebianVirtualRepository,
		constructor,
	)
}
