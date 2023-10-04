package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const GoPackageType = "go"

type GoRemoteRepo struct {
	RepositoryRemoteBaseParams
	VcsGitProvider string `json:"vcsGitProvider"`
}

var GoRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		map[string]*schema.Schema{
			"vcs_git_provider": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "ARTIFACTORY",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"GITHUB", "ARTIFACTORY"}, false)),
				Description:      `Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is "ARTIFACTORY".`,
			},
		},
		repository.RepoLayoutRefSchema(rclass, GoPackageType),
	)
}

func ResourceArtifactoryRemoteGoRepository() *schema.Resource {

	var unpackGoRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := GoRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, GoPackageType),
			VcsGitProvider:             d.GetString("vcs_git_provider", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, GoPackageType)()
		if err != nil {
			return nil, err
		}

		return &GoRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   GoPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	goSchema := GoRemoteSchema(true)

	return mkResourceSchema(goSchema, packer.Default(goSchema), unpackGoRemoteRepo, constructor)
}
