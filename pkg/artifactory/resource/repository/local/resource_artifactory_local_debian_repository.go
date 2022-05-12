package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

var debianLocalSchema = util.MergeSchema(BaseLocalRepoSchema, map[string]*schema.Schema{
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
		Description: "When set, the repository will use the deprecated trivial layout.",
		Deprecated:  "You shouldn't be using this",
	},
}, repository.RepoLayoutRefSchema("local", "debian"), repository.CompressionFormats)

func ResourceArtifactoryLocalDebianRepository() *schema.Resource {

	type DebianLocalRepositoryParams struct {
		LocalRepositoryBaseParams
		TrivialLayout           bool     `hcl:"trivial_layout" json:"debianTrivialLayout,omitempty"`
		IndexCompressionFormats []string `hcl:"index_compression_formats" json:"optionalIndexCompressionFormats,omitempty"`
		PrimaryKeyPairRef       string   `hcl:"primary_keypair_ref" json:"primaryKeyPairRef,omitempty"`
		SecondaryKeyPairRef     string   `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef,omitempty"`
	}

	var unPackLocalDebianRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: data}
		repo := DebianLocalRepositoryParams{
			LocalRepositoryBaseParams: UnpackBaseRepo("local", data, "debian"),
			PrimaryKeyPairRef:         d.GetString("primary_keypair_ref", false),
			SecondaryKeyPairRef:       d.GetString("secondary_keypair_ref", false),
			TrivialLayout:             d.GetBool("trivial_layout", false),
			IndexCompressionFormats:   d.GetSet("index_compression_formats"),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(debianLocalSchema, repository.DefaultPacker(debianLocalSchema), unPackLocalDebianRepository, func() interface{} {
		return &DebianLocalRepositoryParams{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "debian",
				Rclass:      "local",
			},
		}
	})
}
