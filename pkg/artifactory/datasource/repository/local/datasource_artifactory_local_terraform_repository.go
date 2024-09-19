package local

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalTerraformRepository(registryType string) *schema.Resource {
	terraformLocalSchemas := local.GetTerraformSchemas(registryType)

	constructor := func() (interface{}, error) {
		return &local.RepositoryBaseParams{
			PackageType: "terraform_" + registryType,
			Rclass:      local.Rclass,
		}, nil
	}

	return &schema.Resource{
		Schema:      terraformLocalSchemas[local.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(terraformLocalSchemas[local.CurrentSchemaVersion]), constructor),
		Description: fmt.Sprintf("Data Source for a local terraform_%s repository", registryType),
	}
}
