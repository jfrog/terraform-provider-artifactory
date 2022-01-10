package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var helmRemoteSchema = mergeSchema(baseRemoteSchema, map[string]*schema.Schema{
	"helm_charts_base_url": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.IsURLWithHTTPorHTTPS,
		Description: "Base URL for the translation of chart source URLs in the index.yaml of virtual repos. " +
			"Artifactory will only translate URLs matching the index.yamls hostname or URLs starting with this base url.",
	},
	"external_dependencies_enabled": {
		Type:        schema.TypeBool,
		Default:     false,
		Optional:    true,
		Description: "When set, external dependencies are rewritten.",
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
})

type HelmRemoteRepo struct {
	RemoteRepositoryBaseParams
	HelmChartsBaseURL            string   `hcl:"helm_charts_base_url" json:"chartsBaseUrl,omitempty"`
	ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled,omitempty"`
	ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns,omitempty"`
}

func resourceArtifactoryRemoteHelmRepository() *schema.Resource {
	return mkResourceSchema(helmRemoteSchema, defaultPacker, unpackhelmRemoteRepo, func() interface{} {
		return &HelmRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:      "remote",
				PackageType: "helm",
			},
		}
	})
}

func unpackhelmRemoteRepo(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	repo := HelmRemoteRepo{
		RemoteRepositoryBaseParams:   unpackBaseRemoteRepo(s, "helm"),
		HelmChartsBaseURL:            d.getString("helm_charts_base_url", false),
		ExternalDependenciesEnabled:  d.getBool("external_dependencies_enabled", false),
		ExternalDependenciesPatterns: d.getList("external_dependencies_patterns"),
	}
	return repo, repo.Id(), nil
}
