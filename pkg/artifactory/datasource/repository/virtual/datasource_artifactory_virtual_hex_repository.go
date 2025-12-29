// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package virtual

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	datasource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

var _ datasource.DataSource = &VirtualHexRepositoryDataSource{}

func NewVirtualHexRepositoryDataSource() datasource.DataSource {
	return &VirtualHexRepositoryDataSource{}
}

type VirtualHexRepositoryDataSource struct {
	ProviderData util.ProviderMetadata
}

type VirtualHexRepositoryDataSourceModel struct {
	Key                  types.String `tfsdk:"key"`
	ProjectKey           types.String `tfsdk:"project_key"`
	ProjectEnvironments  types.Set    `tfsdk:"project_environments"`
	Description          types.String `tfsdk:"description"`
	Notes                types.String `tfsdk:"notes"`
	IncludesPattern      types.String `tfsdk:"includes_pattern"`
	ExcludesPattern      types.String `tfsdk:"excludes_pattern"`
	PackageType          types.String `tfsdk:"package_type"`
	RepoLayoutRef        types.String `tfsdk:"repo_layout_ref"`
	HexPrimaryKeyPairRef types.String `tfsdk:"hex_primary_keypair_ref"`
}

type VirtualHexRepositoryAPIModel struct {
	virtual.VirtualAPIModel
	HexPrimaryKeyPairRef string `json:"primaryKeyPairRef"`
}

func (d *VirtualHexRepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_virtual_hex_repository"
}

func (d *VirtualHexRepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := lo.Assign(
		VirtualDataSourceAttributes,
		datasource_repository.HexDataSourceAttributes,
	)

	resp.Schema = schema.Schema{
		Attributes:  attributes,
		Description: "Provides a data source for a virtual Hex repository",
	}
}

func (d *VirtualHexRepositoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (d *VirtualHexRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VirtualHexRepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiModel VirtualHexRepositoryAPIModel
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

func (m *VirtualHexRepositoryDataSourceModel) FromAPIModel(ctx context.Context, apiModel *VirtualHexRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Base fields
	m.Key = types.StringValue(apiModel.VirtualAPIModel.Key)
	m.ProjectKey = types.StringValue(apiModel.VirtualAPIModel.ProjectKey)
	m.Description = types.StringValue(apiModel.VirtualAPIModel.Description)
	m.Notes = types.StringValue(apiModel.VirtualAPIModel.Notes)
	m.IncludesPattern = types.StringValue(apiModel.VirtualAPIModel.IncludesPattern)
	m.ExcludesPattern = types.StringValue(apiModel.VirtualAPIModel.ExcludesPattern)
	m.PackageType = types.StringValue(repository.HexPackageType)
	m.RepoLayoutRef = types.StringValue(apiModel.VirtualAPIModel.RepoLayoutRef)

	// Hex-specific field
	m.HexPrimaryKeyPairRef = types.StringValue(apiModel.HexPrimaryKeyPairRef)

	// Project environments
	var projectEnvironments []types.String
	for _, env := range apiModel.VirtualAPIModel.ProjectEnvironments {
		projectEnvironments = append(projectEnvironments, types.StringValue(env))
	}
	if len(projectEnvironments) > 0 {
		envSet, d := types.SetValueFrom(ctx, types.StringType, projectEnvironments)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}
		m.ProjectEnvironments = envSet
	} else {
		m.ProjectEnvironments = types.SetNull(types.StringType)
	}

	return diags
}
