package local

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
)

var _ datasource.DataSource = &AlpineLocalRepositoryDataSource{}

type AlpineLocalRepositoryDataSource struct{}

func (d *AlpineLocalRepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_local_alpine_repository"
}

func (d *AlpineLocalRepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: local.AlpineLocalRepositoryAttributes,
	}
}

//
// func DataSourceArtifactoryLocalAlpineRepository() *schema.Resource {
// 	constructor := func() (interface{}, error) {
// 		return &local.AlpineLocalRepoParams{
// 			RepositoryBaseParams: local.RepositoryBaseParams{
// 				PackageType: resource_repository.AlpinePackageType,
// 				Rclass:      local.Rclass,
// 			},
// 		}, nil
// 	}
//
// 	return &schema.Resource{
// 		Schema:      local.AlpineLocalSchemas[local.CurrentSchemaVersion],
// 		ReadContext: repository.MkRepoReadDataSource(packer.Default(local.AlpineLocalSchemas[local.CurrentSchemaVersion]), constructor),
// 		Description: "Data source for a local alpine repository",
// 	}
// }
