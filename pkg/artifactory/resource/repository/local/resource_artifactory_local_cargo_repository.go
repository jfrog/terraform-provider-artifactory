package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

var cargoLocalSchema = util.MergeSchema(
	BaseLocalRepoSchema,
	map[string]*schema.Schema{
		"anonymous_access": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: `Cargo client does not send credentials when performing download and search for crates. Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option. Default value is 'false'.`,
		},
	},
	repository.RepoLayoutRefSchema("local", "cargo"),
	repository.CompressionFormats,
)

func ResourceArtifactoryLocalCargoRepository() *schema.Resource {

	type CargoLocalRepo struct {
		LocalRepositoryBaseParams
		AnonymousAccess bool `json:"cargoAnonymousAccess"`
	}

	var unPackLocalCargoRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: data}
		repo := CargoLocalRepo{
			LocalRepositoryBaseParams: UnpackBaseRepo("local", data, "cargo"),
			AnonymousAccess:           d.GetBool("anonymous_access", false),
		}

		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(cargoLocalSchema, repository.DefaultPacker(cargoLocalSchema), unPackLocalCargoRepository, func() interface{} {
		return &CargoLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "cargo",
				Rclass:      "local",
			},
		}
	})
}
