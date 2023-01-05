package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type VcsRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	MaxUniqueSnapshots int `json:"maxUniqueSnapshots"`
}

func ResourceArtifactoryRemoteVcsRepository() *schema.Resource {
	const packageType = "vcs"

	var vcsRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, VcsRemoteRepoSchema, map[string]*schema.Schema{
		"max_unique_snapshots": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			Description: "The maximum number of unique snapshots of a single artifact to store. Once the number of " +
				"snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is " +
				"no limit, and unique snapshots are not cleaned up.",
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var UnpackVcsRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := VcsRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, packageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			MaxUniqueSnapshots:         d.GetInt("max_unique_snapshots", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef("remote", packageType)()
		if err != nil {
			return nil, err
		}

		return &VcsRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        "remote",
				PackageType:   packageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	return repository.MkResourceSchema(vcsRemoteSchema, packer.Default(vcsRemoteSchema), UnpackVcsRemoteRepo, constructor)
}
