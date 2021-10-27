package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var alpineLocalSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
	"primary_keypair_ref": {
		Type:     schema.TypeString,
		Optional: true,
		Description: "Used to sign index files in Alpine Linux repositories. " +
			"See: https://www.jfrog.com/confluence/display/JFROG/Alpine+Linux+Repositories#AlpineLinuxRepositories-SigningAlpineLinuxIndex",
	},

	"index_compression_formats": {
		Type: schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional: true,
	},
})

func resourceArtifactoryLocalAlpineRepository() *schema.Resource {
	return mkResourceSchema(alpineLocalSchema, universalPack, unPackLocalAlpineRepository, func() interface{} {
		return &AlpineLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "alpine",
				Rclass:      "local",
			},
		}
	})
}

type AlpineLocalRepo struct {
	LocalRepositoryBaseParams
	PrimaryKeyPairRef string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
}

func unPackLocalAlpineRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{ResourceData: data}
	repo := AlpineLocalRepo{
		LocalRepositoryBaseParams: LocalRepositoryBaseParams{
			Rclass:                          "local",
			Key:                             d.getString("key", false),
			PackageType:                     "alpine",
			Description:                     d.getString("description", false),
			Notes:                           d.getString("notes", false),
			IncludesPattern:                 d.getString("includes_pattern", false),
			ExcludesPattern:                 d.getString("excludes_pattern", false),
			RepoLayoutRef:                   d.getString("repo_layout_ref", false),
			BlackedOut:                      d.getBoolRef("blacked_out", false),
			ArchiveBrowsingEnabled:          d.getBoolRef("archive_browsing_enabled", false),
			PropertySets:                    d.getSet("property_sets"),
			OptionalIndexCompressionFormats: d.getList("index_compression_formats"),
			XrayIndex:                       d.getBoolRef("xray_index", false),
		},
		PrimaryKeyPairRef: d.getString("primary_keypair_ref", false),
	}

	return repo, repo.Key, nil
}
