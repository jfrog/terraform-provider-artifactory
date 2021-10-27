package artifactory

import (
	"fmt"
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
	HelmChartsBaseURL string `json:"chartsBaseUrl,omitempty"`
}

var helmRemoteRepoReadFun = mkRepoRead(packhelmRemoteRepo, func() interface{} {
	return &HelmRemoteRepo{
		RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
			Rclass:      "remote",
			PackageType: "helm",
		},
	}
})

func resourceArtifactoryRemoteHelmRepository() *schema.Resource {
	return &schema.Resource{
		Create: mkRepoCreate(unpackhelmRemoteRepo, helmRemoteRepoReadFun),
		Read:   helmRemoteRepoReadFun,
		Update: mkRepoUpdate(unpackhelmRemoteRepo, helmRemoteRepoReadFun),
		Delete: deleteRepo,
		Exists: repoExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: helmRemoteSchema,
	}
}

func unpackhelmRemoteRepo(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	repo := HelmRemoteRepo{
		RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s),
		HelmChartsBaseURL:          d.getString("helm_charts_base_url", false),
	}
	repo.PackageType = "helm"
	return repo, repo.Key, nil
}

func packhelmRemoteRepo(r interface{}, d *schema.ResourceData) error {
	repo := r.(*HelmRemoteRepo)
	setValue := packBaseRemoteRepo(d, repo.RemoteRepositoryBaseParams)
	errors := setValue("helm_charts_base_url", repo.HelmChartsBaseURL)

	if len(errors) > 0 {
		return fmt.Errorf("%q", errors)
	}

	return nil
}
