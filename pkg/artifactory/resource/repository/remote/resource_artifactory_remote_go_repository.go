package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

type GoRemoteRepo struct {
	RemoteRepositoryBaseParams
	VcsGitProvider string `json:"vcsGitProvider"`
}

func ResourceArtifactoryRemoteGoRepository() *schema.Resource {
	const packageType = "go"

	var goRemoteSchema = util.MergeSchema(BaseRemoteRepoSchema, map[string]*schema.Schema{
		"vcs_git_provider": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "ARTIFACTORY",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"GITHUB", "ARTIFACTORY"}, false)),
			Description:      `Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is "ARTIFACTORY".`,
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackGoRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{s}
		repo := GoRemoteRepo{
			RemoteRepositoryBaseParams: UnpackBaseRemoteRepo(s, packageType),
			VcsGitProvider:             d.GetString("vcs_git_provider", false),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(goRemoteSchema, repository.DefaultPacker(goRemoteSchema), unpackGoRemoteRepo, func() interface{} {
		repoLayout, _ := repository.GetDefaultRepoLayoutRef("remote", packageType)()
		return &GoRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
