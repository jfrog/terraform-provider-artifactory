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
})

type HelmRemoteRepo struct {
	RemoteRepositoryBaseParams
	HelmChartsBaseURL string `hcl:"helm_charts_base_url" json:"chartsBaseUrl,omitempty"`
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
		RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s, "helm"),
		HelmChartsBaseURL:          d.getString("helm_charts_base_url", false),
	}
	return repo, repo.Id(), nil
}
