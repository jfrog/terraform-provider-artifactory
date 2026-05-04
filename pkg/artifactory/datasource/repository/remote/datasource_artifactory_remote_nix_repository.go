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

package remote

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	datasource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	resourceremote "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/util"
)

var _ datasource.DataSource = &RemoteNixRepositoryDataSource{}

func NewRemoteNixRepositoryDataSource() datasource.DataSource {
	return &RemoteNixRepositoryDataSource{}
}

type RemoteNixRepositoryDataSource struct {
	ProviderData util.ProviderMetadata
}

type RemoteNixRepositoryDataSourceModel struct {
	BaseRemoteRepositoryDataSourceModel
}

type RemoteNixRepositoryAPIModel struct {
	resourceremote.RemoteAPIModel
}

func (d *RemoteNixRepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_remote_nix_repository"
}

func (d *RemoteNixRepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:  BaseRemoteSchemaAttributes(),
		Description: "Provides a data source for a remote Nix repository",
	}
}

func (d *RemoteNixRepositoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (d *RemoteNixRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RemoteNixRepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiModel RemoteNixRepositoryAPIModel
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

func (m *RemoteNixRepositoryDataSourceModel) FromAPIModel(ctx context.Context, apiModel *RemoteNixRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	baseAPIModel := datasource_repository.BaseRepositoryAPIModel{
		Key:                 apiModel.RemoteAPIModel.Key,
		ProjectKey:          apiModel.RemoteAPIModel.ProjectKey,
		ProjectEnvironments: apiModel.RemoteAPIModel.ProjectEnvironments,
		Description:         apiModel.RemoteAPIModel.Description,
		Notes:               apiModel.RemoteAPIModel.Notes,
		IncludesPattern:     apiModel.RemoteAPIModel.IncludesPattern,
		ExcludesPattern:     apiModel.RemoteAPIModel.ExcludesPattern,
		RepoLayoutRef:       apiModel.RemoteAPIModel.RepoLayoutRef,
		PackageType:         repository.NixPackageType,
	}
	diags.Append(datasource_repository.CommonFromAPIModel(ctx, &m.BaseRepositoryDataSourceModel, baseAPIModel)...)
	if diags.HasError() {
		return diags
	}

	remoteAPIModel := BaseRemoteRepositoryAPIModel{
		BaseRepositoryAPIModel: baseAPIModel,
		RemoteAPIModel:         apiModel.RemoteAPIModel,
	}
	diags.Append(CommonRemoteFromAPIModel(ctx, &m.BaseRemoteRepositoryDataSourceModel, remoteAPIModel)...)
	if diags.HasError() {
		return diags
	}

	return diags
}
