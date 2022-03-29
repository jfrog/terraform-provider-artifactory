package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArtifactoryLocalCargoRepository() *schema.Resource {
	const packageType = "cargo"

	var cargoLocalSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
		"anonymous_access": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `(Optional) Cargo client does not send credentials when performing download and search for crates. Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option. Default value is 'false'.`,
		},
	}, repoLayoutRefSchema("local", packageType), compressionFormats)

	type CargoLocalRepo struct {
		LocalRepositoryBaseParams
		CargoAnonymousAccess bool `hcl:"anonymous_access" json:"cargoAnonymousAccess"`
	}

	var unPackLocalCargoRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{ResourceData: data}
		repo := CargoLocalRepo{
			LocalRepositoryBaseParams: unpackBaseRepo("local", data, packageType),
			CargoAnonymousAccess:      d.getBool("anonymous_access", false),
		}

		return repo, repo.Id(), nil
	}

	return mkResourceSchema(cargoLocalSchema, defaultPacker, unPackLocalCargoRepository, func() interface{} {
		return &CargoLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: packageType,
				Rclass:      "local",
			},
		}
	})
}
