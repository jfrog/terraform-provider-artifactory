package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type DockerLocalRepositoryParams struct {
	RepositoryBaseParams
	MaxUniqueTags       int    `hcl:"max_unique_tags" json:"maxUniqueTags"`
	DockerApiVersion    string `hcl:"api_version" json:"dockerApiVersion"`
	TagRetention        int    `hcl:"tag_retention" json:"dockerTagRetention"`
	BlockPushingSchema1 bool   `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
}

var dockerV2Schema = lo.Assign(
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
		"block_pushing_schema1": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "When set, Artifactory will block the pushing of Docker images with manifest v2 schema 1 to this repository.",
		},
		"api_version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The Docker API version to use. This cannot be set",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.DockerPackageType),
)

var DockerV2Schemas = GetSchemas(dockerV2Schema)

func UnpackLocalDockerV2Repository(data *schema.ResourceData, Rclass string) DockerLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return DockerLocalRepositoryParams{
		RepositoryBaseParams: UnpackBaseRepo(Rclass, data, repository.DockerPackageType),
		MaxUniqueTags:        d.GetInt("max_unique_tags", false),
		DockerApiVersion:     "V2",
		TagRetention:         d.GetInt("tag_retention", false),
		BlockPushingSchema1:  d.GetBool("block_pushing_schema1", false),
	}
}

func ResourceArtifactoryLocalDockerV2Repository() *schema.Resource {
	pkr := packer.Default(DockerV2Schemas[CurrentSchemaVersion])

	var unpackLocalDockerV2Repository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalDockerV2Repository(data, Rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &DockerLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: repository.DockerPackageType,
				Rclass:      Rclass,
			},
			DockerApiVersion:    "V2",
			TagRetention:        1,
			MaxUniqueTags:       0, // no limit
			BlockPushingSchema1: true,
		}, nil
	}

	return repository.MkResourceSchema(
		DockerV2Schemas,
		pkr,
		unpackLocalDockerV2Repository,
		constructor,
	)
}

var dockerV1Schema = utilsdk.MergeMaps(
	map[string]*schema.Schema{
		"max_unique_tags": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"tag_retention": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"block_pushing_schema1": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"api_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.DockerPackageType),
)

var DockerV1Schemas = GetSchemas(dockerV1Schema)

func UnpackLocalDockerV1Repository(data *schema.ResourceData, Rclass string) DockerLocalRepositoryParams {
	return DockerLocalRepositoryParams{
		RepositoryBaseParams: UnpackBaseRepo(Rclass, data, repository.DockerPackageType),
		DockerApiVersion:     "V1",
		MaxUniqueTags:        0,
		TagRetention:         1,
		BlockPushingSchema1:  false,
	}
}

func ResourceArtifactoryLocalDockerV1Repository() *schema.Resource {
	var unPackLocalDockerV1Repository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalDockerV1Repository(data, Rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &DockerLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: repository.DockerPackageType,
				Rclass:      Rclass,
			},
			DockerApiVersion:    "V1",
			TagRetention:        1,
			MaxUniqueTags:       0,
			BlockPushingSchema1: false,
		}, nil
	}

	return repository.MkResourceSchema(
		DockerV1Schemas,
		packer.Default(DockerV1Schemas[CurrentSchemaVersion]),
		unPackLocalDockerV1Repository,
		constructor,
	)
}
