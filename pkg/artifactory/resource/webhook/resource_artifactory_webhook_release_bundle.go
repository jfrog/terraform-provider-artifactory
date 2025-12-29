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

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

var _ resource.Resource = &ReleaseBundleWebhookResource{}

func NewArtifactoryReleaseBundleWebhookResource() resource.Resource {
	return &ReleaseBundleWebhookResource{
		WebhookResource: WebhookResource{
			TypeName: fmt.Sprintf("artifactory_%s_webhook", ArtifactoryReleaseBundleDomain),
			Domain:   ArtifactoryReleaseBundleDomain,
			Description: "Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.\n\n" +
				"!>This resource is being deprecated and replaced by `artifactory_destination_webhook` resource.",
		},
	}
}

func NewDestinationWebhookResource() resource.Resource {
	return &ReleaseBundleWebhookResource{
		WebhookResource: WebhookResource{
			TypeName:    fmt.Sprintf("artifactory_%s_webhook", DestinationDomain),
			Domain:      DestinationDomain,
			Description: "Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.:",
		},
	}
}

func NewDistributionWebhookResource() resource.Resource {
	return &ReleaseBundleWebhookResource{
		WebhookResource: WebhookResource{
			TypeName:    fmt.Sprintf("artifactory_%s_webhook", DistributionDomain),
			Domain:      DistributionDomain,
			Description: "Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.:",
		},
	}
}

func NewReleaseBundleWebhookResource() resource.Resource {
	return &ReleaseBundleWebhookResource{
		WebhookResource: WebhookResource{
			TypeName:    fmt.Sprintf("artifactory_%s_webhook", ReleaseBundleDomain),
			Domain:      ReleaseBundleDomain,
			Description: "Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.:",
		},
	}
}

type ReleaseBundleWebhookResourceModel struct {
	WebhookResourceModel
}

type ReleaseBundleWebhookResource struct {
	WebhookResource
}

func (r *ReleaseBundleWebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	r.WebhookResource.Metadata(ctx, req, resp)
}

var releaseBundleCriteriaBlock = schema.SetNestedBlock{
	NestedObject: schema.NestedBlockObject{
		Attributes: lo.Assign(
			patternsSchemaAttributes("Simple wildcard patterns for Release Bundle names.\nAnt-style path expressions are supported (*, **, ?).\nFor example: `product_*`"),
			map[string]schema.Attribute{
				"any_release_bundle": schema.BoolAttribute{
					Required:    true,
					Description: "Trigger on any release bundles or distributions",
				},
				"registered_release_bundle_names": schema.SetAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Description: "Trigger on this list of release bundle names",
				},
			},
		),
	},
	Validators: []validator.Set{
		setvalidator.SizeBetween(1, 1),
		setvalidator.IsRequired(),
	},
	Description: "Specifies where the webhook will be applied, on which release bundles or distributions.",
}

func (r *ReleaseBundleWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.CreateSchema(r.Domain, &releaseBundleCriteriaBlock, handlerBlock)
	if r.Domain == ArtifactoryReleaseBundleDomain {
		resp.Schema.DeprecationMessage = "This resource is being deprecated and replaced by artifactory_destination_webhook resource"
	}
}

func (r *ReleaseBundleWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.WebhookResource.Configure(ctx, req, resp)
}

func releaseBundleValidateConfig(criteria basetypes.SetValue, resp *resource.ValidateConfigResponse) {
	if criteria.IsNull() || criteria.IsUnknown() {
		return
	}

	criteriaObj := criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	anyReleaseBundle := criteriaAttrs["any_release_bundle"].(types.Bool)
	registeredReleaseBundleNames := criteriaAttrs["registered_release_bundle_names"].(types.Set)
	includePatterns := criteriaAttrs["include_patterns"].(types.Set)
	excludePatterns := criteriaAttrs["exclude_patterns"].(types.Set)

	if anyReleaseBundle.IsUnknown() || registeredReleaseBundleNames.IsUnknown() || includePatterns.IsUnknown() || excludePatterns.IsUnknown() {
		return
	}

	if !anyReleaseBundle.ValueBool() && len(registeredReleaseBundleNames.Elements()) == 0 && len(includePatterns.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("any_release_bundle"),
			"Invalid Attribute Configuration",
			"registered_release_bundle_names cannot be empty when any_release_bundle is false",
		)
	}

	if anyReleaseBundle.ValueBool() && (!includePatterns.IsNull() && len(includePatterns.Elements()) > 0) {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("include_patterns"),
			"Invalid Attribute Configuration",
			"include_patterns cannot be set when any_release_bundle is true",
		)
	}

	if anyReleaseBundle.ValueBool() && (!excludePatterns.IsNull() && len(excludePatterns.Elements()) > 0) {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("exclude_patterns"),
			"Invalid Attribute Configuration",
			"exclude_patterns cannot be set when any_release_bundle is true",
		)
	}
}

func (r ReleaseBundleWebhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ReleaseBundleWebhookResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	releaseBundleValidateConfig(data.Criteria, resp)
}

