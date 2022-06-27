package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryRemoteHelmRepository() *schema.Resource {
	const packageType = "helm"

	var helmRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, map[string]*schema.Schema{
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
	}, repository.RepoLayoutRefSchema("remote", packageType))

	type HelmRemoteRepo struct {
		RepositoryBaseParams
		HelmChartsBaseURL            string   `hcl:"helm_charts_base_url" json:"chartsBaseUrl"`
		ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled"`
		ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns"`
	}

	var unpackHelmRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := HelmRemoteRepo{
			RepositoryBaseParams:         UnpackBaseRemoteRepo(s, packageType),
			HelmChartsBaseURL:            d.GetString("helm_charts_base_url", false),
			ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
			ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
		}
		if len(repo.ExternalDependenciesPatterns) == 0 {
			repo.ExternalDependenciesPatterns = []string{"**"}
		}
		return repo, repo.Id(), nil
	}

	// Special handling for "external_dependencies_patterns" attribute to match default value behavior in UI.
	helmRemoteRepoPacker := packer.Universal(
		predicate.All(
			predicate.SchemaHasKey(helmRemoteSchema),
			predicate.NoPassword,
			predicate.Ignore("external_dependencies_patterns"),
		),
	)

	return repository.MkResourceSchema(helmRemoteSchema, helmRemoteRepoPacker, unpackHelmRemoteRepo, func() interface{} {
		return &HelmRemoteRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "remote",
				PackageType: packageType,
			},
		}
	})
}
