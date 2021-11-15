package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"

	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
)

type GoVirtualRepositoryParams struct {
	RepositoryBaseParams
	ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled,omitempty"`
	ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns,omitempty"`
}

var goVirtualSchema = util.MergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{

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

func ResourceArtifactoryGoVirtualRepository() *schema.Resource {
	return repos.MkResourceSchema(goVirtualSchema, util.DefaultPacker, unpackGoVirtualRepository, func() interface{} {
		return &GoVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: "go",
			},
		}
	})

}

func unpackGoVirtualRepository(s *schema.ResourceData) (interface{}, string, error) {
	d := &util.ResourceData{s}

	repo := GoVirtualRepositoryParams{
		RepositoryBaseParams:         unpackBaseVirtRepo(s),
		ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
		ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
	}
	repo.PackageType = "go"
	return &repo, repo.Key, nil
}
