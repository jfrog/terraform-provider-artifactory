package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type GenericRemoteRepo struct {
	RepositoryRemoteBaseParams
	PropagateQueryParams bool `json:"propagateQueryParams"`
}

const GenericPackageType = "generic"

var GenericRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	genericSchema := util.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		map[string]*schema.Schema{
			"propagate_query_params": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.",
			},
		},
		repository.RepoLayoutRefSchema(rclass, GenericPackageType),
	)

	return genericSchema
}

func ResourceArtifactoryRemoteGenericRepository() *schema.Resource {

	var unpackGenericRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := GenericRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, GenericPackageType),
			PropagateQueryParams:       d.GetBool("propagate_query_params", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, GenericPackageType)()
		if err != nil {
			return nil, err
		}

		return &GenericRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   GenericPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	genericSchema := GenericRemoteSchema(true)

	return mkResourceSchema(genericSchema, packer.Default(genericSchema), unpackGenericRemoteRepo, constructor)
}

var BasicRepoSchema = func(packageType string, isResource bool) map[string]*schema.Schema {
	return util.MergeMaps(BaseRemoteRepoSchema(isResource), repository.RepoLayoutRefSchema(rclass, packageType))
}

func ResourceArtifactoryRemoteBasicRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, packageType)()
		if err != nil {
			return nil, err
		}

		return &RepositoryRemoteBaseParams{
			PackageType:   packageType,
			Rclass:        rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRemoteRepo(data, packageType)
		return repo, repo.Id(), nil
	}

	mergedRemoteRepoSchema := BasicRepoSchema(packageType, true)

	return mkResourceSchema(mergedRemoteRepoSchema, packer.Default(mergedRemoteRepoSchema), unpack, constructor)
}
