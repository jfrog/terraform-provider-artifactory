package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteAnsibleRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.AnsiblePackageType)
		if err != nil {
			return nil, err
		}

		return &remote.AnsibleRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.AnsiblePackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	ansibleSchema := getSchema(remote.AnsibleSchemas)

	return &schema.Resource{
		Schema:      ansibleSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(ansibleSchema), constructor),
		Description: "Provides a data source for a remote Ansible repository",
	}
}
