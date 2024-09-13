package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const HelmOciPackageType = "helmoci"

type HelmOciLocalRepositoryParams struct {
	RepositoryBaseParams
	MaxUniqueTags int `json:"maxUniqueTags"`
	TagRetention  int `json:"dockerTagRetention"`
}

var HelmOciLocalSchema = utilsdk.MergeMaps(
	BaseLocalRepoSchema,
	map[string]*schema.Schema{
		"max_unique_tags": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			Description: "The maximum number of unique tags of a single Docker image to store in this repository.\n" +
				"Once the number tags for an image exceeds this setting, older tags are removed. A value of 0 (default) indicates there is no limit.\n" +
				"This only applies to manifest v2",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"tag_retention": {
			Type:             schema.TypeInt,
			Optional:         true,
			Computed:         false,
			Description:      "If greater than 1, overwritten tags will be saved by their digest, up to the set up number. This only applies to manifest V2",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
		},
	},
	repository.RepoLayoutRefSchema(rclass, HelmOciPackageType),
)

func UnpackLocalHelmOciRepository(data *schema.ResourceData, rclass string) HelmOciLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return HelmOciLocalRepositoryParams{
		RepositoryBaseParams: UnpackBaseRepo(rclass, data, HelmOciPackageType),
		MaxUniqueTags:        d.GetInt("max_unique_tags", false),
		TagRetention:         d.GetInt("tag_retention", false),
	}
}

func ResourceArtifactoryLocalHelmOciRepository() *schema.Resource {
	pkr := packer.Default(HelmOciLocalSchema)

	var unpackLocalHelmOciRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalHelmOciRepository(data, rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &HelmOciLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: HelmOciPackageType,
				Rclass:      rclass,
			},
			TagRetention:  1,
			MaxUniqueTags: 0, // no limit
		}, nil
	}

	return repository.MkResourceSchema(HelmOciLocalSchema, pkr, unpackLocalHelmOciRepository, constructor)
}
