package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemoteComposerRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.ComposerPackageType)
		if err != nil {
			return nil, err
		}

		return &remote.ComposerRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.ComposerPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	composerSchema := getSchema(remote.ComposerSchemas)

	return &schema.Resource{
		Schema:      composerSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(composerSchema), constructor),
		Description: "Provides a data source for a remote Composer repository",
	}
}
