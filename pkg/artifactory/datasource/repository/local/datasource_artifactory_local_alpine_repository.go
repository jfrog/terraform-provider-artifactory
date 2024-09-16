package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryLocalAnsibleRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		return &local.AnsibleLocalRepoParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: "ansible",
				Rclass:      rclass,
			},
		}, nil
	}

	return &schema.Resource{
		Schema:      local.AnsibleLocalSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.AnsibleLocalSchema), constructor),
		Description: "Data source for a local Ansible repository",
	}
}
