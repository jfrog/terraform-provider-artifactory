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

package configuration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	"gopkg.in/yaml.v3"
)

type TrashCanConfigAPIModel struct {
	Enabled             bool  `xml:"enabled" yaml:"enabled"`
	RetentionPeriodDays int64 `xml:"retentionPeriodDays" yaml:"retentionPeriodDays"`
}

type TrashCanConfig struct {
	TrashcanConfig *TrashCanConfigAPIModel `xml:"trashcanConfig"`
}

type TrashCanConfigResourceModel struct {
	Enabled             types.Bool  `tfsdk:"enabled"`
	RetentionPeriodDays types.Int64 `tfsdk:"retention_period_days"`
}

func (r *TrashCanConfigResourceModel) ToAPIModel(_ context.Context, apiModel *TrashCanConfigAPIModel) diag.Diagnostics {
	*apiModel = TrashCanConfigAPIModel{
		Enabled:             r.Enabled.ValueBool(),
		RetentionPeriodDays: r.RetentionPeriodDays.ValueInt64(),
	}

	return nil
}

func (r *TrashCanConfigResourceModel) FromAPIModel(_ context.Context, apiModel *TrashCanConfigAPIModel) diag.Diagnostics {
	r.Enabled = types.BoolValue(apiModel.Enabled)
	r.RetentionPeriodDays = types.Int64Value(apiModel.RetentionPeriodDays)

	return nil
}

func NewTrashCanConfigResource() resource.Resource {
	return &TrashCanConfigResource{
		TypeName: "artifactory_trashcan_config",
	}
}

type TrashCanConfigResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

func (r *TrashCanConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *TrashCanConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory Trash Can configuration resource. This resource configuration corresponds to 'trashcanConfig' config block in system configuration XML (REST endpoint: artifactory/api/system/configuration). Manages the trash can settings of the Artifactory instance. When enabled, deleted items are stored in the trash can for the specified retention period before being permanently deleted.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "If set, trash can will be enabled and deleted items will be stored in the trash can for the specified retention period.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"retention_period_days": schema.Int64Attribute{
				MarkdownDescription: "The number of days to keep deleted items in the trash can before deleting permanently. Default value is `14`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(14),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
		},
	}
}

func (r *TrashCanConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *TrashCanConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan *TrashCanConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var trashCanConfig TrashCanConfigAPIModel
	resp.Diagnostics.Append(plan.ToAPIModel(ctx, &trashCanConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var constructBody = map[string]TrashCanConfigAPIModel{}
	constructBody["trashcanConfig"] = trashCanConfig
	content, err := yaml.Marshal(&constructBody)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData.Client)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TrashCanConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state *TrashCanConfigResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var trashCanConfig TrashCanConfig
	res, err := r.ProviderData.Client.R().
		SetResult(&trashCanConfig).
		Get(ConfigurationEndpoint)
	if err != nil || res.IsError() {
		utilfw.UnableToRefreshResourceError(resp, "failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		return
	}

	if trashCanConfig.TrashcanConfig == nil {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("enabled"),
			"no trash can configuration found",
			"",
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.FromAPIModel(ctx, trashCanConfig.TrashcanConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TrashCanConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan *TrashCanConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var trashCanConfig TrashCanConfigAPIModel
	resp.Diagnostics.Append(plan.ToAPIModel(ctx, &trashCanConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var constructBody = map[string]TrashCanConfigAPIModel{}
	constructBody["trashcanConfig"] = trashCanConfig
	content, err := yaml.Marshal(&constructBody)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData.Client)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	resp.Diagnostics.Append(plan.FromAPIModel(ctx, &trashCanConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TrashCanConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state TrashCanConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Trash can config is a required block in Artifactory system configuration
	// and cannot be removed. Reset to defaults (enabled with 14 day retention).
	defaultTrashCanConfig := TrashCanConfigAPIModel{
		Enabled:             true,
		RetentionPeriodDays: 14,
	}
	var constructBody = map[string]TrashCanConfigAPIModel{}
	constructBody["trashcanConfig"] = defaultTrashCanConfig
	content, err := yaml.Marshal(&constructBody)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData.Client)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *TrashCanConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Trash can config is a singleton — the import ID is unused.
	// Set placeholder values so the subsequent Read can populate real state from the API.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), types.BoolValue(true))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("retention_period_days"), types.Int64Value(14))...)
}
