package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var cargoSchema = lo.Assign(
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.CargoPackageType),
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
	repository.CompressionFormatsSDKv2,
)

var CargoSchemas = GetSchemas(cargoSchema)

type CargoLocalRepoParams struct {
	RepositoryBaseParams
	AnonymousAccess   bool `json:"cargoAnonymousAccess"`
	EnableSparseIndex bool `json:"cargoInternalIndex"`
}

func UnpackLocalCargoRepository(data *schema.ResourceData, Rclass string) CargoLocalRepoParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return CargoLocalRepoParams{
		RepositoryBaseParams: UnpackBaseRepo(Rclass, data, repository.CargoPackageType),
		AnonymousAccess:      d.GetBool("anonymous_access", false),
		EnableSparseIndex:    d.GetBool("enable_sparse_index", false),
	}
}

func ResourceArtifactoryLocalCargoRepository() *schema.Resource {

	var unpackLocalCargoRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalCargoRepository(data, Rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &CargoLocalRepoParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: repository.CargoPackageType,
				Rclass:      Rclass,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		CargoSchemas,
		packer.Default(CargoSchemas[CurrentSchemaVersion]),
		unpackLocalCargoRepository,
		constructor,
	)
}
