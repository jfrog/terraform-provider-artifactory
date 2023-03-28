package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

const dockerPackageType = "docker"

type DockerLocalRepositoryParams struct {
	RepositoryBaseParams
	MaxUniqueTags       int    `hcl:"max_unique_tags" json:"maxUniqueTags"`
	DockerApiVersion    string `hcl:"api_version" json:"dockerApiVersion"`
	TagRetention        int    `hcl:"tag_retention" json:"dockerTagRetention"`
	BlockPushingSchema1 bool   `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
}

var DockerV2LocalSchema = util.MergeMaps(
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
	repository.RepoLayoutRefSchema(rclass, dockerPackageType),
)

func UnpackLocalDockerV2Repository(data *schema.ResourceData, rclass string) DockerLocalRepositoryParams {
	d := &util.ResourceData{ResourceData: data}
	return DockerLocalRepositoryParams{
		RepositoryBaseParams: UnpackBaseRepo(rclass, data, dockerPackageType),
		MaxUniqueTags:        d.GetInt("max_unique_tags", false),
		DockerApiVersion:     "V2",
		TagRetention:         d.GetInt("tag_retention", false),
		BlockPushingSchema1:  d.GetBool("block_pushing_schema1", false),
	}
}

func ResourceArtifactoryLocalDockerV2Repository() *schema.Resource {
	pkr := packer.Default(DockerV2LocalSchema)

	var unpackLocalDockerV2Repository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalDockerV2Repository(data, rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &DockerLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: dockerPackageType,
				Rclass:      rclass,
			},
			DockerApiVersion:    "V2",
			TagRetention:        1,
			MaxUniqueTags:       0, // no limit
			BlockPushingSchema1: true,
		}, nil
	}

	return repository.MkResourceSchema(DockerV2LocalSchema, pkr, unpackLocalDockerV2Repository, constructor)
}

var DockerV1LocalSchema = util.MergeMaps(
	BaseLocalRepoSchema,
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
	repository.RepoLayoutRefSchema(rclass, dockerPackageType),
)

func UnpackLocalDockerV1Repository(data *schema.ResourceData, rclass string) DockerLocalRepositoryParams {
	return DockerLocalRepositoryParams{
		RepositoryBaseParams: UnpackBaseRepo(rclass, data, dockerPackageType),
		DockerApiVersion:     "V1",
		MaxUniqueTags:        0,
		TagRetention:         1,
		BlockPushingSchema1:  false,
	}
}

func ResourceArtifactoryLocalDockerV1Repository() *schema.Resource {
	// this is necessary because of the pointers
	skeema := util.MergeMaps(DockerV1LocalSchema)

	var unPackLocalDockerV1Repository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalDockerV1Repository(data, rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &DockerLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: dockerPackageType,
				Rclass:      rclass,
			},
			DockerApiVersion:    "V1",
			TagRetention:        1,
			MaxUniqueTags:       0,
			BlockPushingSchema1: false,
		}, nil
	}

	return repository.MkResourceSchema(skeema, packer.Default(DockerV1LocalSchema), unPackLocalDockerV1Repository, constructor)
}
