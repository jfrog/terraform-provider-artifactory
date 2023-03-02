package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

var AlpineLocalSchema = util.MergeMaps(
	BaseLocalRepoSchema,
	map[string]*schema.Schema{
		"primary_keypair_ref": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "Used to sign index files in Alpine Linux repositories. " +
				"See: https://www.jfrog.com/confluence/display/JFROG/Alpine+Linux+Repositories#AlpineLinuxRepositories-SigningAlpineLinuxIndex",
		},
	},
	repository.RepoLayoutRefSchema("local", "alpine"),
	repository.CompressionFormats,
)

type AlpineLocalRepoParams struct {
	RepositoryBaseParams
	PrimaryKeyPairRef string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
}

func UnpackLocalAlpineRepository(data *schema.ResourceData, rclass string) AlpineLocalRepoParams {
	d := &util.ResourceData{ResourceData: data}
	return AlpineLocalRepoParams{
		RepositoryBaseParams: UnpackBaseRepo(rclass, data, "alpine"),
		PrimaryKeyPairRef:    d.GetString("primary_keypair_ref", false),
	}
}

func ResourceArtifactoryLocalAlpineRepository() *schema.Resource {
	var unpackLocalAlpineRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalAlpineRepository(data, rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &AlpineLocalRepoParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: "alpine",
				Rclass:      rclass,
			},
		}, nil
	}

	return repository.MkResourceSchema(AlpineLocalSchema, packer.Default(AlpineLocalSchema), unpackLocalAlpineRepo, constructor)
}
