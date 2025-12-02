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

package webhook

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jfrog/terraform-provider-shared/util"
)

var _ resource.Resource = &ReleaseBundleCustomWebhookResource{}

func NewArtifactoryReleaseBundleCustomWebhookResource() resource.Resource {
	return &ReleaseBundleCustomWebhookResource{
		CustomWebhookResource: CustomWebhookResource{
			WebhookResource: WebhookResource{
				TypeName:    fmt.Sprintf("artifactory_%s_custom_webhook", ArtifactoryReleaseBundleDomain),
				Domain:      ArtifactoryReleaseBundleDomain,
				Description: "Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.:",
			},
		},
	}
}

func NewDestinationCustomWebhookResource() resource.Resource {
	return &ReleaseBundleCustomWebhookResource{
		CustomWebhookResource: CustomWebhookResource{
			WebhookResource: WebhookResource{
				TypeName:    fmt.Sprintf("artifactory_%s_custom_webhook", DestinationDomain),
				Domain:      DestinationDomain,
				Description: "Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.:",
			},
		},
	}
}

func NewDistributionCustomWebhookResource() resource.Resource {
	return &ReleaseBundleCustomWebhookResource{
		CustomWebhookResource: CustomWebhookResource{
			WebhookResource: WebhookResource{
				TypeName:    fmt.Sprintf("artifactory_%s_custom_webhook", DistributionDomain),
				Domain:      DistributionDomain,
				Description: "Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.:",
			},
		},
	}
}

func NewReleaseBundleCustomWebhookResource() resource.Resource {
	return &ReleaseBundleCustomWebhookResource{
		CustomWebhookResource: CustomWebhookResource{
			WebhookResource: WebhookResource{
				TypeName: fmt.Sprintf("artifactory_%s_custom_webhook", ReleaseBundleDomain),
				Domain:   ReleaseBundleDomain,
				Description: "Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.\n\n" +
					"!>This resource is being deprecated and replaced by `artifactory_destination_webhook` resource.",
			},
		},
	}
}

type ReleaseBundleCustomWebhookResourceModel struct {
	CustomWebhookResourceModel
}

type ReleaseBundleCustomWebhookResource struct {
	CustomWebhookResource
}

func (r *ReleaseBundleCustomWebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	r.WebhookResource.Metadata(ctx, req, resp)
}

func (r *ReleaseBundleCustomWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.CreateSchema(r.Domain, &releaseBundleCriteriaBlock)
	if r.Domain == ReleaseBundleDomain {
		resp.Schema.DeprecationMessage = "This resource is being deprecated and replaced by artifactory_destination_webhook resource"
	}
}

func (r *ReleaseBundleCustomWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.WebhookResource.Configure(ctx, req, resp)
}

func (r ReleaseBundleCustomWebhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ReleaseBundleCustomWebhookResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	releaseBundleValidateConfig(data.Criteria, resp)
}

func (r *ReleaseBundleCustomWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ReleaseBundleCustomWebhookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook CustomWebhookAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, r.Domain, &webhook)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.CustomWebhookResource.Create(ctx, webhook, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ReleaseBundleCustomWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleCustomWebhookResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook CustomWebhookAPIModel
	found := r.CustomWebhookResource.Read(ctx, state.Key.ValueString(), &webhook, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		return
	}

	resp.Diagnostics.Append(state.fromAPIModel(ctx, webhook, state.Handlers)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ReleaseBundleCustomWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ReleaseBundleCustomWebhookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook CustomWebhookAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, r.Domain, &webhook)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.CustomWebhookResource.Update(ctx, plan.Key.ValueString(), webhook, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ReleaseBundleCustomWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleCustomWebhookResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	r.WebhookResource.Delete(ctx, state.Key.ValueString(), resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ReleaseBundleCustomWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.WebhookResource.ImportState(ctx, req, resp)
}

func (m ReleaseBundleCustomWebhookResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *CustomWebhookAPIModel) (diags diag.Diagnostics) {
	criteriaObj := m.Criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	baseCriteria, d := m.CustomWebhookResourceModel.toBaseCriteriaAPIModel(ctx, criteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaAPIModel, d := toReleaseBundleCriteriaAPIModel(ctx, baseCriteria, criteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.CustomWebhookResourceModel.toAPIModel(ctx, domain, criteriaAPIModel, apiModel)
	if d.HasError() {
		diags.Append(d...)
	}

	return
}

func (m *ReleaseBundleCustomWebhookResourceModel) fromAPIModel(ctx context.Context, apiModel CustomWebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	criteriaAPIModel := apiModel.EventFilter.Criteria.(map[string]interface{})

	baseCriteriaAttrs, d := m.CustomWebhookResourceModel.fromBaseCriteriaAPIModel(ctx, criteriaAPIModel)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaSet, d := fromReleaseBundleCriteriaAPIModel(ctx, criteriaAPIModel, baseCriteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.CustomWebhookResourceModel.fromAPIModel(ctx, apiModel, stateHandlers, &criteriaSet)
	if d.HasError() {
		diags.Append(d...)
	}

	return diags
}
