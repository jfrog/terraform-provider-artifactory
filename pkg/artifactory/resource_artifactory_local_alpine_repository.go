package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func resourceArtifactoryLocalAlpineRepository() *schema.Resource {
	const packageType = "alpine"

	var alpineLocalSchema = utils.MergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
		"primary_keypair_ref": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "Used to sign index files in Alpine Linux repositories. " +
				"See: https://www.jfrog.com/confluence/display/JFROG/Alpine+Linux+Repositories#AlpineLinuxRepositories-SigningAlpineLinuxIndex",
		},
	}, repoLayoutRefSchema("local", packageType), compressionFormats)

	type AlpineLocalRepo struct {
		LocalRepositoryBaseParams
		PrimaryKeyPairRef string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
	}

	var unPackLocalAlpineRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utils.ResourceData{ResourceData: data}
		repo := AlpineLocalRepo{
			LocalRepositoryBaseParams: unpackBaseRepo("local", data, packageType),
			PrimaryKeyPairRef:         d.GetString("primary_keypair_ref", false),
		}

		return repo, repo.Id(), nil
	}

	return mkResourceSchema(alpineLocalSchema, defaultPacker(alpineLocalSchema), unPackLocalAlpineRepository, func() interface{} {
		return &AlpineLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: packageType,
				Rclass:      "local",
			},
		}
	})
}
