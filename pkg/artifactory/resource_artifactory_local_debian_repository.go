package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var debianLocalSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
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
	"index_compression_formats": {
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Set:      schema.HashString,
		Optional: true,
	},
})

func resourceArtifactoryLocalDebianRepository() *schema.Resource {

	return mkResourceSchema(debianLocalSchema, universalPack, unPackLocalDebianRepository, func() interface{} {
		return &DebianLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "debian",
				Rclass:      "local",
			},
		}
	})
}

type DebianLocalRepo struct {
	LocalRepositoryBaseParams
	TrivialLayout           bool     `hcl:"trivial_layout" json:"debianTrivialLayout,omitempty"`
	IndexCompressionFormats []string `hcl:"index_compression_formats" json:"optionalIndexCompressionFormats,omitempty"`
	PrimaryKeyPairRef       string   `hcl:"primary_keypair_ref" json:"primaryKeyPairRef,omitempty"`
	SecondaryKeyPairRef     string   `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef,omitempty"`
}

func unPackLocalDebianRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{ResourceData: data}
	repo := DebianLocalRepo{
		LocalRepositoryBaseParams: unpackBaseLocalRepo(data, "debian"),
		PrimaryKeyPairRef:         d.getString("primary_keypair_ref", false),
		SecondaryKeyPairRef:       d.getString("secondary_keypair_ref", false),
		TrivialLayout:             d.getBool("trivial_layout", false),
		IndexCompressionFormats:   d.getSet("index_compression_formats"),
	}
	repo.PackageType = "debian"
	return repo, repo.Id(), nil
}
