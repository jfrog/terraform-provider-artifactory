package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

type DockerRemoteRepository struct {
	RepositoryBaseParams
	ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns"`
	EnableTokenAuthentication    bool     `hcl:"enable_token_authentication" json:"enableTokenAuthentication"`
	BlockPushingSchema1          bool     `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
}

func ResourceArtifactoryRemoteDockerRepository() *schema.Resource {
	const packageType = "docker"

	var dockerRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, map[string]*schema.Schema{
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
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackDockerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := DockerRemoteRepository{
			RepositoryBaseParams:         UnpackBaseRemoteRepo(s, packageType),
			EnableTokenAuthentication:    d.GetBool("enable_token_authentication", false),
			ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
			BlockPushingSchema1:          d.GetBool("block_pushing_schema1", false),
			ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
		}
		if len(repo.ExternalDependenciesPatterns) == 0 {
			repo.ExternalDependenciesPatterns = []string{"**"}
		}
		return repo, repo.Id(), nil
	}

	// Special handling for "external_dependencies_patterns" attribute to match default value behavior in UI.
	dockerRemoteRepoPacker := packer.Universal(
		predicate.All(
			predicate.SchemaHasKey(dockerRemoteSchema),
			predicate.NoPassword,
			predicate.Ignore("external_dependencies_patterns"),
		),
	)

	constructor := func() (interface{}, error) {
		return &DockerRemoteRepository{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "remote",
				PackageType: packageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(dockerRemoteSchema, dockerRemoteRepoPacker, unpackDockerRemoteRepo, constructor)
}
