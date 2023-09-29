package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type VcsRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	MaxUniqueSnapshots int `json:"maxUniqueSnapshots"`
}

const VcsPackageType = "vcs"

var VcsRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseRemoteRepoSchema(isResource),
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
		repository.RepoLayoutRefSchema(rclass, VcsPackageType),
	)
}

func ResourceArtifactoryRemoteVcsRepository() *schema.Resource {
	var UnpackVcsRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := VcsRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, VcsPackageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			MaxUniqueSnapshots:         d.GetInt("max_unique_snapshots", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, VcsPackageType)()
		if err != nil {
			return nil, err
		}

		return &VcsRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   VcsPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	vcsSchema := VcsRemoteSchema(true)

	return mkResourceSchema(vcsSchema, packer.Default(vcsSchema), UnpackVcsRemoteRepo, constructor)
}
