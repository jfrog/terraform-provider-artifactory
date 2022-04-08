package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type VcsRemoteRepo struct {
	RemoteRepositoryBaseParams
	RemoteRepositoryVcsParams
	MaxUniqueSnapshots int `json:"maxUniqueSnapshots"`
}

func resourceArtifactoryRemoteVcsRepository() *schema.Resource {
	const packageType = "vcs"

	var vcsRemoteSchema = mergeSchema(baseRemoteRepoSchema, vcsRemoteRepoSchema, map[string]*schema.Schema{
		"max_unique_snapshots": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			Description: "The maximum number of unique snapshots of a single artifact to store. Once the number of " +
				"snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is " +
				"no limit, and unique snapshots are not cleaned up.",
		},
	}, repoLayoutRefSchema("remote", packageType))

	var unpackVcsRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}
		repo := VcsRemoteRepo{
			RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s, packageType),
			RemoteRepositoryVcsParams:  unpackVcsRemoteRepo(s),
			MaxUniqueSnapshots:         d.getInt("max_unique_snapshots", false),
		}
		return repo, repo.Id(), nil
	}

	return mkResourceSchema(vcsRemoteSchema, defaultPacker(vcsRemoteSchema), unpackVcsRemoteRepo, func() interface{} {
		repoLayout, _ := getDefaultRepoLayoutRef("remote", packageType)()
		return &VcsRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
