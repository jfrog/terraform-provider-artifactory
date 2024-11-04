package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type GoRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryCurationParams
	VcsGitProvider string `json:"vcsGitProvider"`
}

var SupportedGoVCSGitProviders = []string{
	"ARTIFACTORY",
	"BITBUCKET",
	"GITHUB",
	"GITHUBENTERPRISE",
	"GITLAB",
	"STASH",
}

var GoSchema = lo.Assign(
	baseSchema,
	CurationRemoteRepoSchema,
	map[string]*schema.Schema{
		"vcs_git_provider": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "ARTIFACTORY",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(SupportedGoVCSGitProviders, false)),
			Description:      "Artifactory supports proxying the following Git providers out-of-the-box: GitHub (`GITHUB`), GitHub Enterprise (`GITHUBENTERPRISE`), BitBucket Cloud (`BITBUCKET`), BitBucket Server (`STASH`), GitLab (`GITLAB`), or a remote Artifactory instance (`ARTIFACTORY`). Default value is `ARTIFACTORY`.",
		},
	},
	repository.RepoLayoutRefSchema(Rclass, repository.GoPackageType),
)

var GoSchemas = GetSchemas(GoSchema)

func ResourceArtifactoryRemoteGoRepository() *schema.Resource {

	var unpackGoRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := GoRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.GoPackageType),
			RepositoryCurationParams: RepositoryCurationParams{
				Curated: d.GetBool("curated", false),
			},
			VcsGitProvider: d.GetString("vcs_git_provider", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.GoPackageType)()
		if err != nil {
			return nil, err
		}

		return &GoRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.GoPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	return mkResourceSchema(
		GoSchemas,
		packer.Default(GoSchemas[CurrentSchemaVersion]),
		unpackGoRemoteRepo,
		constructor,
	)
}
