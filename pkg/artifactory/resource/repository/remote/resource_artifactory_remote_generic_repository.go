package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type GenericRemoteRepo struct {
	RepositoryRemoteBaseParams
	PropagateQueryParams bool `json:"propagateQueryParams"`
}

var genericSchema = lo.Assign(
	baseSchema,
	map[string]*schema.Schema{
		"propagate_query_params": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.",
		},
	},
	repository.RepoLayoutRefSchema(Rclass, repository.GenericPackageType),
)

var GenericSchemas = GetSchemas(genericSchema)

func ResourceArtifactoryRemoteGenericRepository() *schema.Resource {

	var unpackGenericRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := GenericRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.GenericPackageType),
			PropagateQueryParams:       d.GetBool("propagate_query_params", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.GenericPackageType)()
		if err != nil {
			return nil, err
		}

		return &GenericRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.GenericPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	return mkResourceSchema(
		GenericSchemas,
		packer.Default(GenericSchemas[CurrentSchemaVersion]),
		unpackGenericRemoteRepo,
		constructor,
	)
}

var BasicSchema = func(packageType string) map[string]*schema.Schema {
	return lo.Assign(
		baseSchema,
		repository.RepoLayoutRefSchema(Rclass, packageType),
	)
}

func ResourceArtifactoryRemoteBasicRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, packageType)()
		if err != nil {
			return nil, err
		}

		return &RepositoryRemoteBaseParams{
			PackageType:   packageType,
			Rclass:        Rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRemoteRepo(data, packageType)
		return repo, repo.Id(), nil
	}

	basicSchema := BasicSchema(packageType)
	basicSchemas := GetSchemas(basicSchema)

	return mkResourceSchema(
		basicSchemas,
		packer.Default(basicSchemas[CurrentSchemaVersion]),
		unpack,
		constructor,
	)
}
