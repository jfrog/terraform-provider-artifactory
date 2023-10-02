package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type ConanRepo struct {
	RepositoryRemoteBaseParams
	repository.ConanBaseParams
}

var ConanSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		repository.ConanBaseSchema,
		repository.RepoLayoutRefSchema(rclass, repository.ConanPackageType),
	)
}

func ResourceArtifactoryRemoteConanRepository() *schema.Resource {
	var unpackConanRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := ConanRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.ConanPackageType),
			ConanBaseParams: repository.ConanBaseParams{
				EnableConanSupport:       true,
				ForceConanAuthentication: d.GetBool("force_conan_authentication", false),
			},
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, repository.ConanPackageType)()
		if err != nil {
			return nil, err
		}

		return &ConanRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   repository.ConanPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
			ConanBaseParams: repository.ConanBaseParams{
				EnableConanSupport: true,
			},
		}, nil
	}

	conanSchema := ConanSchema(true)

	return mkResourceSchema(
		conanSchema,
		packer.Default(conanSchema),
		unpackConanRepo,
		constructor,
	)
}
