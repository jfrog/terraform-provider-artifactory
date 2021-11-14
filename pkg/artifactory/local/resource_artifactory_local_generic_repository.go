package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
)

func ResourceArtifactoryLocalGenericRepository(pkt string) *schema.Resource {
	constructor := func() interface{} {
		return &RepositoryBaseParams{
			PackageType: pkt,
			Rclass:      "local",
		}
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := unpackBaseLocalRepo(data, pkt)
		return repo, repo.Id(), nil
	}
	packer := util.UniversalPack(util.SchemaHasKey(baseLocalRepoSchema))
	return repos.MkResourceSchema(baseLocalRepoSchema, packer, unpack, constructor)
}
