package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type VcsRemoteRepo struct {
	RemoteRepositoryBaseParams
	VcsGitProvider     string `json:"vcsGitProvider"`
	MaxUniqueSnapshots int    `json:"maxUniqueSnapshots"`
}

func resourceArtifactoryRemoteVcsRepository() *schema.Resource {
	const packageType = "vcs"

	var vcsRemoteSchema = mergeSchema(baseRemoteRepoSchema, map[string]*schema.Schema{
		"vcs_git_provider": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "GITHUB",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"GITHUB", "BITBUCKET", "OLDSTASH", "STASH", "ARTIFACTORY", "CUSTOM"}, false)),
			Description: "Artifactory supports proxying the following Git providers out-of-the-box: GitHub, Bitbucket, " +
				"Stash, a remote Artifactory instance or a custom Git repository. Default value is 'GITHUB'.",
		},
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
			VcsGitProvider:             d.getString("vcs_git_provider", false),
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
