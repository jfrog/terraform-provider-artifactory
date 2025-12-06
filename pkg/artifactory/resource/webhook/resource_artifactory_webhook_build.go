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

var _ resource.Resource = &BuildWebhookResource{}

func NewBuildWebhookResource() resource.Resource {
	return &BuildWebhookResource{
		WebhookResource: WebhookResource{
			TypeName:    fmt.Sprintf("artifactory_%s_webhook", BuildDomain),
			Domain:      BuildDomain,
			Description: "Provides a build webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.",
		},
	}
}

type BuildWebhookResourceModel struct {
	WebhookResourceModel
}

type BuildWebhookResource struct {
	WebhookResource
}

func (r *BuildWebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	r.WebhookResource.Metadata(ctx, req, resp)
}

var buildCriteriaBlock = schema.SetNestedBlock{
	NestedObject: schema.NestedBlockObject{
		Attributes: lo.Assign(
			patternsSchemaAttributes("Use Ant-style wildcard patterns to specify build names (i.e. artifact paths) in the build info repository (without a leading slash) that will be excluded from this permission target.\nAnt-style path expressions are supported (*, **, ?).\nFor example, an `apache/**` pattern will exclude the `apache` build info from the permission."),
			map[string]schema.Attribute{
				"any_build": schema.BoolAttribute{
					Required:    true,
					Description: "Trigger on any builds",
				},
				"selected_builds": schema.SetAttribute{
					ElementType: types.StringType,
					Optional: true,
					Description: "Trigger on this list of build IDs",
				},
			},
		),
	},
	Validators: []validator.Set{
		setvalidator.SizeBetween(1, 1),
		setvalidator.IsRequired(),
	},
	Description: "Specifies where the webhook will be applied on which builds.",
}

func (r *BuildWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.CreateSchema(r.Domain, &buildCriteriaBlock, handlerBlock)
}

func (r *BuildWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.WebhookResource.Configure(ctx, req, resp)
}

func buildValidateConfig(criteria basetypes.SetValue, resp *resource.ValidateConfigResponse) {
	if criteria.IsNull() || criteria.IsUnknown() {
		return
	}

	criteriaObj := criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	anyBuild := criteriaAttrs["any_build"].(types.Bool)
	selectedBuilds := criteriaAttrs["selected_builds"].(types.Set)
	includePatterns := criteriaAttrs["include_patterns"].(types.Set)
	excludePatterns := criteriaAttrs["exclude_patterns"].(types.Set)

	if anyBuild.IsUnknown() || selectedBuilds.IsUnknown() || includePatterns.IsUnknown() || excludePatterns.IsUnknown() {
		return
	}

	if !anyBuild.ValueBool() && len(selectedBuilds.Elements()) == 0 && len(includePatterns.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("any_build"),
			"Invalid Attribute Configuration",
			"selected_builds or include_patterns cannot be empty when any_build is false",
		)
	}

	if anyBuild.ValueBool() && (!includePatterns.IsNull() && len(includePatterns.Elements()) > 0) {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("include_patterns"),
			"Invalid Attribute Configuration",
			"include_patterns cannot be set when any_build is true",
		)
	}

	if anyBuild.ValueBool() && (!excludePatterns.IsNull() && len(excludePatterns.Elements()) > 0) {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("exclude_patterns"),
			"Invalid Attribute Configuration",
			"exclude_patterns cannot be set when any_build is true",
		)
	}
}

func (r BuildWebhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data BuildWebhookResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildValidateConfig(data.Criteria, resp)
}

