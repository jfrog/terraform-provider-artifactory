package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func resourceArtifactoryLocalDebianRepository() *schema.Resource {
	const packageType = "debian"

	var debianLocalSchema = utils.MergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
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
	}, repoLayoutRefSchema("local", packageType), compressionFormats)

	type DebianLocalRepositoryParams struct {
		LocalRepositoryBaseParams
		TrivialLayout           bool     `hcl:"trivial_layout" json:"debianTrivialLayout,omitempty"`
		IndexCompressionFormats []string `hcl:"index_compression_formats" json:"optionalIndexCompressionFormats,omitempty"`
		PrimaryKeyPairRef       string   `hcl:"primary_keypair_ref" json:"primaryKeyPairRef,omitempty"`
		SecondaryKeyPairRef     string   `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef,omitempty"`
	}

	var unPackLocalDebianRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utils.ResourceData{ResourceData: data}
		repo := DebianLocalRepositoryParams{
			LocalRepositoryBaseParams: unpackBaseRepo("local", data, packageType),
			PrimaryKeyPairRef:         d.GetString("primary_keypair_ref", false),
			SecondaryKeyPairRef:       d.GetString("secondary_keypair_ref", false),
			TrivialLayout:             d.GetBool("trivial_layout", false),
			IndexCompressionFormats:   d.GetSet("index_compression_formats"),
		}
		return repo, repo.Id(), nil
	}

	return mkResourceSchema(debianLocalSchema, defaultPacker(debianLocalSchema), unPackLocalDebianRepository, func() interface{} {
		return &DebianLocalRepositoryParams{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: packageType,
				Rclass:      "local",
			},
		}
	})
}
