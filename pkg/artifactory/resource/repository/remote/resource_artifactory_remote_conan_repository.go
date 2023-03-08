package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type ConanRemoteRepo struct {
	RepositoryRemoteBaseParams
	ForceConanAuthentication bool `json:"forceConanAuthentication"`
}

const ConanPackageType = "conan"

var ConanRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return util.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		map[string]*schema.Schema{
			"force_conan_authentication": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `Force basic authentication credentials in order to use this repository. Default value is 'false'.`,
			},
		},
		repository.RepoLayoutRefSchema(rclass, ConanPackageType),
	)
}

func ResourceArtifactoryRemoteConanRepository() *schema.Resource {
	var unpackConanRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := ConanRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, ConanPackageType),
			ForceConanAuthentication:   d.GetBool("force_conan_authentication", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, ConanPackageType)()
		if err != nil {
			return nil, err
		}

		return &ConanRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   ConanPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	conanSchema := ConanRemoteSchema(true)

	return mkResourceSchema(conanSchema, packer.Default(conanSchema), unpackConanRemoteRepo, constructor)
}
