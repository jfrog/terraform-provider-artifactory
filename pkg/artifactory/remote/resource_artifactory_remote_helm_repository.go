package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
)

var helmRemoteSchema = util.MergeSchema(baseRemoteSchema, map[string]*schema.Schema{
	"helm_charts_base_url": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.IsURLWithHTTPorHTTPS,
		Description: "Base URL for the translation of chart source URLs in the index.yaml of virtual  " +
			"Artifactory will only translate URLs matching the index.yamls hostname or URLs starting with this base url.",
	},
})

type HelmRemoteRepo struct {
	RepositoryBaseParams
	HelmChartsBaseURL string `hcl:"helm_charts_base_url" json:"chartsBaseUrl,omitempty"`
}

func ResourceArtifactoryRemoteHelmRepository() *schema.Resource {
	return repos.MkResourceSchema(helmRemoteSchema, util.DefaultPacker, unpackhelmRemoteRepo, func() interface{} {
		return &HelmRemoteRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "remote",
				PackageType: "helm",
			},
		}
	})
}

func unpackhelmRemoteRepo(s *schema.ResourceData) (interface{}, string, error) {
	d := &util.ResourceData{ResourceData: s}
	repo := HelmRemoteRepo{
		RepositoryBaseParams: unpackBaseRemoteRepo(s, "helm"),
		HelmChartsBaseURL:    d.GetString("helm_charts_base_url", false),
	}
	return repo, repo.Id(), nil
}
