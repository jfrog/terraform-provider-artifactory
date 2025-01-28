package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

// SDKv2
type GoRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryCurationParams
	VcsGitProvider string `json:"vcsGitProvider"`
}

var GoSchema = lo.Assign(
	remote.BaseSchema,
	remote.CurationRemoteRepoSchema,
	map[string]*schema.Schema{
		"vcs_git_provider": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "ARTIFACTORY",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(remote.SupportedGoVCSGitProviders, false)),
			Description:      "Artifactory supports proxying the following Git providers out-of-the-box: GitHub (`GITHUB`), GitHub Enterprise (`GITHUBENTERPRISE`), BitBucket Cloud (`BITBUCKET`), BitBucket Server (`STASH`), GitLab (`GITLAB`), or a remote Artifactory instance (`ARTIFACTORY`). Default value is `ARTIFACTORY`.",
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.GoPackageType),
)

var GoSchemas = remote.GetSchemas(GoSchema)

func DataSourceArtifactoryRemoteGoRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.GoPackageType)
		if err != nil {
			return nil, err
		}

		return &GoRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.GoPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	goSchema := getSchema(GoSchemas)

	return &schema.Resource{
		Schema:      goSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(goSchema), constructor),
		Description: "Provides a data source for a remote Go repository",
	}
}
