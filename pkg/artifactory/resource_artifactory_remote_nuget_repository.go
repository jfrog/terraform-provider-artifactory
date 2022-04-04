package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type NugetRemoteRepo struct {
	RemoteRepositoryBaseParams
	FeedContextPath          string `json:"feedContextPath"`
	DownloadContextPath      string `json:"downloadContextPath"`
	V3FeedUrl                string `hcl:"v3_feed_url" json:"v3FeedUrl"` // Forced to specify hcl tag because predicate is not parsed by universalPack function.
	ForceNugetAuthentication bool   `json:"forceNugetAuthentication"`
}

func resourceArtifactoryRemoteNugetRepository() *schema.Resource {
	const packageType = "nuget"

	var nugetRemoteSchema = mergeSchema(baseRemoteRepoSchema, map[string]*schema.Schema{
		"feed_context_path": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "api/v2",
			Description: `(Optional) When proxying a remote NuGet repository, customize feed resource location using this attribute. Default value is 'api/v2'.`,
		},
		"download_context_path": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "api/v2/package",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      `(Optional) The context path prefix through which NuGet downloads are served. Default value is 'api/v2/package'.`,
		},
		"v3_feed_url": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "https://api.nuget.org/v3/index.json",
			ValidateDiagFunc: validation.ToDiagFunc(validation.Any(validation.IsURLWithHTTPorHTTPS, validation.StringIsEmpty)),
			Description:      `(Optional) The URL to the NuGet v3 feed. Default value is 'https://api.nuget.org/v3/index.json'.`,
		},
		"force_nuget_authentication": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `(Optional) Force basic authentication credentials in order to use this repository. Default value is 'false'.`,
		},
	}, repoLayoutRefSchema("remote", packageType))

	var unpackNugetRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}
		repo := NugetRemoteRepo{
			RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s, packageType),
			FeedContextPath:            d.getString("feed_context_path", false),
			DownloadContextPath:        d.getString("download_context_path", false),
			V3FeedUrl:                  d.getString("v3_feed_url", false),
			ForceNugetAuthentication:   d.getBool("force_nuget_authentication", false),
		}
		return repo, repo.Id(), nil
	}

	nugetRemoteRepoPacker := universalPack(nugetRemoteSchema, noPassword)

	return mkResourceSchema(nugetRemoteSchema, nugetRemoteRepoPacker, unpackNugetRemoteRepo, func() interface{} {
		repoLayout, _ := getDefaultRepoLayoutRef("remote", packageType)()
		return &NugetRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
