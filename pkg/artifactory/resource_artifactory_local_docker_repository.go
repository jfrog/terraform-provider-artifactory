package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceArtifactoryLocalDockerV2Repository() *schema.Resource {

	const packageType = "docker"

	var dockerV2LocalSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
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
	}, repoLayoutRefSchema("local", packageType))

	packer := universalPack(
		allHclPredicate(
			noClass, schemaHasKey(dockerV2LocalSchema),
		),
	)

	var unPackLocalDockerV2Repository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{ResourceData: data}
		repo := DockerLocalRepositoryParams{
			LocalRepositoryBaseParams: unpackBaseRepo("local", data, packageType),
			MaxUniqueTags:             d.getInt("max_unique_tags", false),
			DockerApiVersion:          "V2",
			TagRetention:              d.getInt("tag_retention", false),
			BlockPushingSchema1:       d.getBool("block_pushing_schema1", false),
		}

		return repo, repo.Id(), nil
	}

	return mkResourceSchema(dockerV2LocalSchema, packer, unPackLocalDockerV2Repository, func() interface{} {
		return &DockerLocalRepositoryParams{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: packageType,
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

	const packageType = "docker"

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
	}, repoLayoutRefSchema("local", packageType))

	// this is necessary because of the pointers
	skeema := mergeSchema(map[string]*schema.Schema{}, dockerV1LocalSchema)

	var unPackLocalDockerV1Repository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := DockerLocalRepositoryParams{
			LocalRepositoryBaseParams: unpackBaseRepo("local", data, packageType),
			MaxUniqueTags:             0,
			DockerApiVersion:          "V1",
			TagRetention:              1,
			BlockPushingSchema1:       false,
		}

		return repo, repo.Id(), nil
	}

	return mkResourceSchema(skeema, defaultPacker, unPackLocalDockerV1Repository, func() interface{} {
		return &DockerLocalRepositoryParams{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: packageType,
				Rclass:      "local",
			},
			DockerApiVersion:    "V1",
			TagRetention:        1,
			MaxUniqueTags:       0,
			BlockPushingSchema1: false,
		}
	})
}

type DockerLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	MaxUniqueTags       int    `hcl:"max_unique_tags" json:"maxUniqueTags"`
	DockerApiVersion    string `hcl:"api_version" json:"dockerApiVersion"`
	TagRetention        int    `hcl:"tag_retention" json:"dockerTagRetention"`
	BlockPushingSchema1 bool   `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
}
