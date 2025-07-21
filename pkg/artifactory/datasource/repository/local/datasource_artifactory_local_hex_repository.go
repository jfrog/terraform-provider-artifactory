package local

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

func NewHexLocalRepositoryDataSource() datasource.DataSource {
	return &HexLocalRepositoryDataSource{}
}

type HexLocalRepositoryDataSource struct {
	ProviderData util.ProviderMetadata
}

func (d *HexLocalRepositoryDataSource) SetProviderData(providerData util.ProviderMetadata) {
	d.ProviderData = providerData
}

type HexLocalRepositoryDataSourceModel struct {
	repository.BaseRepositoryDataSourceModel
	HexPrimaryKeyPairRef types.String `tfsdk:"hex_primary_keypair_ref"`
}

type HexLocalRepositoryAPIModel struct {
	repository.BaseRepositoryAPIModel
	HexPrimaryKeyPairRef string `json:"primaryKeyPairRef"`
}

func (d *HexLocalRepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_local_hex_repository"
}

func (d *HexLocalRepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for a local hex repository",
		Attributes: lo.Assign(
			repository.BaseSchemaAttributes(),
			map[string]schema.Attribute{
				"hex_primary_keypair_ref": schema.StringAttribute{
					MarkdownDescription: "Reference to the RSA key pair used to sign Hex repository index files.",
					Computed:            true,
				},
			},
		),
	}
}

func (d *HexLocalRepositoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (d *HexLocalRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HexLocalRepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repo HexLocalRepositoryAPIModel
	response, err := d.ProviderData.Client.R().
		SetResult(&repo).
		Get("artifactory/api/repositories/" + data.Key.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetching the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if response.IsError() {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetching the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+response.String(),
		)
		return
	}

	// Convert from the API data model to the Terraform data model
	resp.Diagnostics.Append(data.FromAPIModel(ctx, repo)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *HexLocalRepositoryDataSourceModel) FromAPIModel(ctx context.Context, apiModel HexLocalRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Convert common fields using the base repository function
	diags.Append(repository.CommonFromAPIModel(ctx, &d.BaseRepositoryDataSourceModel, apiModel.BaseRepositoryAPIModel)...)
	if diags.HasError() {
		return diags
	}

	// Convert Hex-specific fields
	d.HexPrimaryKeyPairRef = types.StringValue(apiModel.HexPrimaryKeyPairRef)

	return diags
}
