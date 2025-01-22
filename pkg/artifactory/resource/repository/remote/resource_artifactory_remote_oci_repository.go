package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var OCISchema = lo.Assign(
	BaseSchema,
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
		// We need to set default to ["**"] once we migrate to plugin-framework. SDKv2 doesn't support that.
		"external_dependencies_patterns": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			RequiredWith: []string{"external_dependencies_enabled"},
			Description: "Optional include patterns to match external URLs. Ant-style path expressions are supported (*, **, ?). " +
				"For example, specifying `**/github.com/**` will only allow downloading foreign layers from github.com host." +
				"By default, this is set to '**' in the UI, which means that foreign layers may be downloaded from any external host." +
				"Due to Terraform SDKv2 limitations, we can't set the default value for the list." +
				"This value must be assigned to the attribute manually, if user don't specify any other non-default values." +
				"This attribute must be set together with `external_dependencies_enabled = true`",
		},
		"project_id": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "Use this attribute to enter your GCR, GAR Project Id to limit the scope of this remote repo to a specific " +
				"project in your third-party registry. When leaving this field blank or unset, remote repositories that support project id " +
				"will default to their default project as you have set up in your account.",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.OCIPackageType),
)

var OCISchemas = GetSchemas(OCISchema)

type OciRemoteRepo struct {
	RepositoryRemoteBaseParams
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns,omitempty"`
	EnableTokenAuthentication    bool     `json:"enableTokenAuthentication"`
	ProjectId                    string   `json:"dockerProjectId"`
}

func ResourceArtifactoryRemoteOciRepository() *schema.Resource {
	var unpackOciRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := OciRemoteRepo{
			RepositoryRemoteBaseParams:   UnpackBaseRemoteRepo(s, repository.OCIPackageType),
			EnableTokenAuthentication:    d.GetBool("enable_token_authentication", false),
			ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
			ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
			ProjectId:                    d.GetString("project_id", false),
		}
		return repo, repo.Id(), nil
	}

	ociRemoteRepoPacker := packer.Universal(
		predicate.All(
			predicate.SchemaHasKey(OCISchemas[CurrentSchemaVersion]),
			predicate.NoPassword,
		),
	)

	constructor := func() (interface{}, error) {
		return &OciRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      Rclass,
				PackageType: repository.OCIPackageType,
			},
		}, nil
	}

	return mkResourceSchema(
		OCISchemas,
		ociRemoteRepoPacker,
		unpackOciRemoteRepo,
		constructor,
	)
}
