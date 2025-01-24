package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

type NpmRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryCurationParams
}

var NPMSchema = lo.Assign(
	remote.BaseSchema,
	remote.CurationRemoteRepoSchema,
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.NPMPackageType),
)

var NPMSchemas = remote.GetSchemas(NPMSchema)

func DataSourceArtifactoryRemoteNpmRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.NPMPackageType)
		if err != nil {
			return nil, err
		}

		return &NpmRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.NPMPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	npmSchema := getSchema(NPMSchemas)

	return &schema.Resource{
		Schema:      npmSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(npmSchema), constructor),
		Description: "Provides a data source for a remote NPM repository",
	}
}
