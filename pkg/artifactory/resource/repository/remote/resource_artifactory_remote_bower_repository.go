package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

type BowerRemoteRepo struct {
	RemoteRepositoryBaseParams
	RemoteRepositoryVcsParams
	BowerRegistryUrl string `json:"bowerRegistryUrl"`
}

func ResourceArtifactoryRemoteBowerRepository() *schema.Resource {
	const packageType = "bower"

	var bowerRemoteSchema = utils.MergeSchema(BaseRemoteRepoSchema, VcsRemoteRepoSchema, map[string]*schema.Schema{
		"bower_registry_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://registry.bower.io",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `Proxy remote Bower repository. Default value is "https://registry.bower.io".`,
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackBowerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utils.ResourceData{s}
		repo := BowerRemoteRepo{
			RemoteRepositoryBaseParams: UnpackBaseRemoteRepo(s, packageType),
			RemoteRepositoryVcsParams:  UnpackVcsRemoteRepo(s),
			BowerRegistryUrl:           d.GetString("bower_registry_url", false),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(bowerRemoteSchema, repository.DefaultPacker(bowerRemoteSchema), unpackBowerRemoteRepo, func() interface{} {
		repoLayout, _ := utils.GetDefaultRepoLayoutRef("remote", packageType)()
		return &BowerRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
