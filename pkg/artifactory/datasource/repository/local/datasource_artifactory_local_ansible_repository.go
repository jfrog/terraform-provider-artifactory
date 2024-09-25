package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalAnsibleRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.AnsibleLocalRepoParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: resource_repository.AnsiblePackageType,
				Rclass:      local.Rclass,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      local.AnsibleSchemas[local.CurrentSchemaVersion],
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.AnsibleSchemas[local.CurrentSchemaVersion]), constructor),
		Description: "Data source for a local Ansible repository",
	}
}
