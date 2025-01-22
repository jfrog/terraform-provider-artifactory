package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type CargoRemoteRepo struct {
	RepositoryRemoteBaseParams
	RegistryUrl       string `hcl:"git_registry_url" json:"gitRegistryUrl"`
	AnonymousAccess   bool   `json:"cargoAnonymousAccess"`
	EnableSparseIndex bool   `json:"cargoInternalIndex"`
}

var cargoSchema = lo.Assign(
	BaseSchema,
	map[string]*schema.Schema{
		"git_registry_url": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `This is the index url, expected to be a git repository. Default value in UI is "https://index.crates.io/"`,
		},
		"anonymous_access": {
			Type:     schema.TypeBool,
			Optional: true,
			Description: "(On the UI: Anonymous download and search) Cargo client does not send credentials when performing download and search for crates. " +
				"Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option.",
		},
		"enable_sparse_index": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable internal index support based on Cargo sparse index specifications, instead of the default git index. Default value is 'false'.",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.CargoPackageType),
)

var CargoSchemas = GetSchemas(cargoSchema)

func ResourceArtifactoryRemoteCargoRepository() *schema.Resource {

	var unpackCargoRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := CargoRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.CargoPackageType),
			RegistryUrl:                d.GetString("git_registry_url", false),
			AnonymousAccess:            d.GetBool("anonymous_access", false),
			EnableSparseIndex:          d.GetBool("enable_sparse_index", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &CargoRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      Rclass,
				PackageType: repository.CargoPackageType,
			},
		}, nil
	}

	return mkResourceSchema(
		CargoSchemas,
		packer.Default(CargoSchemas[CurrentSchemaVersion]),
		unpackCargoRemoteRepo,
		constructor,
	)
}
