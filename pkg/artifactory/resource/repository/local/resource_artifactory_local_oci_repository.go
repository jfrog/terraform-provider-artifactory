package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type OciLocalRepositoryParams struct {
	RepositoryBaseParams
	MaxUniqueTags    int    `json:"maxUniqueTags"`
	DockerApiVersion string `json:"dockerApiVersion"`
	TagRetention     int    `json:"dockerTagRetention"`
}

var ociSchema = lo.Assign(
	map[string]*schema.Schema{
		"max_unique_tags": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			Description: "The maximum number of unique tags of a single OCI image to store in this repository.\n" +
				"Once the number tags for an image exceeds this setting, older tags are removed. A value of 0 (default) indicates there is no limit.",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"tag_retention": {
			Type:             schema.TypeInt,
			Optional:         true,
			Computed:         false,
			Description:      "If greater than 1, overwritten tags will be saved by their digest, up to the set up number.",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.OCIPackageType),
)

var OCILocalSchemas = GetSchemas(ociSchema)

func UnpackLocalOciRepository(data *schema.ResourceData, Rclass string) OciLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return OciLocalRepositoryParams{
		RepositoryBaseParams: UnpackBaseRepo(Rclass, data, repository.OCIPackageType),
		MaxUniqueTags:        d.GetInt("max_unique_tags", false),
		DockerApiVersion:     "V2",
		TagRetention:         d.GetInt("tag_retention", false),
	}
}

func ResourceArtifactoryLocalOciRepository() *schema.Resource {
	var unpackLocalOciRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalOciRepository(data, Rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &OciLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: repository.OCIPackageType,
				Rclass:      Rclass,
			},
			DockerApiVersion: "V2",
			TagRetention:     1,
			MaxUniqueTags:    0, // no limit
		}, nil
	}

	return repository.MkResourceSchema(
		OCILocalSchemas,
		packer.Default(OCILocalSchemas[CurrentSchemaVersion]),
		unpackLocalOciRepository,
		constructor,
	)
}
