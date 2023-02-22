package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalAlpineRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.AlpineLocalRepoParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: "alpine",
				Rclass:      rclass,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      local.AlpineLocalSchema,
		ReadContext: MkRepoReadDataSource(packer.Default(local.AlpineLocalSchema), constructor),
		Description: "Dummy description",
	}
}
