package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
)

func ResourceArtifactoryLocalNugetRepository() *schema.Resource {
	var nugetLocalSchema = util.MergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
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
	packer := util.UniversalPack(util.SchemaHasKey(nugetLocalSchema))
	return repos.MkResourceSchema(nugetLocalSchema, packer, unPackLocalNugetRepository, func() interface{} {
		return &NugetLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: "nuget",
				Rclass:      "local",
			},
			MaxUniqueSnapshots:       0,
			ForceNugetAuthentication: false,
		}
	})
}

type NugetLocalRepositoryParams struct {
	RepositoryBaseParams
	MaxUniqueSnapshots       int  `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots"`
	ForceNugetAuthentication bool `hcl:"force_nuget_authentication" json:"forceNugetAuthentication"`
}

func unPackLocalNugetRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &util.ResourceData{ResourceData: data}
	repo := NugetLocalRepositoryParams{
		RepositoryBaseParams:     unpackBaseLocalRepo(data, "nuget"),
		MaxUniqueSnapshots:       d.GetInt("max_unique_snapshots", false),
		ForceNugetAuthentication: d.GetBool("force_nuget_authentication", false),
	}

	return repo, repo.Id(), nil
}
