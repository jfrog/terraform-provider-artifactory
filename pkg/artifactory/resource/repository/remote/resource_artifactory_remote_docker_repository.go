package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryRemoteDockerRepository() *schema.Resource {
	const packageType = "docker"

	var dockerRemoteSchema = util.MergeMaps(baseRemoteRepoSchemaV2, map[string]*schema.Schema{
		"external_dependencies_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Also known as 'Foreign Layers Caching' on the UI, default is `false`.",
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
		// We need to set default to ["**"] once we migrate to plugin-framework. SDKv2 doesn't support that.
		"external_dependencies_patterns": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			RequiredWith: []string{"external_dependencies_enabled"},
			Description: "An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will " +
				"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response. " +
				"By default, this is set to '[**]' in the UI, which means that remote modules may be downloaded from any external VCS source." +
				"Due to SDKv2 limitations, we can't set the default value for the list." +
				"This value [**] must be assigned to the attribute manually, if user don't specify any other non-default values." +
				"We don't want to make this attribute required, but it must be set to avoid the state drift on update. Note: Artifactory assigns " +
				"[**] on update if HCL doesn't have the attribute set or the list is empty.",
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	type DockerRemoteRepository struct {
		RepositoryRemoteBaseParams
		ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled"`
		ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns"`
		EnableTokenAuthentication    bool     `hcl:"enable_token_authentication" json:"enableTokenAuthentication"`
		BlockPushingSchema1          bool     `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
	}

	var unpackDockerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := DockerRemoteRepository{
			RepositoryRemoteBaseParams:   UnpackBaseRemoteRepo(s, packageType),
			EnableTokenAuthentication:    d.GetBool("enable_token_authentication", false),
			ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
			BlockPushingSchema1:          d.GetBool("block_pushing_schema1", false),
			ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
		}
		return repo, repo.Id(), nil
	}

	dockerRemoteRepoPacker := packer.Universal(
		predicate.All(
			predicate.SchemaHasKey(dockerRemoteSchema),
			predicate.NoPassword,
		),
	)

	constructor := func() (interface{}, error) {
		return &DockerRemoteRepository{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      "remote",
				PackageType: packageType,
			},
		}, nil
	}

	return mkResourceSchema(dockerRemoteSchema, dockerRemoteRepoPacker, unpackDockerRemoteRepo, constructor)
}
