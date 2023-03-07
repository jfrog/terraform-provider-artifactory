package remote

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteBasicRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, packageType)()
		if err != nil {
			return nil, err
		}

		return &remote.RepositoryRemoteBaseParams{
			PackageType:   packageType,
			Rclass:        rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	basicRepoSchema := remote.BasicRepoSchema(packageType, false)

	return &schema.Resource{
		Schema:      basicRepoSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(basicRepoSchema), constructor),
		Description: fmt.Sprintf("Provides a data source for a remote %s repository", packageType),
	}
}
