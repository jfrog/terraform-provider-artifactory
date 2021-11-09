package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var nugetLocalSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
	"max_unique_snapshots": {
		Type:     schema.TypeInt,
		Optional: true,
		Default:  0,
		Description: "The maximum number of unique snapshots of a single artifact to store.\nOnce the number of " +
			"snapshots exceeds this setting, older versions are removed.\nA value of 0 (default) indicates there is no limit, and unique snapshots are not cleaned up.",
	},

	"force_nuget_authentication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Force basic authentication credentials in order to use this repository.",
	},
})

func resourceArtifactoryLocalNugetRepository() *schema.Resource {

	return mkResourceSchema(nugetLocalSchema, universalPack, unPackLocalNugetRepository, func() interface{} {
		return &NugetLocalRepositoryParams{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "nuget",
				Rclass:      "local",
			},
			MaxUniqueSnapshots:       0,
			ForceNugetAuthentication: false,
		}
	})
}

type NugetLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	MaxUniqueSnapshots       int  `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots"`
	ForceNugetAuthentication bool `hcl:"force_nuget_authentication" json:"forceNugetAuthentication"`
}

func unPackLocalNugetRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{ResourceData: data}
	repo := NugetLocalRepositoryParams{
		LocalRepositoryBaseParams: unpackBaseLocalRepo(data, "nuget"),
		MaxUniqueSnapshots:        d.getInt("max_unique_snapshots", false),
		ForceNugetAuthentication:  d.getBool("force_nuget_authentication", false),
	}

	return repo, repo.Id(), nil
}
