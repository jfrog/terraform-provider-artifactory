package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

const packageType = "docker"

func DataSourceArtifactoryLocalDockerV2Repository() *schema.Resource {
	pkr := packer.Default(local.DockerV2LocalSchema)

	constructor := func() (interface{}, error) {
		return &local.DockerLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: packageType,
				Rclass:      rclass,
			},
			DockerApiVersion:    "V2",
			TagRetention:        1,
			MaxUniqueTags:       0, // no limit
			BlockPushingSchema1: true,
		}, nil
	}

	return &schema.Resource{
		Schema:      local.DockerV2LocalSchema,
		ReadContext: repository.MkRepoReadDataSource(pkr, constructor),
	}
}

func DataSourceArtifactoryLocalDockerV1Repository() *schema.Resource {
	// this is necessary because of the pointers
	skeema := util.MergeMaps(local.DockerV1LocalSchema)

	constructor := func() (interface{}, error) {
		return &local.DockerLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: packageType,
				Rclass:      rclass,
			},
			DockerApiVersion:    "V1",
			TagRetention:        1,
			MaxUniqueTags:       0,
			BlockPushingSchema1: false,
		}, nil
	}

	return &schema.Resource{
		Schema:      skeema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.DockerV1LocalSchema), constructor),
		Description: "Provides a data source for a local docker (v1) repository",
	}
}
