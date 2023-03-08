package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteTerraformRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.TerraformPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.TerraformRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   remote.TerraformPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	terraformSchema := remote.TerraformRemoteSchema(false)

	return &schema.Resource{
		Schema:      terraformSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(terraformSchema), constructor),
		Description: "Provides a data source for a remote Terraform repository",
	}
}
