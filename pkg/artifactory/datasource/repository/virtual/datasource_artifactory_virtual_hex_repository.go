package virtual

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

func NewHexVirtualRepositoryDataSource() datasource.DataSource {
	return &HexVirtualRepositoryDataSource{}
}

type HexVirtualRepositoryDataSource struct {
	ProviderData util.ProviderMetadata
}

func (d *HexVirtualRepositoryDataSource) SetProviderData(providerData util.ProviderMetadata) {
	d.ProviderData = providerData
}

type HexVirtualRepositoryDataSourceModel struct {
	BaseVirtualRepositoryDataSourceModel
	HexPrimaryKeyPairRef types.String `tfsdk:"hex_primary_keypair_ref"`
}

type HexVirtualRepositoryAPIModel struct {
	BaseVirtualRepositoryAPIModel
	HexPrimaryKeyPairRef string `json:"primaryKeyPairRef"`
}

func (d *HexVirtualRepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_virtual_hex_repository"
}

func (d *HexVirtualRepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for a virtual hex repository",
		Attributes: lo.Assign(
			BaseVirtualSchemaAttributes(),
			map[string]schema.Attribute{
				"hex_primary_keypair_ref": schema.StringAttribute{
					MarkdownDescription: "Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.",
					Computed:            true,
				},
			},
		),
	}
}

func (d *HexVirtualRepositoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (d *HexVirtualRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HexVirtualRepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repo HexVirtualRepositoryAPIModel
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

func (d *HexVirtualRepositoryDataSourceModel) FromAPIModel(ctx context.Context, apiModel HexVirtualRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Convert common fields using the base virtual repository function
	diags.Append(CommonVirtualFromAPIModel(ctx, &d.BaseVirtualRepositoryDataSourceModel, apiModel.BaseVirtualRepositoryAPIModel)...)
	if diags.HasError() {
		return diags
	}

	// Convert Hex-specific fields
	d.HexPrimaryKeyPairRef = types.StringValue(apiModel.HexPrimaryKeyPairRef)

	return diags
}
