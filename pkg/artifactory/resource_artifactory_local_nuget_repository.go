package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArtifactoryLocalNugetRepository() *schema.Resource {

	const packageType = "nuget"

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
	}, repoLayoutRefSchema("local", packageType))

	type NugetLocalRepositoryParams struct {
		LocalRepositoryBaseParams
		MaxUniqueSnapshots       int  `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots"`
		ForceNugetAuthentication bool `hcl:"force_nuget_authentication" json:"forceNugetAuthentication"`
	}

	var unPackLocalNugetRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{ResourceData: data}
		repo := NugetLocalRepositoryParams{
			LocalRepositoryBaseParams: unpackBaseRepo("local", data, packageType),
			MaxUniqueSnapshots:        d.getInt("max_unique_snapshots", false),
			ForceNugetAuthentication:  d.getBool("force_nuget_authentication", false),
		}

		return repo, repo.Id(), nil
	}

	return mkResourceSchema(nugetLocalSchema, defaultPacker, unPackLocalNugetRepository, func() interface{} {
		return &NugetLocalRepositoryParams{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: packageType,
				Rclass:      "local",
			},
			MaxUniqueSnapshots:       0,
			ForceNugetAuthentication: false,
		}
	})
}
