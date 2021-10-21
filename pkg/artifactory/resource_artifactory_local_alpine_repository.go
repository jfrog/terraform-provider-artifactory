package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var alpineLocalSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
	"primary_keypair_ref": {
		Type:     schema.TypeString,
		Optional: true,
		Description: "Used to sign index files in Alpine Linux repositories. " +
			"See: https://www.jfrog.com/confluence/display/JFROG/Alpine+Linux+Repositories#AlpineLinuxRepositories-SigningAlpineLinuxIndex",
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
	PrimaryKeyPairRef string `hcl:"primary_key_pair_ref" json:"primaryKeyPairRef"`
}

func unPackLocalAlpineRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{ResourceData: data}
	repo := AlpineLocalRepo{
		LocalRepositoryBaseParams: LocalRepositoryBaseParams{
			Rclass:                 "local",
			Key:                    d.getString("key", false),
			PackageType:            "alpine",
			Description:            d.getString("description", false),
			Notes:                  d.getString("notes", false),
			IncludesPattern:        d.getString("includes_pattern", false),
			ExcludesPattern:        d.getString("excludes_pattern", false),
			RepoLayoutRef:          d.getString("repo_layout_ref", false),
			BlackedOut:             d.getBoolRef("blacked_out", false),
			ArchiveBrowsingEnabled: d.getBoolRef("archive_browsing_enabled", false),
			PropertySets:           d.getSet("property_sets"),
			XrayIndex:              d.getBoolRef("xray_index", false),
		},
		PrimaryKeyPairRef: d.getString("primary_key_pair_ref", false),
	}

	return repo, repo.Key, nil
}

func packLocalAlpineRepository(r interface{}, d *schema.ResourceData) error {

	repo := r.(*AlpineLocalRepo)
	setValue := mkLens(d)

	setValue("key", repo.Key)
	// type 'yum' is not to be supported, as this is really of type 'rpm'. When 'yum' is used on create, RT will
	// respond with 'rpm' and thus confuse TF into think there has been a state change.
	setValue("package_type", repo.PackageType)
	setValue("description", repo.Description)
	setValue("notes", repo.Notes)
	setValue("includes_pattern", repo.IncludesPattern)
	setValue("excludes_pattern", repo.ExcludesPattern)
	setValue("repo_layout_ref", repo.RepoLayoutRef)
	setValue("blacked_out", repo.BlackedOut)
	setValue("archive_browsing_enabled", repo.ArchiveBrowsingEnabled)
	setValue("property_sets", schema.NewSet(schema.HashString, castToInterfaceArr(repo.PropertySets)))
	setValue("primary_key_pair_ref", repo.PrimaryKeyPairRef)
	errors := setValue("xray_index", repo.XrayIndex)

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed saving state for local repos %q", errors)
	}

	return nil
}
