package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

const HelmPackageType = "helm"

var HelmRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return util.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		map[string]*schema.Schema{
			"helm_charts_base_url": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "",
				ValidateDiagFunc: validation.ToDiagFunc(validation.Any(validation.IsURLWithHTTPorHTTPS, validation.StringIsEmpty)),
				Description: "Base URL for the translation of chart source URLs in the index.yaml of virtual repos. " +
					"Artifactory will only translate URLs matching the index.yamls hostname or URLs starting with this base url.",
			},
			"external_dependencies_enabled": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "When set, external dependencies are rewritten. External Dependency Rewrite in the UI.",
			},
			// We need to set default to ["**"] once we migrate to plugin-framework. SDKv2 doesn't support that.
			"external_dependencies_patterns": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				RequiredWith: []string{"external_dependencies_enabled"},
				Description: "An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will " +
					"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response." +
					"Default value in UI is empty. This attribute must be set together with `external_dependencies_enabled = true`",
			},
		},
		repository.RepoLayoutRefSchema(rclass, HelmPackageType),
	)
}

type HelmRemoteRepo struct {
	RepositoryRemoteBaseParams
	HelmChartsBaseURL            string   `hcl:"helm_charts_base_url" json:"chartsBaseUrl"`
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns"`
}

func ResourceArtifactoryRemoteHelmRepository() *schema.Resource {
	var unpackHelmRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := HelmRemoteRepo{
			RepositoryRemoteBaseParams:   UnpackBaseRemoteRepo(s, HelmPackageType),
			HelmChartsBaseURL:            d.GetString("helm_charts_base_url", false),
			ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
			ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
		}
		return repo, repo.Id(), nil
	}

	helmSchema := HelmRemoteSchema(true)

	helmRemoteRepoPacker := packer.Universal(
		predicate.All(
			predicate.SchemaHasKey(helmSchema),
			predicate.NoPassword,
		),
	)

	constructor := func() (interface{}, error) {
		return &HelmRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      rclass,
				PackageType: HelmPackageType,
			},
		}, nil
	}

	return mkResourceSchema(helmSchema, helmRemoteRepoPacker, unpackHelmRemoteRepo, constructor)
}
