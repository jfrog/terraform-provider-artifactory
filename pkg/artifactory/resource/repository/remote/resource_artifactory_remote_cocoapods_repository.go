package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type CocoapodsRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	PodsSpecsRepoUrl string `json:"podsSpecsRepoUrl"`
}

const CocoapodsPackageType = "cocoapods"

var CocoapodsRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(BaseRemoteRepoSchema(isResource), VcsRemoteRepoSchema, map[string]*schema.Schema{
		"pods_specs_repo_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://github.com/CocoaPods/Specs",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `Proxy remote CocoaPods Specs repositories. Default value is "https://github.com/CocoaPods/Specs".`,
		},
	}, repository.RepoLayoutRefSchema(rclass, CocoapodsPackageType))
}

func ResourceArtifactoryRemoteCocoapodsRepository() *schema.Resource {
	var unpackCocoapodsRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := CocoapodsRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, CocoapodsPackageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			PodsSpecsRepoUrl:           d.GetString("pods_specs_repo_url", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, CocoapodsPackageType)()
		if err != nil {
			return nil, err
		}

		return &CocoapodsRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   CocoapodsPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	cocoapodsSchema := CocoapodsRemoteSchema(true)

	return mkResourceSchema(cocoapodsSchema, packer.Default(cocoapodsSchema), unpackCocoapodsRemoteRepo, constructor)
}
