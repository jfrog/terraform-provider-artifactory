package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type DockerRemoteRepository struct {
	RemoteRepositoryBaseParams
	ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns"`
	EnableTokenAuthentication    bool     `hcl:"enable_token_authentication" json:"enableTokenAuthentication"`
	BlockPushingSchema1          bool     `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
}

func resourceArtifactoryRemoteDockerRepository() *schema.Resource {
	const packageType = "docker"

	var dockerRemoteSchema = mergeSchema(getBaseRemoteRepoSchema(packageType), map[string]*schema.Schema{
		"external_dependencies_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Also known as 'Foreign Layers Caching' on the UI",
		},
		"enable_token_authentication": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Enable token (Bearer) based authentication.",
		},
		"block_pushing_schema1": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "When set, Artifactory will block the pulling of Docker images with manifest v2 schema 1 from the remote repository (i.e. the upstream). It will be possible to pull images with manifest v2 schema 1 that exist in the cache.",
		},
		"external_dependencies_patterns": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			RequiredWith: []string{"external_dependencies_enabled"},
			Description: "An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will " +
				"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response. " +
				"By default, this is set to '**', which means that remote modules may be downloaded from any external VCS source.",
		},
	})

	var unpackDockerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}
		repo := DockerRemoteRepository{
			RemoteRepositoryBaseParams:   unpackBaseRemoteRepo(s, packageType),
			EnableTokenAuthentication:    d.getBool("enable_token_authentication", false),
			ExternalDependenciesEnabled:  d.getBool("external_dependencies_enabled", false),
			BlockPushingSchema1:          d.getBool("block_pushing_schema1", false),
			ExternalDependenciesPatterns: d.getList("external_dependencies_patterns"),
		}
		return repo, repo.Id(), nil
	}

	return mkResourceSchema(dockerRemoteSchema, defaultPacker, unpackDockerRemoteRepo, func() interface{} {
		return &DockerRemoteRepository{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:      "remote",
				PackageType: packageType,
			},
		}
	})
}
