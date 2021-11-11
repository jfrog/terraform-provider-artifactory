package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArtifactoryLocalNpmRepository() *schema.Resource {

	return mkResourceSchema(baseLocalRepoSchema, universalPack, unPackLocalNpmRepository, func() interface{} {
		return &LocalRepositoryBaseParams{
			PackageType: "npm",
			Rclass:      "local",
		}
	})
}

func unPackLocalNpmRepository(data *schema.ResourceData) (interface{}, string, error) {
	repo := unpackBaseLocalRepo(data, "npm")
	return repo, repo.Id(), nil
}
