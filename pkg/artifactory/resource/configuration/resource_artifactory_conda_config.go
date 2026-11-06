// Copyright (c) JFrog Ltd. (2026)
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
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
)

const condaConfigEndpoint = "artifactory/ui/generalConfig"

var _ resource.Resource = (*CondaConfigResource)(nil)

// NewCondaConfigResource returns a resource for the global Conda package settings.
func NewCondaConfigResource() resource.Resource {
	return &CondaConfigResource{
		JFrogResource: util.JFrogResource{TypeName: "artifactory_conda_config"},
	}
}

// CondaConfigResource manages the instance-wide Conda package settings.
type CondaConfigResource struct {
	util.JFrogResource
}

// CondaConfigResourceModel stores the singleton ID and shard flag in Terraform state.
type CondaConfigResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ShardsEnabled types.Bool   `tfsdk:"shards_enabled"`
}

// Schema defines the singleton identity and required shard setting.
func (r *CondaConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the instance-wide Conda settings under Administration > Artifactory Settings > Packages Settings. Define only one resource per Artifactory instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the global Conda configuration, always `conda`.",
			},
			"shards_enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Enables Conda metadata sharding for the entire Artifactory instance. Existing local repositories must be reindexed after enabling sharding. Legacy repodata endpoints remain available.",
			},
		},
	}
}

// Create adopts the global configuration and applies the planned shard setting.
func (r *CondaConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan CondaConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.setShardsEnabled(ctx, plan.ShardsEnabled.ValueBool()); err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	plan.ID = types.StringValue("conda")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the global shard setting in Terraform state.
func (r *CondaConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state CondaConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, conda, err := r.readConfig(ctx)
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	state.ID = types.StringValue("conda")
	state.ShardsEnabled = types.BoolValue(conda["shardsEnabled"].(bool))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update applies the planned shard setting and verifies its saved value.
func (r *CondaConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan CondaConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.setShardsEnabled(ctx, plan.ShardsEnabled.ValueBool()); err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	plan.ID = types.StringValue("conda")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete stops managing the global setting without changing Artifactory.
func (r *CondaConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	resp.Diagnostics.AddWarning(
		"Conda configuration cannot be deleted",
		"Removing this resource from Terraform leaves the global Conda settings unchanged. Apply shards_enabled = false before removing the resource to disable sharding.",
	)
}

// ImportState accepts the singleton import ID "conda".
func (r *CondaConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "conda" {
		resp.Diagnostics.AddError("Invalid Conda configuration import ID", "The global Conda configuration must be imported using the ID 'conda'.")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readConfig loads a fresh configuration and validates the exposed shard flag.
func (r *CondaConfigResource) readConfig(ctx context.Context) (map[string]interface{}, map[string]interface{}, error) {
	response, err := r.ProviderData.Client.R().SetContext(ctx).Get(condaConfigEndpoint)
	if err != nil {
		return nil, nil, err
	}
	if response.IsError() {
		return nil, nil, fmt.Errorf("unable to read Conda configuration: %s", response.Status())
	}

	// Preserve unknown settings and exact numbers when saving the full UI configuration.
	var config map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(response.Body()))
	decoder.UseNumber()
	if err := decoder.Decode(&config); err != nil {
		return nil, nil, fmt.Errorf("unable to decode Artifactory general configuration: %w", err)
	}

	repositories, ok := config["repositoriesConfig"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("general configuration does not expose repositoriesConfig")
	}
	conda, ok := repositories["conda"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("global Conda package settings are not exposed by Artifactory")
	}
	if _, ok := conda["shardsEnabled"].(bool); !ok {
		return nil, nil, fmt.Errorf("repositoriesConfig.conda.shardsEnabled is not exposed as a boolean; a compatible Artifactory version with Conda sharding support is required")
	}
	return config, conda, nil
}

// setShardsEnabled preserves unrelated settings and verifies the flag after a write.
func (r *CondaConfigResource) setShardsEnabled(ctx context.Context, enabled bool) error {
	config, conda, err := r.readConfig(ctx)
	if err != nil {
		return err
	}
	if conda["shardsEnabled"].(bool) == enabled {
		return nil
	}

	conda["shardsEnabled"] = enabled
	response, err := r.ProviderData.Client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(config).
		Put(condaConfigEndpoint)
	if err != nil {
		return err
	}
	if response.IsError() {
		return fmt.Errorf("unable to update Conda configuration: %s", response.Status())
	}

	_, saved, err := r.readConfig(ctx)
	if err != nil {
		return err
	}
	if saved["shardsEnabled"].(bool) != enabled {
		return fmt.Errorf("server did not persist the requested Conda shardsEnabled setting")
	}
	return nil
}
