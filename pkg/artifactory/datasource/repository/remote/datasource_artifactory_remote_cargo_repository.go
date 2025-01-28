package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

type CargoRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	RegistryUrl       string `hcl:"git_registry_url" json:"gitRegistryUrl"`
	AnonymousAccess   bool   `json:"cargoAnonymousAccess"`
	EnableSparseIndex bool   `json:"cargoInternalIndex"`
}

var cargoSchema = lo.Assign(
	remote.BaseSchema,
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
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.CargoPackageType),
)

var CargoSchemas = remote.GetSchemas(cargoSchema)

func DataSourceArtifactoryRemoteCargoRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.CargoPackageType)
		if err != nil {
			return nil, err
		}

		return &CargoRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.CargoPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	cargoSchema := getSchema(CargoSchemas)
	cargoSchema["git_registry_url"].Required = false
	cargoSchema["git_registry_url"].Optional = true

	return &schema.Resource{
		Schema:      cargoSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(cargoSchema), constructor),
		Description: "Provides a data source for a remote Cargo repository",
	}
}
