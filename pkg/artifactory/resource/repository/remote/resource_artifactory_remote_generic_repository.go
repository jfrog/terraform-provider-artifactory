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

func ResourceArtifactoryRemoteGenericRepository() *schema.Resource {
	const packageType = "generic"

	var genericRemoteSchema = util.MergeMaps(baseRemoteRepoSchemaV2, map[string]*schema.Schema{
		"propagate_query_params": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.",
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackGenericRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := GenericRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, packageType),
			PropagateQueryParams:       d.GetBool("propagate_query_params", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef("remote", packageType)()
		if err != nil {
			return nil, err
		}

		return &GenericRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        "remote",
				PackageType:   packageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	return mkResourceSchema(genericRemoteSchema, packer.Default(genericRemoteSchema), unpackGenericRemoteRepo, constructor)
}

func ResourceArtifactoryRemoteBasicRepository(pkt string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef("remote", pkt)()
		if err != nil {
			return nil, err
		}

		return &RepositoryRemoteBaseParams{
			PackageType:   pkt,
			Rclass:        "remote",
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRemoteRepo(data, pkt)
		return repo, repo.Id(), nil
	}

	mergedRemoteRepoSchema := util.MergeMaps(baseRemoteRepoSchema, repository.RepoLayoutRefSchema("remote", pkt))

	return mkResourceSchema(mergedRemoteRepoSchema, packer.Default(mergedRemoteRepoSchema), unpack, constructor)
}
