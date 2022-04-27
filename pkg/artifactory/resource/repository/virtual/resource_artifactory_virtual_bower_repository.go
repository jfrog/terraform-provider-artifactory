package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func ResourceArtifactoryVirtualBowerRepository() *schema.Resource {

	const packageType = "bower"

	var bowerVirtualSchema = utils.MergeSchema(BaseVirtualRepoSchema, map[string]*schema.Schema{
		"external_dependencies_enabled": {
			Type:        schema.TypeBool,
			Default:     false,
			Optional:    true,
			Description: "When set, external dependencies are rewritten. Default value is false.",
		},
		"external_dependencies_remote_repo": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			RequiredWith:     []string{"external_dependencies_enabled"},
			Description:      "The remote repository aggregated by this virtual repository in which the external dependency will be cached.",
		},
		"external_dependencies_patterns": {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			RequiredWith: []string{"external_dependencies_enabled"},
			Description: "An Allow List of Ant-style path expressions that specify where external dependencies may be downloaded from. " +
				"By default, this is set to ** which means that dependencies may be downloaded from any external source.",
		},
	}, repository.RepoLayoutRefSchema("virtual", packageType))

	type BowerVirtualRepositoryParams struct {
		VirtualRepositoryBaseParams
		ExternalDependenciesEnabled    bool     `json:"externalDependenciesEnabled"`
		ExternalDependenciesRemoteRepo string   `json:"externalDependenciesRemoteRepo"`
		ExternalDependenciesPatterns   []string `json:"externalDependenciesPatterns"`
	}

	var unpackBowerVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utils.ResourceData{s}

		repo := BowerVirtualRepositoryParams{
			VirtualRepositoryBaseParams:    UnpackBaseVirtRepo(s, packageType),
			ExternalDependenciesEnabled:    d.GetBool("external_dependencies_enabled", false),
			ExternalDependenciesRemoteRepo: d.GetString("external_dependencies_remote_repo", false),
			ExternalDependenciesPatterns:   d.GetList("external_dependencies_patterns"),
		}
		repo.PackageType = packageType
		return &repo, repo.Key, nil
	}

	return repository.MkResourceSchema(bowerVirtualSchema, repository.DefaultPacker(bowerVirtualSchema), unpackBowerVirtualRepository, func() interface{} {
		return &BowerVirtualRepositoryParams{
			VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: packageType,
			},
		}
	})
}
