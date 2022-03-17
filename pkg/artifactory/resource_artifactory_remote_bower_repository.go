package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type BowerRemoteRepo struct {
	RemoteRepositoryBaseParams
	RemoteRepositoryVcsParams
	BowerRegistryUrl string `json:"bowerRegistryUrl"`
}

func resourceArtifactoryRemoteBowerRepository() *schema.Resource {
	const packageType = "bower"

	var bowerRemoteSchema = mergeSchema(baseRemoteRepoSchema, vcsRemoteRepoSchema, map[string]*schema.Schema{
		"bower_registry_url": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "https://registry.bower.io",
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `(Optional) Proxy remote Bower repository. Default value is "https://registry.bower.io".`,
		},
	}, repoLayoutRefSchema("remote", packageType))

	var unpackBowerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}
		repo := BowerRemoteRepo{
			RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s, packageType),
			RemoteRepositoryVcsParams:  unpackVcsRemoteRepo(s),
			BowerRegistryUrl:           d.getString("bower_registry_url", false),
		}
		return repo, repo.Id(), nil
	}

	return mkResourceSchema(bowerRemoteSchema, defaultPacker, unpackBowerRemoteRepo, func() interface{} {
		repoLayout, _ := getDefaultRepoLayoutRef("remote", packageType)()
		return &BowerRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         packageType,
				RemoteRepoLayoutRef: repoLayout.(string),
			},
		}
	})
}
