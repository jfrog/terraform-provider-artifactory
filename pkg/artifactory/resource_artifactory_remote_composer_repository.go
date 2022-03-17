package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type ComposerRemoteRepo struct {
	RemoteRepositoryBaseParams
	VcsType             string `json:"vcsType"`
	VcsGitProvider      string `json:"vcsGitProvider"`
	VcsGitDownloadUrl   string `json:"vcsGitDownloadUrl"`
	ComposerRegistryUrl string `json:"composerRegistryUrl"`
}

func resourceArtifactoryRemoteComposerRepository() *schema.Resource {
	const packageType = "composer"

	var composerRemoteSchema = mergeSchema(baseRemoteRepoSchema, map[string]*schema.Schema{
		"vcs_type": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "GIT",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"GIT"}, false)),
			Description:      `(Optional) Artifactory supports proxying the Git providers. Default value is "GIT".`,
		},
		"vcs_git_provider": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "ARTIFACTORY",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"GITHUB", "BITBUCKET", "OLDSTASH", "STASH", "ARTIFACTORY", "CUSTOM"}, false)),
			Description:      `(Optional) Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is "ARTIFACTORY".`,
		},
		"vcs_git_download_url": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.All(validation.StringIsNotEmpty, validation.IsURLWithHTTPorHTTPS)),
			Description:      `(Optional) This attribute is used when vcs_git_provider is set to 'CUSTOM'. Provided URL will be used as proxy.`,
		},
		"composer_registry_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://packagist.org",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `(Optional) Proxy remote Composer repository. Default value is "https://packagist.org".`,
		},
	}, repoLayoutRefSchema("remote", packageType))

	var unpackComposerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}
		repo := ComposerRemoteRepo{
			RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s, packageType),
			VcsType:                    d.getString("vcs_type", false),
			VcsGitProvider:             d.getString("vcs_git_provider", false),
			VcsGitDownloadUrl:          d.getString("vcs_git_download_url", false),
			ComposerRegistryUrl:        d.getString("composer_registry_url", false),
		}
		return repo, repo.Id(), nil
	}

	return mkResourceSchema(composerRemoteSchema, defaultPacker, unpackComposerRemoteRepo, func() interface{} {
		repoLayout, _ := getDefaultRepoLayoutRef("remote", packageType)()
		return &ComposerRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
