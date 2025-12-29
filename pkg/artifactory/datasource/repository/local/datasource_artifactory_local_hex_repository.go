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

package local

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	datasource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

var _ datasource.DataSource = &LocalHexRepositoryDataSource{}

func NewLocalHexRepositoryDataSource() datasource.DataSource {
	return &LocalHexRepositoryDataSource{}
}

type LocalHexRepositoryDataSource struct {
	ProviderData util.ProviderMetadata
}

type LocalHexRepositoryDataSourceModel struct {
	Key                    types.String `tfsdk:"key"`
	ProjectKey             types.String `tfsdk:"project_key"`
	ProjectEnvironments    types.Set    `tfsdk:"project_environments"`
	Description            types.String `tfsdk:"description"`
	Notes                  types.String `tfsdk:"notes"`
	IncludesPattern        types.String `tfsdk:"includes_pattern"`
	ExcludesPattern        types.String `tfsdk:"excludes_pattern"`
	RepoLayoutRef          types.String `tfsdk:"repo_layout_ref"`
	BlackedOut             types.Bool   `tfsdk:"blacked_out"`
	XrayIndex              types.Bool   `tfsdk:"xray_index"`
	PropertySets           types.Set    `tfsdk:"property_sets"`
	ArchiveBrowsingEnabled types.Bool   `tfsdk:"archive_browsing_enabled"`
	DownloadDirect         types.Bool   `tfsdk:"download_direct"`
	PriorityResolution     types.Bool   `tfsdk:"priority_resolution"`
	CDNRedirect            types.Bool   `tfsdk:"cdn_redirect"`
	PackageType            types.String `tfsdk:"package_type"`
	HexPrimaryKeyPairRef   types.String `tfsdk:"hex_primary_keypair_ref"`
}

type LocalHexRepositoryAPIModel struct {
	repository.BaseAPIModel
	local.LocalAPIModel
	HexPrimaryKeyPairRef string `json:"primaryKeyPairRef"`
}

func (d *LocalHexRepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_local_hex_repository"
}

func (d *LocalHexRepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := lo.Assign(
		LocalDataSourceAttributes,
		datasource_repository.HexDataSourceAttributes,
	)

	resp.Schema = schema.Schema{
		Attributes:  attributes,
		Description: "Provides a data source for a local Hex repository",
	}
}

func (d *LocalHexRepositoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (d *LocalHexRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LocalHexRepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiModel LocalHexRepositoryAPIModel
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

func (m *LocalHexRepositoryDataSourceModel) FromAPIModel(ctx context.Context, apiModel *LocalHexRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Base fields
	m.Key = types.StringValue(apiModel.Key)
	m.ProjectKey = types.StringValue(apiModel.ProjectKey)
	m.Description = types.StringValue(apiModel.Description)
	m.Notes = types.StringValue(apiModel.Notes)
	m.IncludesPattern = types.StringValue(apiModel.IncludesPattern)
	m.ExcludesPattern = types.StringValue(apiModel.ExcludesPattern)
	m.PackageType = types.StringValue(repository.HexPackageType)

	// Local-specific fields
	m.RepoLayoutRef = types.StringValue(apiModel.RepoLayoutRef)
	m.BlackedOut = types.BoolValue(apiModel.LocalAPIModel.BlackedOut)
	m.XrayIndex = types.BoolValue(apiModel.LocalAPIModel.XrayIndex)
	m.ArchiveBrowsingEnabled = types.BoolValue(apiModel.LocalAPIModel.ArchiveBrowsingEnabled)
	m.DownloadDirect = types.BoolValue(apiModel.LocalAPIModel.DownloadRedirect)
	m.PriorityResolution = types.BoolValue(apiModel.LocalAPIModel.PriorityResolution)
	m.CDNRedirect = types.BoolValue(apiModel.LocalAPIModel.CDNRedirect)

	// Hex-specific field
	m.HexPrimaryKeyPairRef = types.StringValue(apiModel.HexPrimaryKeyPairRef)

	// Property sets
	var propertySets []types.String
	for _, ps := range apiModel.LocalAPIModel.PropertySets {
		propertySets = append(propertySets, types.StringValue(ps))
	}
	if len(propertySets) > 0 {
		psSet, d := types.SetValueFrom(ctx, types.StringType, propertySets)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}
		m.PropertySets = psSet
	} else {
		m.PropertySets = types.SetNull(types.StringType)
	}

	// Project environments
	var projectEnvironments []types.String
	for _, env := range apiModel.ProjectEnvironments {
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
