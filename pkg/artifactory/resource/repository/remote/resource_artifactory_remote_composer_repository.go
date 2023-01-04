package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type ComposerRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	ComposerRegistryUrl string `json:"composerRegistryUrl"`
}

func ResourceArtifactoryRemoteComposerRepository() *schema.Resource {
	const packageType = "composer"

	var composerRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, VcsRemoteRepoSchema, map[string]*schema.Schema{
		"composer_registry_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://packagist.org",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `Proxy remote Composer repository. Default value is "https://packagist.org".`,
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackComposerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := ComposerRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, packageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			ComposerRegistryUrl:        d.GetString("composer_registry_url", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef("remote", packageType)()
		if err != nil {
			return nil, err
		}

		return &ComposerRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        "remote",
				PackageType:   packageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	return repository.MkResourceSchema(composerRemoteSchema, packer.Default(composerRemoteSchema), unpackComposerRemoteRepo, constructor)
}
