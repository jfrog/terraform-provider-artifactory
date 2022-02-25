package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArtifactoryGoVirtualRepository() *schema.Resource {

	const packageType = "go"

	var goVirtualSchema = mergeSchema(getBaseVirtualRepoSchema(packageType), map[string]*schema.Schema{

		"external_dependencies_enabled": {
			Type:        schema.TypeBool,
			Default:     true,
			Optional:    true,
			Description: "When set (default), Artifactory will automatically follow remote VCS roots in 'go-import' meta tags to download remote modules.",
		},
		"external_dependencies_patterns": {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			RequiredWith: []string{"external_dependencies_enabled"},
			Description: "An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will " +
				"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response. " +
				"By default, this is set to '**', which means that remote modules may be downloaded from any external VCS source.",
		},
	})

	type GoVirtualRepositoryParams struct {
		VirtualRepositoryBaseParams
		ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled,omitempty"`
		ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns,omitempty"`
	}

	var unpackGoVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}

		repo := GoVirtualRepositoryParams{
			VirtualRepositoryBaseParams:  unpackBaseVirtRepo(s, packageType),
			ExternalDependenciesPatterns: d.getList("external_dependencies_patterns"),
			ExternalDependenciesEnabled:  d.getBool("external_dependencies_enabled", false),
		}
		repo.PackageType = "go"
		return &repo, repo.Key, nil
	}

	return mkResourceSchema(goVirtualSchema, defaultPacker, unpackGoVirtualRepository, func() interface{} {
		return &GoVirtualRepositoryParams{
			VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: packageType,
			},
		}
	})

}
