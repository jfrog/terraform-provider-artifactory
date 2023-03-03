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

	var helmRemoteSchema = util.MergeMaps(baseRemoteRepoSchemaV2, map[string]*schema.Schema{
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
				"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response. " +
				"By default, this is set to '[**]' in the UI, which means that remote modules may be downloaded from any external VCS source." +
				"Due to SDKv2 limitations, we can't set the default value for the list." +
				"This value [**] must be assigned to the attribute manually, if user don't specify any other non-default values." +
				"We don't want to make this attribute required, but it must be set to avoid the state drift on update. Note: Artifactory assigns " +
				"[**] on update if HCL doesn't have the attribute set or the list is empty.",
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	type HelmRemoteRepo struct {
		RepositoryRemoteBaseParams
		HelmChartsBaseURL            string   `hcl:"helm_charts_base_url" json:"chartsBaseUrl"`
		ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled"`
		ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns"`
	}

	var unpackHelmRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := HelmRemoteRepo{
			RepositoryRemoteBaseParams:   UnpackBaseRemoteRepo(s, packageType),
			HelmChartsBaseURL:            d.GetString("helm_charts_base_url", false),
			ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
			ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
		}
		return repo, repo.Id(), nil
	}

	helmRemoteRepoPacker := packer.Universal(
		predicate.All(
			predicate.SchemaHasKey(helmRemoteSchema),
			predicate.NoPassword,
		),
	)

	constructor := func() (interface{}, error) {
		return &HelmRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      "remote",
				PackageType: packageType,
			},
		}, nil
	}

	return mkResourceSchema(helmRemoteSchema, helmRemoteRepoPacker, unpackHelmRemoteRepo, constructor)
}
