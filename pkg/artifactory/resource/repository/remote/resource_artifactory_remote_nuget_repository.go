package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type NugetRemoteRepo struct {
	RepositoryBaseParams
	FeedContextPath          string `json:"feedContextPath"`
	DownloadContextPath      string `json:"downloadContextPath"`
	V3FeedUrl                string `hcl:"v3_feed_url" json:"v3FeedUrl"` // Forced to specify hcl tag because predicate is not parsed by packer.Universal function.
	ForceNugetAuthentication bool   `json:"forceNugetAuthentication"`
}

func ResourceArtifactoryRemoteNugetRepository() *schema.Resource {
	const packageType = "nuget"

	var nugetRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, map[string]*schema.Schema{
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
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackNugetRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := NugetRemoteRepo{
			RepositoryBaseParams:     UnpackBaseRemoteRepo(s, packageType),
			FeedContextPath:          d.GetString("feed_context_path", false),
			DownloadContextPath:      d.GetString("download_context_path", false),
			V3FeedUrl:                d.GetString("v3_feed_url", false),
			ForceNugetAuthentication: d.GetBool("force_nuget_authentication", false),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(nugetRemoteSchema, packer.Default(nugetRemoteSchema), unpackNugetRemoteRepo, func() interface{} {
		repoLayout, _ := repository.GetDefaultRepoLayoutRef("remote", packageType)()
		return &NugetRemoteRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
