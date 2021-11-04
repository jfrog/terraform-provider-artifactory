package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var dockerV2LocalSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
	"max_unique_tags": {
		Type:     schema.TypeInt,
		Optional: true,
		Computed: false,
		Description: "The maximum number of unique tags of a single Docker image to store in this repository.\n" +
			"Once the number tags for an image exceeds this setting, older tags are removed. A value of 0 (default) indicates there is no limit.\n" +
			"This only applies to manifest v2",
		ValidateDiagFunc: upgrade(validation.IntAtLeast(0), "max_unique_tags"),
	},
	"tag_retention": {
		Type:             schema.TypeInt,
		Optional:         true,
		Computed:         false,
		Description:      "If greater than 1, overwritten tags will be saved by their digest, up to the set up number. This only applies to manifest V2",
		ValidateDiagFunc: upgrade(validation.IntAtLeast(1), "tag_retention"),
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
})
var dockerV1LocalSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
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
})

func resourceArtifactoryLocalDockerV2Repository() *schema.Resource {
	return mkResourceSchema(dockerV2LocalSchema, universalPack, unPackLocalDockerV2Repository, func() interface{} {
		return &DockerLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "docker",
				Rclass:      "local",
			},
			DockerApiVersion:    "V2",
			TagRetention:        1,
			MaxUniqueTags:       0, // no limit
			BlockPushingSchema1: true,
		}
	})
}

func resourceArtifactoryLocalDockerV1Repository() *schema.Resource {
	// this is necessary because of the pointers
	skeema := mergeSchema(map[string]*schema.Schema{}, dockerV1LocalSchema)
	for key, value := range dockerV2LocalSchema {
		skeema[key].Description = value.Description
	}

	return mkResourceSchema(skeema, universalPack, unPackLocalDockerV1Repository, func() interface{} {
		return &DockerLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "docker",
				Rclass:      "local",
			},
			DockerApiVersion:    "V1",
			TagRetention:        1,
			MaxUniqueTags:       0,
			BlockPushingSchema1: false,
		}
	})
}

type DockerLocalRepo struct {
	LocalRepositoryBaseParams
	MaxUniqueTags       int    `hcl:"max_unique_tags" json:"maxUniqueTags,omitempty"`
	DockerApiVersion    string `hcl:"api_version" json:"dockerApiVersion"`
	TagRetention        int    `hcl:"tag_retention" json:"dockerTagRetention"`
	BlockPushingSchema1 bool   `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
}

func unPackLocalDockerV1Repository(data *schema.ResourceData) (interface{}, string, error) {
	repo := DockerLocalRepo{
		LocalRepositoryBaseParams: unpackBaseLocalRepo(data, "docker"),
		MaxUniqueTags:             0,
		DockerApiVersion:          "V1",
		TagRetention:              1,
		BlockPushingSchema1:       false,
	}

	return repo, repo.Id(), nil
}
func unPackLocalDockerV2Repository(data *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{ResourceData: data}
	repo := DockerLocalRepo{
		LocalRepositoryBaseParams: unpackBaseLocalRepo(data, "docker"),
		MaxUniqueTags:             d.getInt("max_unique_tags", false),
		DockerApiVersion:          "V2",
		TagRetention:              d.getInt("tag_retention", false),
		BlockPushingSchema1:       d.getBool("block_pushing_schema1", false),
	}

	return repo, repo.Id(), nil
}
