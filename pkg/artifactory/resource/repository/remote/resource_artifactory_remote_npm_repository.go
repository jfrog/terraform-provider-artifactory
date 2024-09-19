package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type NpmRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryCurationParams
}

var NPMSchema = lo.Assign(
	baseSchema,
	CurationRemoteRepoSchema,
	repository.RepoLayoutRefSchema(Rclass, repository.NPMPackageType),
)

var NPMSchemas = GetSchemas(NPMSchema)

func ResourceArtifactoryRemoteNpmRepository() *schema.Resource {
	var unpack = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := NpmRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.NPMPackageType),
			RepositoryCurationParams: RepositoryCurationParams{
				Curated: d.GetBool("curated", false),
			},
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.NPMPackageType)()
		if err != nil {
			return nil, err
		}

		return &NpmRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.NPMPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	return mkResourceSchema(
		NPMSchemas,
		packer.Default(NPMSchemas[CurrentSchemaVersion]),
		unpack,
		constructor,
	)
}
