package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type ComposerRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	ComposerRegistryUrl string `json:"composerRegistryUrl"`
}

var composerSchema = lo.Assign(
	baseSchema,
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
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.ComposerPackageType),
)

var ComposerSchemas = GetSchemas(composerSchema)

func ResourceArtifactoryRemoteComposerRepository() *schema.Resource {
	var unpackComposerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := ComposerRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.ComposerPackageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			ComposerRegistryUrl:        d.GetString("composer_registry_url", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.ComposerPackageType)
		if err != nil {
			return nil, err
		}

		return &ComposerRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.ComposerPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	return mkResourceSchema(
		ComposerSchemas,
		packer.Default(ComposerSchemas[CurrentSchemaVersion]),
		unpackComposerRemoteRepo,
		constructor,
	)
}
