package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type VcsRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	MaxUniqueSnapshots int `json:"maxUniqueSnapshots"`
}

var VCSSchema = lo.Assign(
	baseSchema,
	VcsRemoteRepoSchema,
	map[string]*schema.Schema{
		"max_unique_snapshots": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			Description: "The maximum number of unique snapshots of a single artifact to store. Once the number of " +
				"snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is " +
				"no limit, and unique snapshots are not cleaned up.",
		},
	},
	repository.RepoLayoutRefSchema(Rclass, repository.VCSPackageType),
)

var VCSSchemas = GetSchemas(VCSSchema)

func ResourceArtifactoryRemoteVcsRepository() *schema.Resource {
	var UnpackVcsRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := VcsRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.VCSPackageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			MaxUniqueSnapshots:         d.GetInt("max_unique_snapshots", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.VCSPackageType)()
		if err != nil {
			return nil, err
		}

		return &VcsRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.VCSPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	return mkResourceSchema(
		VCSSchemas,
		packer.Default(VCSSchemas[CurrentSchemaVersion]),
		UnpackVcsRemoteRepo,
		constructor,
	)
}
