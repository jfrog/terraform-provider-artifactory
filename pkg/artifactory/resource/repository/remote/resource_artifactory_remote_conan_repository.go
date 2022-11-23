package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type ConanRemoteRepo struct {
	RepositoryBaseParams
	ForceConanAuthentication bool `json:"forceConanAuthentication"`
}

func ResourceArtifactoryRemoteConanRepository() *schema.Resource {
	const packageType = "conan"

	var conanRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, map[string]*schema.Schema{
		"force_conan_authentication": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `Force basic authentication credentials in order to use this repository. Default value is 'false'.`,
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackConanRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := ConanRemoteRepo{
			RepositoryBaseParams:     UnpackBaseRemoteRepo(s, packageType),
			ForceConanAuthentication: d.GetBool("force_conan_authentication", false),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(conanRemoteSchema, packer.Default(conanRemoteSchema), unpackConanRemoteRepo, func() interface{} {
		repoLayout, _ := repository.GetDefaultRepoLayoutRef("remote", packageType)()
		return &ConanRemoteRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
