package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemotecoComposerRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(rclass, remote.ComposerPackageType)()
		if err != nil {
			return nil, err
		}

		return &remote.ComposerRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   remote.ComposerPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	composerSchema := remote.ComposerRemoteSchema(false)

	return &schema.Resource{
		Schema:      composerSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(composerSchema), constructor),
		Description: "Provides a data source for a remote Composer repository",
	}
}
