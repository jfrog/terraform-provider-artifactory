package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const debianPackageType = "debian"

var DebianLocalSchema = utilsdk.MergeMaps(
	BaseLocalRepoSchema,
	map[string]*schema.Schema{
		"primary_keypair_ref": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Used to sign index files in Debian artifacts. ",
		},
		"secondary_keypair_ref": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Used to sign index files in Debian artifacts. ",
		},
		"trivial_layout": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, the repository will use the deprecated trivial layout.",
			Deprecated:  "You shouldn't be using this",
		},
	},
	repository.RepoLayoutRefSchema(rclass, debianPackageType),
	repository.CompressionFormats,
)

type DebianLocalRepositoryParams struct {
	RepositoryBaseParams
	TrivialLayout           bool     `hcl:"trivial_layout" json:"debianTrivialLayout"`
	IndexCompressionFormats []string `hcl:"index_compression_formats" json:"optionalIndexCompressionFormats,omitempty"`
	PrimaryKeyPairRef       string   `hcl:"primary_keypair_ref" json:"primaryKeyPairRef,omitempty"`
	SecondaryKeyPairRef     string   `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef,omitempty"`
}

func UnpackLocalDebianRepository(data *schema.ResourceData, rclass string) DebianLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return DebianLocalRepositoryParams{
		RepositoryBaseParams:    UnpackBaseRepo(rclass, data, debianPackageType),
		PrimaryKeyPairRef:       d.GetString("primary_keypair_ref", false),
		SecondaryKeyPairRef:     d.GetString("secondary_keypair_ref", false),
		TrivialLayout:           d.GetBool("trivial_layout", false),
		IndexCompressionFormats: d.GetSet("index_compression_formats"),
	}
}

func ResourceArtifactoryLocalDebianRepository() *schema.Resource {

	var unpackLocalDebianRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalDebianRepository(data, rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &DebianLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: debianPackageType,
				Rclass:      rclass,
			},
		}, nil
	}

	return repository.MkResourceSchema(DebianLocalSchema, packer.Default(DebianLocalSchema), unpackLocalDebianRepository, constructor)
}
