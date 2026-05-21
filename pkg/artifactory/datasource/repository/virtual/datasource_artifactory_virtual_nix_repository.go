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
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	resourcevirtual "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/jfrog/terraform-provider-shared/util"
)

var _ datasource.DataSource = &VirtualNixRepositoryDataSource{}

func NewVirtualNixRepositoryDataSource() datasource.DataSource {
	return &VirtualNixRepositoryDataSource{}
}

type VirtualNixRepositoryDataSource struct {
	ProviderData util.ProviderMetadata
}

type VirtualNixRepositoryDataSourceModel struct {
	Key                 types.String `tfsdk:"key"`
	ProjectKey          types.String `tfsdk:"project_key"`
	ProjectEnvironments types.Set    `tfsdk:"project_environments"`
	Description         types.String `tfsdk:"description"`
	Notes               types.String `tfsdk:"notes"`
	IncludesPattern     types.String `tfsdk:"includes_pattern"`
	ExcludesPattern     types.String `tfsdk:"excludes_pattern"`
	PackageType         types.String `tfsdk:"package_type"`
	RepoLayoutRef       types.String `tfsdk:"repo_layout_ref"`
}

type VirtualNixRepositoryAPIModel struct {
	resourcevirtual.VirtualAPIModel
}

func (d *VirtualNixRepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_virtual_nix_repository"
}

func (d *VirtualNixRepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:  VirtualDataSourceAttributes,
		Description: "Provides a data source for a virtual Nix repository",
	}
}

func (d *VirtualNixRepositoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (d *VirtualNixRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VirtualNixRepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiModel VirtualNixRepositoryAPIModel
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

	data.FromAPIModel(ctx, &apiModel)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m *VirtualNixRepositoryDataSourceModel) FromAPIModel(ctx context.Context, apiModel *VirtualNixRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	m.Key = types.StringValue(apiModel.VirtualAPIModel.Key)
	m.ProjectKey = types.StringValue(apiModel.VirtualAPIModel.ProjectKey)
	m.Description = types.StringValue(apiModel.VirtualAPIModel.Description)
	m.Notes = types.StringValue(apiModel.VirtualAPIModel.Notes)
	m.IncludesPattern = types.StringValue(apiModel.VirtualAPIModel.IncludesPattern)
	m.ExcludesPattern = types.StringValue(apiModel.VirtualAPIModel.ExcludesPattern)
	m.PackageType = types.StringValue(repository.NixPackageType)
	m.RepoLayoutRef = types.StringValue(apiModel.VirtualAPIModel.RepoLayoutRef)

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
