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
	"github.com/hashicorp/terraform-plugin-framework/types"
	datasource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	resourceremote "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

var _ datasource.DataSource = &RemoteAIEditorExtensionsRepositoryDataSource{}

func NewRemoteAIEditorExtensionsRepositoryDataSource() datasource.DataSource {
	return &RemoteAIEditorExtensionsRepositoryDataSource{}
}

type RemoteAIEditorExtensionsRepositoryDataSource struct {
	ProviderData util.ProviderMetadata
}

type RemoteAIEditorExtensionsRepositoryDataSourceModel struct {
	BaseRemoteRepositoryDataSourceModel
	ExternalDependenciesEnabled  types.Bool `tfsdk:"external_dependencies_enabled"`
	ExternalDependenciesPatterns types.List `tfsdk:"external_dependencies_patterns"`
	EnableTokenAuthentication    types.Bool `tfsdk:"enable_token_authentication"`
	PropagateQueryParams         types.Bool `tfsdk:"propagate_query_params"`
	RetrieveSha256FromServer     types.Bool `tfsdk:"retrieve_sha256_from_server"`
	Curated                      types.Bool `tfsdk:"curated"`
	PassThrough                  types.Bool `tfsdk:"pass_through"`
}

// custom_http_headers is intentionally absent: Artifactory returns header values
// in plaintext, and a data source has no configuration to compare them against,
// so exposing them would place secrets in state.
type RemoteAIEditorExtensionsRepositoryAPIModel struct {
	resourceremote.RemoteAPIModel
	ExternalDependenciesEnabled  bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns,omitempty"`
	PropagateQueryParams         bool     `json:"propagateQueryParams"`
	RetrieveSha256FromServer     bool     `json:"retrieveSha256FromServer"`
	Curated                      bool     `json:"curated"`
	PassThrough                  bool     `json:"passThrough"`
}

func (d *RemoteAIEditorExtensionsRepositoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_remote_aieditorextensions_repository"
}

func (d *RemoteAIEditorExtensionsRepositoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: lo.Assign(
			BaseRemoteSchemaAttributes(),
			map[string]schema.Attribute{
				"external_dependencies_enabled": schema.BoolAttribute{
					MarkdownDescription: "When set, Artifactory can resolve extension dependencies from the external sources matching `external_dependencies_patterns`.",
					Computed:            true,
				},
				"external_dependencies_patterns": schema.ListAttribute{
					ElementType:         types.StringType,
					MarkdownDescription: "An allow list of Ant-style path patterns that determine which remote hosts external extension dependencies may be downloaded from.",
					Computed:            true,
				},
				"enable_token_authentication": schema.BoolAttribute{
					MarkdownDescription: "Whether token (Bearer) based authentication is enabled.",
					Computed:            true,
				},
				"propagate_query_params": schema.BoolAttribute{
					MarkdownDescription: "Whether query params included in the request to Artifactory are passed on to the remote repository.",
					Computed:            true,
				},
				"retrieve_sha256_from_server": schema.BoolAttribute{
					MarkdownDescription: "Whether Artifactory retrieves the SHA256 from the remote server when it is not cached in the remote repo.",
					Computed:            true,
				},
				"curated": schema.BoolAttribute{
					MarkdownDescription: "Whether the repository is protected by the Curation service.",
					Computed:            true,
				},
				"pass_through": schema.BoolAttribute{
					MarkdownDescription: "Whether Pass-through for Curation Audit is enabled.",
					Computed:            true,
				},
			},
		),
		Description: "Provides a data source for a remote AI-Editor Extensions repository",
	}
}

func (d *RemoteAIEditorExtensionsRepositoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (d *RemoteAIEditorExtensionsRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RemoteAIEditorExtensionsRepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiModel RemoteAIEditorExtensionsRepositoryAPIModel
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

func (m *RemoteAIEditorExtensionsRepositoryDataSourceModel) FromAPIModel(ctx context.Context, apiModel *RemoteAIEditorExtensionsRepositoryAPIModel) diag.Diagnostics {
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
		PackageType:         repository.AIEditorExtensionsPackageType,
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

	m.ExternalDependenciesEnabled = types.BoolValue(apiModel.ExternalDependenciesEnabled)
	m.EnableTokenAuthentication = types.BoolValue(apiModel.EnableTokenAuthentication)
	m.PropagateQueryParams = types.BoolValue(apiModel.PropagateQueryParams)
	m.RetrieveSha256FromServer = types.BoolValue(apiModel.RetrieveSha256FromServer)
	m.Curated = types.BoolValue(apiModel.Curated)
	m.PassThrough = types.BoolValue(apiModel.PassThrough)

	externalDependenciesPatterns, d := types.ListValueFrom(ctx, types.StringType, apiModel.ExternalDependenciesPatterns)
	if d != nil {
		diags.Append(d...)
	}
	m.ExternalDependenciesPatterns = externalDependenciesPatterns

	return diags
}
