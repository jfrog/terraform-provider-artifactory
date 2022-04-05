package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type CocoapodsRemoteRepo struct {
	RemoteRepositoryBaseParams
	RemoteRepositoryVcsParams
	PodsSpecsRepoUrl string `json:"podsSpecsRepoUrl"`
}

func resourceArtifactoryRemoteCocoapodsRepository() *schema.Resource {
	const packageType = "cocoapods"

	var cocoapodsRemoteSchema = mergeSchema(baseRemoteRepoSchema, vcsRemoteRepoSchema, map[string]*schema.Schema{
		"pods_specs_repo_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://github.com/CocoaPods/Specs",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `(Optional) Proxy remote CocoaPods Specs repositories. Default value is "https://github.com/CocoaPods/Specs".`,
		},
	}, repoLayoutRefSchema("remote", packageType))

	var unpackCocoapodsRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}
		repo := CocoapodsRemoteRepo{
			RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s, packageType),
			RemoteRepositoryVcsParams:  unpackVcsRemoteRepo(s),
			PodsSpecsRepoUrl:           d.getString("pods_specs_repo_url", false),
		}
		return repo, repo.Id(), nil
	}

	return mkResourceSchema(cocoapodsRemoteSchema, defaultPacker(cocoapodsRemoteSchema), unpackCocoapodsRemoteRepo, func() interface{} {
		repoLayout, _ := getDefaultRepoLayoutRef("remote", packageType)()
		return &CocoapodsRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
