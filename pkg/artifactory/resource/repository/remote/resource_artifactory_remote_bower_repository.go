package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type BowerRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	BowerRegistryUrl string `json:"bowerRegistryUrl"`
}

const BowerPackageType = "bower"

var BowerRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		VcsRemoteRepoSchema,
		map[string]*schema.Schema{
			"bower_registry_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "https://registry.bower.io",
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Description:  `Proxy remote Bower repository. Default value is "https://registry.bower.io".`,
			},
		},
		repository.RepoLayoutRefSchema(rclass, BowerPackageType),
	)
}

func ResourceArtifactoryRemoteBowerRepository() *schema.Resource {

	var unpackBowerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := BowerRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, BowerPackageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			BowerRegistryUrl:           d.GetString("bower_registry_url", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, BowerPackageType)()
		if err != nil {
			return nil, err
		}

		return &BowerRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   BowerPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	bowerSchema := BowerRemoteSchema(true)

	return mkResourceSchema(bowerSchema, packer.Default(bowerSchema), unpackBowerRemoteRepo, constructor)
}
