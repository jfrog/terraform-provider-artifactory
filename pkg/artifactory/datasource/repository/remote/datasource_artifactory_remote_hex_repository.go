package remote

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	datasource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

var _ datasource.DataSource = &RemoteHexRepositoryDataSource{}

func NewRemoteHexRepositoryDataSource() datasource.DataSource {
	return &RemoteHexRepositoryDataSource{}
}

type RemoteHexRepositoryDataSource struct {
	ProviderData util.ProviderMetadata
}

type RemoteHexRepositoryDataSourceModel struct {
	BaseRemoteRepositoryDataSourceModel
	HexPrimaryKeyPairRef types.String `tfsdk:"hex_primary_keypair_ref"`
	PublicKey            types.String `tfsdk:"public_key"`
}

type RemoteHexRepositoryAPIModel struct {
	remote.RemoteAPIModel
	HexPrimaryKeyPairRef string `json:"primaryKeyPairRef"`
	PublicKey            string `json:"hexPublicKey"`
}

func (d *RemoteHexRepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_remote_hex_repository"
}

func (d *RemoteHexRepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := lo.Assign(
		BaseRemoteSchemaAttributes(),
		datasource_repository.HexRemoteDataSourceAttributes,
	)

	resp.Schema = schema.Schema{
		Attributes:  attributes,
		Description: "Provides a data source for a remote Hex repository",
	}
}

func (d *RemoteHexRepositoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (d *RemoteHexRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RemoteHexRepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiModel RemoteHexRepositoryAPIModel
	var jfrogErrors util.JFrogErrors

	response, err := d.ProviderData.Client.R().
		SetPathParam("key", data.Key.ValueString()).
		SetResult(&apiModel).
		SetError(&jfrogErrors).
		Get(repository.RepositoriesEndpoint)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetching the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if response.StatusCode() == http.StatusBadRequest || response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetching the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+jfrogErrors.String(),
		)
		return
	}

	// Convert from API model to Terraform model
	data.FromAPIModel(ctx, &apiModel)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m *RemoteHexRepositoryDataSourceModel) FromAPIModel(ctx context.Context, apiModel *RemoteHexRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Convert base repository fields
	baseAPIModel := datasource_repository.BaseRepositoryAPIModel{
		Key:                 apiModel.RemoteAPIModel.Key,
		ProjectKey:          apiModel.RemoteAPIModel.ProjectKey,
		ProjectEnvironments: apiModel.RemoteAPIModel.ProjectEnvironments,
		Description:         apiModel.RemoteAPIModel.Description,
		Notes:               apiModel.RemoteAPIModel.Notes,
		IncludesPattern:     apiModel.RemoteAPIModel.IncludesPattern,
		ExcludesPattern:     apiModel.RemoteAPIModel.ExcludesPattern,
		RepoLayoutRef:       apiModel.RemoteAPIModel.RepoLayoutRef,
		PackageType:         repository.HexPackageType,
	}
	diags.Append(datasource_repository.CommonFromAPIModel(ctx, &m.BaseRepositoryDataSourceModel, baseAPIModel)...)
	if diags.HasError() {
		return diags
	}

	// Convert remote-specific fields
	remoteAPIModel := BaseRemoteRepositoryAPIModel{
		BaseRepositoryAPIModel: baseAPIModel,
		RemoteAPIModel:         apiModel.RemoteAPIModel,
	}
	diags.Append(CommonRemoteFromAPIModel(ctx, &m.BaseRemoteRepositoryDataSourceModel, remoteAPIModel)...)
	if diags.HasError() {
		return diags
	}

	// Hex-specific fields
	m.HexPrimaryKeyPairRef = types.StringValue(apiModel.HexPrimaryKeyPairRef)
	m.PublicKey = types.StringValue(apiModel.PublicKey)

	return diags
}
