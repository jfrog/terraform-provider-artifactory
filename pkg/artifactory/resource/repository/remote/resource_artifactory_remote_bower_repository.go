package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type BowerRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	BowerRegistryUrl string `json:"bowerRegistryUrl"`
}

func ResourceArtifactoryRemoteBowerRepository() *schema.Resource {
	const packageType = "bower"

	var bowerRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, VcsRemoteRepoSchema, map[string]*schema.Schema{
		"bower_registry_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://registry.bower.io",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `Proxy remote Bower repository. Default value is "https://registry.bower.io".`,
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackBowerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := BowerRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, packageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			BowerRegistryUrl:           d.GetString("bower_registry_url", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef("remote", packageType)()
		if err != nil {
			return nil, err
		}

		return &BowerRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	return repository.MkResourceSchema(bowerRemoteSchema, packer.Default(bowerRemoteSchema), unpackBowerRemoteRepo, constructor)
}
