package remote

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteBasicRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, packageType)
		if err != nil {
			return nil, err
		}

		return &remote.RepositoryRemoteBaseParams{
			PackageType:   packageType,
			Rclass:        remote.Rclass,
			RepoLayoutRef: repoLayout,
		}, nil
	}

	basicSchemas := remote.GetSchemas(remote.BasicSchema(packageType))
	basicSchema := getSchema(basicSchemas)

	return &schema.Resource{
		Schema:      basicSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(basicSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a remote %s repository", packageType),
	}
}
