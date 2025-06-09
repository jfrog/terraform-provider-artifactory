package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalHexRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.HexLocalRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: resource_repository.HexPackageType,
				Rclass:      local.Rclass,
			},
			HexPrimaryKeyPairRef: "",
		}, nil
	}

	return &schema.Resource{
		Schema:      local.HexLocalSchemas[local.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.HexLocalSchemas[local.CurrentSchemaVersion]), constructor),
		Description: "Data source for a local hex repository",
	}
}
