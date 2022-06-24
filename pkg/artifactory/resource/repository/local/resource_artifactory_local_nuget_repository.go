package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

var nugetLocalSchema = util.MergeSchema(
	BaseLocalRepoSchema,
	map[string]*schema.Schema{
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
	},
	repository.RepoLayoutRefSchema("local", "nuget"),
)

func ResourceArtifactoryLocalNugetRepository() *schema.Resource {

	type NugetLocalRepositoryParams struct {
		LocalRepositoryBaseParams
		MaxUniqueSnapshots       int  `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots"`
		ForceNugetAuthentication bool `hcl:"force_nuget_authentication" json:"forceNugetAuthentication"`
	}

	var unPackLocalNugetRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: data}
		repo := NugetLocalRepositoryParams{
			LocalRepositoryBaseParams: UnpackBaseRepo("local", data, "nuget"),
			MaxUniqueSnapshots:        d.GetInt("max_unique_snapshots", false),
			ForceNugetAuthentication:  d.GetBool("force_nuget_authentication", false),
		}

		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(nugetLocalSchema, packer.Default(nugetLocalSchema), unPackLocalNugetRepository, func() interface{} {
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
