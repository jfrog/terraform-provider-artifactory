package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var dockerV2LocalSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
	"max_unique_tags": {
		Type:     schema.TypeInt,
		Optional: true,
		Description: "The maximum number of unique tags of a single Docker image to store in this repository.\n" +
			"Once the number tags for an image exceeds this setting, older tags are removed. A value of 0 (default) indicates there is no limit.\n" +
			"This only applies to manifest v2",
	},
	"tag_retention": {
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "If greater than 1, overwritten tags will be saved by their digest, up to the set up number. This only applies to manifest V2",
	},
	"block_pushing_schema1": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "When set, Artifactory will block the pushing of Docker images with manifest v2 schema 1 to this repository.",
	},
	"api_version": {
		Type:             schema.TypeString,
		Computed:         true,
		Description:      "The Docker API version to use. This cannot be set",
	},
})

func resourceArtifactoryLocalDockerV2Repository() *schema.Resource {
	return mkResourceSchema(dockerV2LocalSchema, universalPack, unPackLocalDockerRepository, func() interface{} {
		return &DockerLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "docker",
				Rclass:      "local",
			},
			DockerApiVersion: "V2",
			TagRetention: 1,
			MaxUniqueTags: 0, // no limit
		}
	})
}

func resourceArtifactoryLocalDockerV1Repository() *schema.Resource {
	var dockerV1LocalSchema = dockerV2LocalSchema
	dockerV1LocalSchema["block_pushing_schema1"].Computed = true
	dockerV1LocalSchema["block_pushing_schema1"].Optional = false
	dockerV1LocalSchema["tag_retention"].Computed = true
	dockerV1LocalSchema["tag_retention"].Optional = false
	dockerV1LocalSchema["max_unique_tags"].Computed = true
	dockerV1LocalSchema["max_unique_tags"].Optional = false

	return mkResourceSchema(dockerV2LocalSchema, universalPack, unPackLocalDockerRepository, func() interface{} {
		return &DockerLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "docker",
				Rclass:      "local",
			},
			DockerApiVersion: "V1",
			TagRetention: 1,
			MaxUniqueTags: 0,
			BlockPushingSchema1: false,
		}
	})
}


type DockerLocalRepo struct {
	LocalRepositoryBaseParams
	MaxUniqueTags       int    `hcl:"max_unique_tags" json:"maxUniqueTags,omitempty"`
	DockerApiVersion    string `json:"dockerApiVersion"`
	TagRetention        int    `hcl:"tag_retention" json:"dockerTagRetention"`
	BlockPushingSchema1 bool  `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
}

func unPackLocalDockerRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{ResourceData: data}
	repo := DockerLocalRepo{
		LocalRepositoryBaseParams: unpackBaseLocalRepo(data, "docker"),
		MaxUniqueTags:             d.getInt("max_unique_tags", false),
		DockerApiVersion:          d.getString("api_version", false),
	}

	return repo, repo.Id(), nil
}
