package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type BowerRemoteRepo struct {
	RepositoryRemoteBaseParams
	RepositoryVcsParams
	BowerRegistryUrl string `json:"bowerRegistryUrl"`
}

var bowerSchema = lo.Assign(
	BaseSchema,
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
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.BowerPackageType),
)

var BowerSchemas = GetSchemas(bowerSchema)

func ResourceArtifactoryRemoteBowerRepository() *schema.Resource {

	var unpackBowerRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := BowerRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.BowerPackageType),
			RepositoryVcsParams:        UnpackVcsRemoteRepo(s),
			BowerRegistryUrl:           d.GetString("bower_registry_url", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.BowerPackageType)
		if err != nil {
			return nil, err
		}

		return &BowerRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.BowerPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	return mkResourceSchema(
		BowerSchemas,
		packer.Default(BowerSchemas[CurrentSchemaVersion]),
		unpackBowerRemoteRepo,
		constructor,
	)
}
