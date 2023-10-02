package local

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalTerraformRepository(registryType string) *schema.Resource {

	terraformLocalSchema := local.GetTerraformLocalSchema(registryType)

	constructor := func() (interface{}, error) {
		return &local.RepositoryBaseParams{
			PackageType: "terraform_" + registryType,
			Rclass:      rclass,
		}, nil
	}

	return &schema.Resource{
		Schema:      terraformLocalSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(terraformLocalSchema), constructor),
		Description: fmt.Sprintf("Data Source for a local terraform_%s repository", registryType),
	}
}
