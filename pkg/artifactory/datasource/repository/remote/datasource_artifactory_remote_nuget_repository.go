package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

type NugetRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryCurationParams
	FeedContextPath          string `json:"feedContextPath"`
	DownloadContextPath      string `json:"downloadContextPath"`
	V3FeedUrl                string `hcl:"v3_feed_url" json:"v3FeedUrl"` // Forced to specify hcl tag because predicate is not parsed by packer.Universal function.
	ForceNugetAuthentication bool   `json:"forceNugetAuthentication"`
	SymbolServerUrl          string `json:"symbolServerUrl"`
}

var NugetSchema = lo.Assign(
	remote.BaseSchema,
	remote.CurationRemoteRepoSchema,
	map[string]*schema.Schema{
		"feed_context_path": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "api/v2",
			Description: `When proxying a remote NuGet repository, customize feed resource location using this attribute. Default value is 'api/v2'.`,
		},
		"download_context_path": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "api/v2/package",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      `The context path prefix through which NuGet downloads are served. Default value is 'api/v2/package'.`,
		},
		"v3_feed_url": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "https://api.nuget.org/v3/index.json",
			ValidateDiagFunc: validation.ToDiagFunc(validation.Any(validation.IsURLWithHTTPorHTTPS, validation.StringIsEmpty)),
			Description:      `The URL to the NuGet v3 feed. Default value is 'https://api.nuget.org/v3/index.json'.`,
		},
		"force_nuget_authentication": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `Force basic authentication credentials in order to use this repository. Default value is 'false'.`,
		},
		"symbol_server_url": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "https://symbols.nuget.org/download/symbols",
			ValidateDiagFunc: validation.ToDiagFunc(validation.Any(validation.IsURLWithHTTPorHTTPS, validation.StringIsEmpty)),
			Description:      `NuGet symbol server URL.`,
		},
	}, resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.NugetPackageType),
)

var NugetSchemas = remote.GetSchemas(NugetSchema)

func DataSourceArtifactoryRemoteNugetRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.NugetPackageType)
		if err != nil {
			return nil, err
		}

		return &NugetRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.NugetPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	nugetSchema := getSchema(NugetSchemas)

	return &schema.Resource{
		Schema:      nugetSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(nugetSchema), constructor),
		Description: "Provides a data source for a remote NuGet repository",
	}
}