func (r *ReleaseBundleWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ReleaseBundleWebhookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook WebhookAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, r.Domain, &webhook)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.WebhookResource.Create(ctx, webhook, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ReleaseBundleWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleWebhookResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook WebhookAPIModel
	found := r.WebhookResource.Read(ctx, state.Key.ValueString(), &webhook, resp)
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

func (r *ReleaseBundleWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ReleaseBundleWebhookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook WebhookAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, r.Domain, &webhook)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.WebhookResource.Update(ctx, plan.Key.ValueString(), webhook, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ReleaseBundleWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleWebhookResourceModel

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
func (r *ReleaseBundleWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.WebhookResource.ImportState(ctx, req, resp)
}

func toReleaseBundleCriteriaAPIModel(ctx context.Context, baseCriteria BaseCriteriaAPIModel, criteriaAttrs map[string]attr.Value) (criteriaAPIModel ReleaseBundleCriteriaAPIModel, diags diag.Diagnostics) {
	anyReleaseBundle := criteriaAttrs["any_release_bundle"].(types.Bool).ValueBool()

	releaseBundleNames := []string{}
	if !anyReleaseBundle {
		d := criteriaAttrs["registered_release_bundle_names"].(types.Set).ElementsAs(ctx, &releaseBundleNames, false)
		if d.HasError() {
			diags.Append(d...)
		}
	}

	criteriaAPIModel = ReleaseBundleCriteriaAPIModel{
		BaseCriteriaAPIModel:          baseCriteria,
		AnyReleaseBundle:              anyReleaseBundle,
		RegisteredReleaseBundlesNames: releaseBundleNames,
	}

	return
}

func (m ReleaseBundleWebhookResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *WebhookAPIModel) (diags diag.Diagnostics) {
	criteriaObj := m.Criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	baseCriteria, d := m.WebhookResourceModel.toBaseCriteriaAPIModel(ctx, criteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaAPIModel, d := toReleaseBundleCriteriaAPIModel(ctx, baseCriteria, criteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.WebhookResourceModel.toAPIModel(ctx, domain, criteriaAPIModel, apiModel)
	if d.HasError() {
		diags.Append(d...)
	}

	return
}

var releaseBundleCriteriaSetResourceModelAttributeTypes = lo.Assign(
	patternsCriteriaSetResourceModelAttributeTypes,
	map[string]attr.Type{
		"any_release_bundle":              types.BoolType,
		"registered_release_bundle_names": types.SetType{ElemType: types.StringType},
	},
)

var releaseBundleCriteriaSetResourceModelElementTypes = types.ObjectType{
	AttrTypes: releaseBundleCriteriaSetResourceModelAttributeTypes,
}

func fromReleaseBundleCriteriaAPIModel(ctx context.Context, criteriaAPIModel map[string]interface{}, baseCriteriaAttrs map[string]attr.Value) (criteriaSet basetypes.SetValue, diags diag.Diagnostics) {
	releaseBundleNames, d := types.SetValueFrom(ctx, types.StringType, []string{})
	if d.HasError() {
		diags.Append(d...)
	}
	if v, ok := criteriaAPIModel["registeredReleaseBundlesNames"]; ok && v != nil {
		rb, d := types.SetValueFrom(ctx, types.StringType, v)
		if d.HasError() {
			diags.Append(d...)
		}
		releaseBundleNames = rb
	}

	anyReleaseBundle := false
	if v, ok := criteriaAPIModel["anyReleaseBundle"]; ok && v != nil {
		anyReleaseBundle = v.(bool)
	}

	criteria, d := types.ObjectValue(
		releaseBundleCriteriaSetResourceModelAttributeTypes,
		lo.Assign(
			baseCriteriaAttrs,
			map[string]attr.Value{
				"any_release_bundle":              types.BoolValue(anyReleaseBundle),
				"registered_release_bundle_names": releaseBundleNames,
			},
		),
	)
	if d.HasError() {
		diags.Append(d...)
	}
	criteriaSet, d = types.SetValue(
		releaseBundleCriteriaSetResourceModelElementTypes,
		[]attr.Value{criteria},
	)
	if d.HasError() {
		diags.Append(d...)
	}

	return
}

func (m *ReleaseBundleWebhookResourceModel) fromAPIModel(ctx context.Context, apiModel WebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	criteriaAPIModel := apiModel.EventFilter.Criteria.(map[string]interface{})

	baseCriteriaAttrs, d := m.WebhookResourceModel.fromBaseCriteriaAPIModel(ctx, criteriaAPIModel)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaSet, d := fromReleaseBundleCriteriaAPIModel(ctx, criteriaAPIModel, baseCriteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.WebhookResourceModel.fromAPIModel(ctx, apiModel, stateHandlers, &criteriaSet)
	if d.HasError() {
		diags.Append(d...)
	}

	return diags
}

type ReleaseBundleCriteriaAPIModel struct {
	BaseCriteriaAPIModel
	AnyReleaseBundle              bool     `json:"anyReleaseBundle"`
	RegisteredReleaseBundlesNames []string `json:"registeredReleaseBundlesNames,omitempty"`
}
