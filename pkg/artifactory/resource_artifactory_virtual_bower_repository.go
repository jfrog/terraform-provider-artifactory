package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceArtifactoryBowerVirtualRepository() *schema.Resource {

	const packageType = "bower"

	var bowerVirtualSchema = mergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{
		"external_dependencies_enabled": {
			Type:        schema.TypeBool,
			Default:     false,
			Optional:    true,
			Description: "(Optional) When set, external dependencies are rewritten. Default value is false.",
		},
		"external_dependencies_remote_repo": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			RequiredWith:     []string{"external_dependencies_enabled"},
			Description:      "(Optional) The remote repository aggregated by this virtual repository in which the external dependency will be cached.",
		},
		"external_dependencies_patterns": {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			RequiredWith: []string{"external_dependencies_enabled"},
			Description: "(Optional) An Allow List of Ant-style path expressions that specify where external dependencies may be downloaded from. " +
				"By default, this is set to ** which means that dependencies may be downloaded from any external source.",
		},
	}, repoLayoutRefSchema("virtual", packageType))

	type BowerVirtualRepositoryParams struct {
		VirtualRepositoryBaseParams
		ExternalDependenciesEnabled    bool     `json:"externalDependenciesEnabled"`
		ExternalDependenciesRemoteRepo string   `json:"externalDependenciesRemoteRepo"`
		ExternalDependenciesPatterns   []string `json:"externalDependenciesPatterns"`
	}

	var unpackBowerVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}

		repo := BowerVirtualRepositoryParams{
			VirtualRepositoryBaseParams:    unpackBaseVirtRepo(s, packageType),
			ExternalDependenciesEnabled:    d.getBool("external_dependencies_enabled", false),
			ExternalDependenciesRemoteRepo: d.getString("external_dependencies_remote_repo", false),
			ExternalDependenciesPatterns:   d.getList("external_dependencies_patterns"),
		}
		repo.PackageType = packageType
		return &repo, repo.Key, nil
	}

	return mkResourceSchema(bowerVirtualSchema, defaultPacker, unpackBowerVirtualRepository, func() interface{} {
		return &BowerVirtualRepositoryParams{
			VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: packageType,
			},
		}
	})
}
