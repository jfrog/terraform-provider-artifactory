package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const nugetPackageType = "nuget"

var NugetLocalSchema = utilsdk.MergeMaps(
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
	repository.RepoLayoutRefSchema(rclass, nugetPackageType),
)

type NugetLocalRepositoryParams struct {
	RepositoryBaseParams
	MaxUniqueSnapshots       int  `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots"`
	ForceNugetAuthentication bool `hcl:"force_nuget_authentication" json:"forceNugetAuthentication"`
}

func UnpackLocalNugetRepository(data *schema.ResourceData, rclass string) NugetLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return NugetLocalRepositoryParams{
		RepositoryBaseParams:     UnpackBaseRepo(rclass, data, nugetPackageType),
		MaxUniqueSnapshots:       d.GetInt("max_unique_snapshots", false),
		ForceNugetAuthentication: d.GetBool("force_nuget_authentication", false),
	}
}

func ResourceArtifactoryLocalNugetRepository() *schema.Resource {

	var unPackLocalNugetRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalNugetRepository(data, rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &NugetLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: nugetPackageType,
				Rclass:      rclass,
			},
			MaxUniqueSnapshots:       0,
			ForceNugetAuthentication: false,
		}, nil
	}

	return repository.MkResourceSchema(NugetLocalSchema, packer.Default(NugetLocalSchema), unPackLocalNugetRepository, constructor)
}
