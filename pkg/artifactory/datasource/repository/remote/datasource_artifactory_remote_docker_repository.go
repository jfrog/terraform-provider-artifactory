package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

// SDKv2
var dockerSchema = lo.Assign(
	remote.BaseSchema,
	remote.CurationRemoteRepoSchema,
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
		"project_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Use this attribute to enter your GCR, GAR Project Id to limit the scope of this remote repo to a specific project in your third-party registry. When leaving this field blank or unset, remote repositories that support project id will default to their default project as you have set up in your account.",
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.DockerPackageType),
)

var DockerSchemas = remote.GetSchemas(dockerSchema)

type DockerRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryCurationParams
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns,omitempty"`
	EnableTokenAuthentication    bool     `json:"enableTokenAuthentication"`
	BlockPushingSchema1          bool     `hcl:"block_pushing_schema1" json:"blockPushingSchema1"`
	ProjectId                    string   `json:"dockerProjectId"`
}

func DataSourceArtifactoryRemoteDockerRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.DockerPackageType)
		if err != nil {
			return nil, err
		}

		return &DockerRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.DockerPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	dockerSchema := getSchema(DockerSchemas)

	return &schema.Resource{
		Schema:      dockerSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(dockerSchema), constructor),
		Description: "Provides a data source for a remote Docker repository",
	}
}
