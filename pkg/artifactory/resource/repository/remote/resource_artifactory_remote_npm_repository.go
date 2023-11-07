package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type NpmRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryCurationParams
}

const NpmPackageType = "npm"

var NpmRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		CurationRemoteRepoSchema,
		repository.RepoLayoutRefSchema(rclass, NpmPackageType),
	)
}

func ResourceArtifactoryRemoteNpmRepository() *schema.Resource {
	var unpack = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := NpmRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, NpmPackageType),
			RepositoryCurationParams: RepositoryCurationParams{
				Curated: d.GetBool("curated", false),
			},
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, NpmPackageType)()
		if err != nil {
			return nil, err
		}

		return &NpmRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   NpmPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	npmSchema := NpmRemoteSchema(true)

	return mkResourceSchema(npmSchema, packer.Default(npmSchema), unpack, constructor)
}
