package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type CocoapodsRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	PodsSpecsRepoUrl string `json:"podsSpecsRepoUrl"`
}

var cocoapodsSchema = lo.Assign(
	BaseSchema,
	VcsRemoteRepoSchema,
	map[string]*schema.Schema{
		"pods_specs_repo_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://github.com/CocoaPods/Specs",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `Proxy remote CocoaPods Specs repositories. Default value is "https://github.com/CocoaPods/Specs".`,
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.CocoapodsPackageType),
)

var CocoapodsSchemas = GetSchemas(cocoapodsSchema)

func ResourceArtifactoryRemoteCocoapodsRepository() *schema.Resource {
	var unpackCocoapodsRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := CocoapodsRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.CocoapodsPackageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			PodsSpecsRepoUrl:           d.GetString("pods_specs_repo_url", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.CocoapodsPackageType)
		if err != nil {
			return nil, err
		}

		return &CocoapodsRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.CocoapodsPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	return mkResourceSchema(
		CocoapodsSchemas,
		packer.Default(CocoapodsSchemas[CurrentSchemaVersion]),
		unpackCocoapodsRemoteRepo,
		constructor,
	)
}
