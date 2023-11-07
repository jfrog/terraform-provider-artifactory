package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const DockerPackageType = "docker"

var DockerRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		CurationRemoteRepoSchema,
		map[string]*schema.Schema{
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
					"By default, this is set to '**' in the UI, which means that remote modules may be downloaded from any external VCS source." +
					"Due to SDKv2 limitations, we can't set the default value for the list." +
					"This value must be assigned to the attribute manually, if user don't specify any other non-default values." +
					"This attribute must be set together with `external_dependencies_enabled = true`",
			},
		},
		repository.RepoLayoutRefSchema(rclass, DockerPackageType),
	)
}

type DockerRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryCurationParams
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns,omitempty"`
	EnableTokenAuthentication    bool     `json:"enableTokenAuthentication"`
	BlockPushingSchema1          bool     `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
}

func ResourceArtifactoryRemoteDockerRepository() *schema.Resource {
	var unpackDockerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := DockerRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, DockerPackageType),
			RepositoryCurationParams: RepositoryCurationParams{
				Curated: d.GetBool("curated", false),
			},
			EnableTokenAuthentication:    d.GetBool("enable_token_authentication", false),
			ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
			BlockPushingSchema1:          d.GetBool("block_pushing_schema1", false),
			ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
		}
		return repo, repo.Id(), nil
	}

	dockerSchema := DockerRemoteSchema(true)

	dockerRemoteRepoPacker := packer.Universal(
		predicate.All(
			predicate.SchemaHasKey(dockerSchema),
			predicate.NoPassword,
		),
	)

	constructor := func() (interface{}, error) {
		return &DockerRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      rclass,
				PackageType: DockerPackageType,
			},
		}, nil
	}

	return mkResourceSchema(dockerSchema, dockerRemoteRepoPacker, unpackDockerRemoteRepo, constructor)
}