func (r *BuildWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan BuildWebhookResourceModel

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

func (r *BuildWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state BuildWebhookResourceModel

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

func (r *BuildWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan BuildWebhookResourceModel

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

func (r *BuildWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state BuildWebhookResourceModel

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
func (r *BuildWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.WebhookResource.ImportState(ctx, req, resp)
}

func toBuildCriteriaAPIModel(ctx context.Context, baseCriteria BaseCriteriaAPIModel, criteriaAttrs map[string]attr.Value) (criteriaAPIModel BuildCriteriaAPIModel, diags diag.Diagnostics) {
	anyBuild := criteriaAttrs["any_build"].(types.Bool).ValueBool()

	selectedBuilds := []string{}
	if !anyBuild {
		d := criteriaAttrs["selected_builds"].(types.Set).ElementsAs(ctx, &selectedBuilds, false)
		if d.HasError() {
			diags.Append(d...)
		}
	}

	criteriaAPIModel = BuildCriteriaAPIModel{
		BaseCriteriaAPIModel: baseCriteria,
		AnyBuild:             anyBuild,
		SelectedBuilds:       selectedBuilds,
	}

	return
}

func (m BuildWebhookResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *WebhookAPIModel) (diags diag.Diagnostics) {
	criteriaObj := m.Criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	baseCriteria, d := m.WebhookResourceModel.toBaseCriteriaAPIModel(ctx, criteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaAPIModel, d := toBuildCriteriaAPIModel(ctx, baseCriteria, criteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.WebhookResourceModel.toAPIModel(ctx, domain, criteriaAPIModel, apiModel)
	if d.HasError() {
		diags.Append(d...)
	}

	return
}

var buildCriteriaSetResourceModelAttributeTypes = lo.Assign(
	patternsCriteriaSetResourceModelAttributeTypes,
	map[string]attr.Type{
		"any_build":       types.BoolType,
		"selected_builds": types.SetType{ElemType: types.StringType},
	},
)

var buildCriteriaSetResourceModelElementTypes = types.ObjectType{
	AttrTypes: buildCriteriaSetResourceModelAttributeTypes,
}

func fromBuildAPIModel(ctx context.Context, criteriaAPIModel map[string]interface{}, baseCriteriaAttrs map[string]attr.Value) (criteriaSet basetypes.SetValue, diags diag.Diagnostics) {
	selectedBuilds, d := types.SetValueFrom(ctx, types.StringType, []string{})
	if d.HasError() {
		diags.Append(d...)
	}
	if v, ok := criteriaAPIModel["selectedBuilds"]; ok && v != nil {
		sb, d := types.SetValueFrom(ctx, types.StringType, v)
		if d.HasError() {
			diags.Append(d...)
		}
		selectedBuilds = sb
	}

	anyBuild := false
	if v, ok := criteriaAPIModel["anyBuild"]; ok && v != nil {
		anyBuild = v.(bool)
	}

	criteria, d := types.ObjectValue(
		buildCriteriaSetResourceModelAttributeTypes,
		lo.Assign(
			baseCriteriaAttrs,
			map[string]attr.Value{
				"any_build":       types.BoolValue(anyBuild),
				"selected_builds": selectedBuilds,
			},
		),
	)
	if d.HasError() {
		diags.Append(d...)
	}
	criteriaSet, d = types.SetValue(
		buildCriteriaSetResourceModelElementTypes,
		[]attr.Value{criteria},
	)
	if d.HasError() {
		diags.Append(d...)
	}

	return criteriaSet, diags
}

func (m *BuildWebhookResourceModel) fromAPIModel(ctx context.Context, apiModel WebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	criteriaAPIModel := apiModel.EventFilter.Criteria.(map[string]interface{})

	baseCriteriaAttrs, d := m.WebhookResourceModel.fromBaseCriteriaAPIModel(ctx, criteriaAPIModel)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaSet, d := fromBuildAPIModel(ctx, criteriaAPIModel, baseCriteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.WebhookResourceModel.fromAPIModel(ctx, apiModel, stateHandlers, &criteriaSet)
	if d.HasError() {
		diags.Append(d...)
	}

	return diags
}

type BuildCriteriaAPIModel struct {
	BaseCriteriaAPIModel
	AnyBuild       bool     `json:"anyBuild"`
	SelectedBuilds []string `json:"selectedBuilds,omitempty"`
}
