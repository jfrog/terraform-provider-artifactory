package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
)

var alpineLocalSchema = util.MergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
	"primary_keypair_ref": {
		Type:     schema.TypeString,
		Optional: true,
		Description: "Used to sign index files in Alpine Linux repositories. " +
			"See: https://www.jfrog.com/confluence/display/JFROG/Alpine+Linux+Repositories#AlpineLinuxRepositories-SigningAlpineLinuxIndex",
	},
}, CompressionFormats)

func ResourceArtifactoryLocalAlpineRepository() *schema.Resource {
	return repos.MkResourceSchema(alpineLocalSchema, util.DefaultPacker, unPackLocalAlpineRepository, func() interface{} {
		return &AlpineLocalRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: "alpine",
				Rclass:      "local",
			},
		}
	})
}

type AlpineLocalRepo struct {
	RepositoryBaseParams
	PrimaryKeyPairRef string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
}

func unPackLocalAlpineRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &util.ResourceData{ResourceData: data}
	repo := AlpineLocalRepo{
		RepositoryBaseParams: unpackBaseLocalRepo(data, "alpine"),
		PrimaryKeyPairRef:    d.GetString("primary_keypair_ref", false),
	}

	return repo, repo.Id(), nil
}
