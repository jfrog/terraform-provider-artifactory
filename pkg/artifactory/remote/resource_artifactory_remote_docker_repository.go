package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
)

type DockerRemoteRepository struct {
	RepositoryBaseParams
	ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns"`
	EnableTokenAuthentication    bool     `hcl:"enable_token_authentication" json:"enableTokenAuthentication"`
	BlockPushingSchema1          bool     `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
}

func ResourceArtifactoryRemoteDockerRepository() *schema.Resource {
	var dockerRemoteSchema = util.MergeSchema(baseRemoteSchema, map[string]*schema.Schema{
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
	return repos.MkResourceSchema(dockerRemoteSchema, util.DefaultPacker, unpackDockerRemoteRepo, func() interface{} {
		return &DockerRemoteRepository{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "remote",
				PackageType: "docker",
			},
		}
	})
}

func unpackDockerRemoteRepo(s *schema.ResourceData) (interface{}, string, error) {
	d := &util.ResourceData{ResourceData: s}
	repo := DockerRemoteRepository{
		RepositoryBaseParams:   unpackBaseRemoteRepo(s, "docker"),
		EnableTokenAuthentication:    d.GetBool("enable_token_authentication", false),
		ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
		BlockPushingSchema1:          d.GetBool("block_pushing_schema1", false),
		ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
	}
	return repo, repo.Id(), nil
}
