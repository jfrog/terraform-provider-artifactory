package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var HelmSchema = lo.Assign(
	baseSchema,
	map[string]*schema.Schema{
		"helm_charts_base_url": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.Any(
					validation.IsURLWithScheme([]string{"http", "https", "oci"}),
					validation.StringIsEmpty,
				),
			),
			Description: "Base URL for the translation of chart source URLs in the index.yaml of virtual repos. " +
				"Artifactory will only translate URLs matching the index.yamls hostname or URLs starting with this base url. " +
				"Support http/https/oci protocol scheme.",
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
	repository.RepoLayoutRefSchema(Rclass, repository.HelmPackageType),
)

var HelmSchemas = GetSchemas(HelmSchema)

type HelmRemoteRepo struct {
	RepositoryRemoteBaseParams
	HelmChartsBaseURL            string   `hcl:"helm_charts_base_url" json:"chartsBaseUrl"`
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns"`
}

func ResourceArtifactoryRemoteHelmRepository() *schema.Resource {
	var unpackHelmRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := HelmRemoteRepo{
			RepositoryRemoteBaseParams:   UnpackBaseRemoteRepo(s, repository.HelmPackageType),
			HelmChartsBaseURL:            d.GetString("helm_charts_base_url", false),
			ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
			ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
		}
		return repo, repo.Id(), nil
	}

	helmRemoteRepoPacker := packer.Universal(
		predicate.All(
			predicate.SchemaHasKey(HelmSchemas[CurrentSchemaVersion]),
			predicate.NoPassword,
		),
	)

	constructor := func() (interface{}, error) {
		return &HelmRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      Rclass,
				PackageType: repository.HelmPackageType,
			},
		}, nil
	}

	return mkResourceSchema(
		HelmSchemas,
		helmRemoteRepoPacker,
		unpackHelmRemoteRepo,
		constructor,
	)
}
