package remote

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

func NewHexRemoteRepositoryDataSource() datasource.DataSource {
	return &HexRemoteRepositoryDataSource{}
}

type HexRemoteRepositoryDataSource struct {
	ProviderData util.ProviderMetadata
}

func (d *HexRemoteRepositoryDataSource) SetProviderData(providerData util.ProviderMetadata) {
	d.ProviderData = providerData
}

type HexRemoteRepositoryDataSourceModel struct {
	BaseRemoteRepositoryDataSourceModel
	HexPublicKey         types.String `tfsdk:"public_key_ref"`
	HexPrimaryKeyPairRef types.String `tfsdk:"hex_primary_keypair_ref"`
	Curated              types.Bool   `tfsdk:"curated"`
}

type HexRemoteRepositoryAPIModel struct {
	BaseRemoteRepositoryAPIModel
	HexPublicKey         string `json:"hexPublicKey"`
	HexPrimaryKeyPairRef string `json:"primaryKeyPairRef"`
	Curated              bool   `json:"curated"`
}

func (d *HexRemoteRepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_remote_hex_repository"
}

func (d *HexRemoteRepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for a remote hex repository",
		Attributes: lo.Assign(
			BaseRemoteSchemaAttributes(),
			map[string]schema.Attribute{
				"public_key_ref": schema.StringAttribute{
					MarkdownDescription: "Contains the public key used when downloading packages from the Hex remote registry (public, private, or self-hosted Hex server).",
					Computed:            true,
				},
				"hex_primary_keypair_ref": schema.StringAttribute{
					MarkdownDescription: "Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.",
					Computed:            true,
				},
				"curated": schema.BoolAttribute{
					MarkdownDescription: "Enable repository to be protected by the Curation service.",
					Computed:            true,
				},
			},
		),
	}
}

func (d *HexRemoteRepositoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (d *HexRemoteRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HexRemoteRepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repo HexRemoteRepositoryAPIModel
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

func (d *HexRemoteRepositoryDataSourceModel) FromAPIModel(ctx context.Context, apiModel HexRemoteRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Convert common fields using the base remote repository function
	diags.Append(CommonRemoteFromAPIModel(ctx, &d.BaseRemoteRepositoryDataSourceModel, apiModel.BaseRemoteRepositoryAPIModel)...)
	if diags.HasError() {
		return diags
	}

	// Convert Hex-specific fields
	d.HexPublicKey = types.StringValue(apiModel.HexPublicKey)
	d.HexPrimaryKeyPairRef = types.StringValue(apiModel.HexPrimaryKeyPairRef)
	d.Curated = types.BoolValue(apiModel.Curated)

	return diags
}
