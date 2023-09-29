package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type ComposerRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	ComposerRegistryUrl string `json:"composerRegistryUrl"`
}

const ComposerPackageType = "composer"

var ComposerRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		VcsRemoteRepoSchema,
		map[string]*schema.Schema{
			"composer_registry_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "https://packagist.org",
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Description:  `Proxy remote Composer repository. Default value is "https://packagist.org".`,
			},
		},
		repository.RepoLayoutRefSchema(rclass, ComposerPackageType),
	)
}

func ResourceArtifactoryRemoteComposerRepository() *schema.Resource {
	var unpackComposerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := ComposerRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, ComposerPackageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			ComposerRegistryUrl:        d.GetString("composer_registry_url", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, ComposerPackageType)()
		if err != nil {
			return nil, err
		}

		return &ComposerRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   ComposerPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	composerSchema := ComposerRemoteSchema(true)

	return mkResourceSchema(composerSchema, packer.Default(composerSchema), unpackComposerRemoteRepo, constructor)
}
