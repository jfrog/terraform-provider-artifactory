package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type ComposerRemoteRepo struct {
	RemoteRepositoryBaseParams
	RemoteRepositoryVcsParams
	ComposerRegistryUrl string `json:"composerRegistryUrl"`
}

func resourceArtifactoryRemoteComposerRepository() *schema.Resource {
	const packageType = "composer"

	var composerRemoteSchema = mergeSchema(baseRemoteRepoSchema, vcsRemoteRepoSchema, map[string]*schema.Schema{
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
			RemoteRepositoryVcsParams:  unpackVcsRemoteRepo(s),
			ComposerRegistryUrl:        d.getString("composer_registry_url", false),
		}
		return repo, repo.Id(), nil
	}

	composerRemoteRepoPacker := universalPack(
		allHclPredicate(
			noPassword,
			schemaHasKey(composerRemoteSchema),
		),
	)

	return mkResourceSchema(composerRemoteSchema, composerRemoteRepoPacker, unpackComposerRemoteRepo, func() interface{} {
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
