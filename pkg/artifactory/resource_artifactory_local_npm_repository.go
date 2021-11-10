package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var npmLocalSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{})

func resourceArtifactoryLocalNpmRepository() *schema.Resource {

	return mkResourceSchema(npmLocalSchema, universalPack, unPackLocalNpmRepository, func() interface{} {
		return &NpmLocalRepositoryParams{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "npm",
				Rclass:      "local",
			},
		}
	})
}

type NpmLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	Key string `hcl:"key" json:"key"`
}

func unPackLocalNpmRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{ResourceData: data}
	repo := NpmLocalRepositoryParams{
		LocalRepositoryBaseParams: unpackBaseLocalRepo(data, "npm"),
		Key:                       d.getString("key", false),
	}
	repo.PackageType = "npm"
	return repo, repo.Id(), nil
}
