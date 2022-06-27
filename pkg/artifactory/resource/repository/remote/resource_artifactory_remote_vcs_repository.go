package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type VcsRemoteRepo struct {
	RemoteRepositoryBaseParams
	RemoteRepositoryVcsParams
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
		d := &util.ResourceData{s}
		repo := VcsRemoteRepo{
			RemoteRepositoryBaseParams: UnpackBaseRemoteRepo(s, packageType),
			RemoteRepositoryVcsParams:  UnpackVcsRemoteRepo(s),
			MaxUniqueSnapshots:         d.GetInt("max_unique_snapshots", false),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(vcsRemoteSchema, packer.Default(vcsRemoteSchema), UnpackVcsRemoteRepo, func() interface{} {
		repoLayout, _ := repository.GetDefaultRepoLayoutRef("remote", packageType)()
		return &VcsRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
