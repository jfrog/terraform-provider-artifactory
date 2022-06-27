package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type CargoRemoteRepo struct {
	RepositoryBaseParams
	RegistryUrl     string `hcl:"git_registry_url" json:"gitRegistryUrl"`
	AnonymousAccess bool   `hcl:"anonymous_access" json:"cargoAnonymousAccess"`
}

func ResourceArtifactoryRemoteCargoRepository() *schema.Resource {
	const packageType = "cargo"

	var cargoRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, map[string]*schema.Schema{
		"git_registry_url": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `This is the index url, expected to be a git repository. Default value in UI is "https://github.com/rust-lang/crates.io-index"`,
		},
		"anonymous_access": {
			Type:     schema.TypeBool,
			Optional: true,
			Description: "(On the UI: Anonymous download and search) Cargo client does not send credentials when performing download and search for crates. " +
				"Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option.",
		},
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackCargoRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := CargoRemoteRepo{
			RepositoryBaseParams: UnpackBaseRemoteRepo(s, packageType),
			RegistryUrl:          d.GetString("git_registry_url", false),
			AnonymousAccess:      d.GetBool("anonymous_access", false),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(cargoRemoteSchema, packer.Default(cargoRemoteSchema), unpackCargoRemoteRepo, func() interface{} {
		return &CargoRemoteRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "remote",
				PackageType: packageType,
			},
		}
	})
}
