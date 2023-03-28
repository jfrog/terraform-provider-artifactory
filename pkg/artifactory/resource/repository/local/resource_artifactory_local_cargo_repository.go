package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

const cargoPackageType = "cargo"

var CargoLocalSchema = util.MergeMaps(
	BaseLocalRepoSchema,
	map[string]*schema.Schema{
		"anonymous_access": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Cargo client does not send credentials when performing download and search for crates. Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option. Default value is 'false'.",
		},
		"enable_sparse_index": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable internal index support based on Cargo sparse index specifications, instead of the default git index. Default value is 'false'.",
		},
	},
	repository.RepoLayoutRefSchema(rclass, cargoPackageType),
	repository.CompressionFormats,
)

type CargoLocalRepoParams struct {
	RepositoryBaseParams
	AnonymousAccess   bool `json:"cargoAnonymousAccess"`
	EnableSparseIndex bool `json:"cargoInternalIndex"`
}

func UnpackLocalCargoRepository(data *schema.ResourceData, rclass string) CargoLocalRepoParams {
	d := &util.ResourceData{ResourceData: data}
	return CargoLocalRepoParams{
		RepositoryBaseParams: UnpackBaseRepo(rclass, data, cargoPackageType),
		AnonymousAccess:      d.GetBool("anonymous_access", false),
		EnableSparseIndex:    d.GetBool("enable_sparse_index", false),
	}
}

func ResourceArtifactoryLocalCargoRepository() *schema.Resource {

	var unpackLocalCargoRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalCargoRepository(data, rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &CargoLocalRepoParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: cargoPackageType,
				Rclass:      rclass,
			},
		}, nil
	}

	return repository.MkResourceSchema(CargoLocalSchema, packer.Default(CargoLocalSchema), unpackLocalCargoRepository, constructor)
}
