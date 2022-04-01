package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type CargoRemoteRepo struct {
	RemoteRepositoryBaseParams
	RegistryUrl     string `hcl:"git_registry_url" json:"gitRegistryUrl"`
	AnonymousAccess bool   `hcl:"anonymous_access" json:"cargoAnonymousAccess"`
}

func resourceArtifactoryRemoteCargoRepository() *schema.Resource {
	const packageType = "cargo"

	var cargoRemoteSchema = mergeSchema(baseRemoteRepoSchema, map[string]*schema.Schema{
		"git_registry_url": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `This is the index url, expected to be a git repository. for remote artifactory use "arturl/git/repokey.git"`,
		},
		"anonymous_access": {
			Type:     schema.TypeBool,
			Optional: true,
			Description: "(On the UI: Anonymous download and search) Cargo client does not send credentials when performing download and search for crates. " +
				"Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option.",
		},
	}, repoLayoutRefSchema("remote", packageType))

	var unpackCargoRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}
		repo := CargoRemoteRepo{
			RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s, packageType),
			RegistryUrl:                d.getString("git_registry_url", false),
			AnonymousAccess:            d.getBool("anonymous_access", false),
		}
		return repo, repo.Id(), nil
	}

	return mkResourceSchema(cargoRemoteSchema, noPasswordPacker, unpackCargoRemoteRepo, func() interface{} {
		return &CargoRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:      "remote",
				PackageType: packageType,
			},
		}
	})
}
