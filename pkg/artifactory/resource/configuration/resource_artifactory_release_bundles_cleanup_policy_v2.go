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
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

const (
	resourceBundleItemTypeName = "releaseBundle"
)

func NewResourceBundleCleanupPolicyV2Resource() resource.Resource {
	return &ResourceBundleCleanupPolicyV2Resource{
		JFrogResource: util.JFrogResource{
			TypeName:                "artifactory_release_bundle_v2_cleanup_policy",
			ValidArtifactoryVersion: "7.104.2",
			DocumentEndpoint:        "artifactory/api/cleanup/bundles/policies/{policyKey}",
		},
		EnablementEndpoint: "artifactory/api/cleanup/bundles/policies/{policyKey}/enablement",
	}
}

type ResourceBundleCleanupPolicyV2Resource struct {
	util.JFrogResource
	EnablementEndpoint string
}

type ResourceBundleCleanupPolicyV2ResourceModel struct {
	Key               types.String `tfsdk:"key"`
	Description       types.String `tfsdk:"description"`
	CronExpression    types.String `tfsdk:"cron_expression"`
	ItemType          types.String `tfsdk:"item_type"`
	DurationInMinutes types.Int64  `tfsdk:"duration_in_minutes"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	SearchCriteria    types.Object `tfsdk:"search_criteria"`
}

var ResourceBundleAttributeTypes map[string]attr.Type = map[string]attr.Type{
	"name":        types.StringType,
	"project_key": types.StringType,
}

var SearchCriteriaResourceBundleAttributeTypes types.ObjectType = types.ObjectType{
	AttrTypes: ResourceBundleAttributeTypes,
}

func (r ResourceBundleCleanupPolicyV2ResourceModel) toAPIModel(ctx context.Context, apiModel *ResourceBundleCleanupPolicyV2APIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	searchCriteriaAttrs := r.SearchCriteria.Attributes()
	releaseBundles := lo.Map(
		searchCriteriaAttrs["release_bundles"].(types.Set).Elements(),
		func(elem attr.Value, _ int) ResourceBundleCleanupPolicyV2ReleaseBundlesAPIModel {
			attrs := elem.(types.Object).Attributes()

			return ResourceBundleCleanupPolicyV2ReleaseBundlesAPIModel{
				Name:       attrs["name"].(types.String).ValueString(),
				ProjectKey: attrs["project_key"].(types.String).ValueString(),
			}
		},
	)
	searchCriteria := ResourceBundleCleanupPolicyV2SearchCriteriaAPIModel{
		IncludeAllProjects:    searchCriteriaAttrs["include_all_projects"].(types.Bool).ValueBoolPointer(),
		CreatedBeforeInMonths: searchCriteriaAttrs["created_before_in_months"].(types.Int64).ValueInt64Pointer(),
		ReleaseBundles:        &releaseBundles,
	}

	diags.Append(searchCriteriaAttrs["exclude_promoted_environments"].(types.Set).ElementsAs(ctx, &searchCriteria.ExcludePromotedEnvironments, false)...)
	diags.Append(searchCriteriaAttrs["included_projects"].(types.Set).ElementsAs(ctx, &searchCriteria.IncludedProjects, false)...)

	*apiModel = ResourceBundleCleanupPolicyV2APIModel{
		Key:               r.Key.ValueString(),
		Description:       r.Description.ValueString(),
		CronExpression:    r.CronExpression.ValueString(),
		DurationInMinutes: r.DurationInMinutes.ValueInt64(),
		ItemType:          resourceBundleItemTypeName,
		SearchCriteria:    searchCriteria,
	}

	return diags
}

func (r *ResourceBundleCleanupPolicyV2ResourceModel) fromAPIModel(ctx context.Context, apiModel ResourceBundleCleanupPolicyV2APIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	r.Key = types.StringValue(apiModel.Key)
	r.Description = types.StringValue(apiModel.Description)
	r.CronExpression = types.StringValue(apiModel.CronExpression)
	r.DurationInMinutes = types.Int64Value(apiModel.DurationInMinutes)
	r.ItemType = types.StringValue(apiModel.ItemType)
	r.Enabled = types.BoolValue(apiModel.Enabled)

	excludePromotedEnvironments, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.ExcludePromotedEnvironments)
	if ds.HasError() {
		diags.Append(ds...)
	}

	includedProjects, ds := types.SetValueFrom(ctx, types.StringType, apiModel.SearchCriteria.IncludedProjects)
	if ds.HasError() {
		diags.Append(ds...)
	}

	releaseBundles := lo.Map(
		*apiModel.SearchCriteria.ReleaseBundles,
		func(property ResourceBundleCleanupPolicyV2ReleaseBundlesAPIModel, _ int) attr.Value {
			releaseBundle, ds := types.ObjectValue(
				ResourceBundleAttributeTypes,
				map[string]attr.Value{
					"name":        types.StringValue(property.Name),
					"project_key": types.StringValue(property.ProjectKey),
				},
			)

			if ds != nil {
				diags.Append(ds...)
			}

			return releaseBundle
		},
	)
	releaseBundlesSet, ds := types.SetValueFrom(
		ctx,
		SearchCriteriaResourceBundleAttributeTypes,
		releaseBundles,
	)

	searchCriteria, ds := types.ObjectValue(
		map[string]attr.Type{
			"exclude_promoted_environments": types.SetType{ElemType: types.StringType},
			"include_all_projects":          types.BoolType,
			"included_projects":             types.SetType{ElemType: types.StringType},
			"created_before_in_months":      types.Int64Type,
			"release_bundles":               types.SetType{ElemType: SearchCriteriaResourceBundleAttributeTypes},
		},
		map[string]attr.Value{
			"exclude_promoted_environments": excludePromotedEnvironments,
			"include_all_projects":          types.BoolPointerValue(apiModel.SearchCriteria.IncludeAllProjects),
			"included_projects":             includedProjects,
			"created_before_in_months":      types.Int64PointerValue(apiModel.SearchCriteria.CreatedBeforeInMonths),
			"release_bundles":               releaseBundlesSet,
		},
	)

	if ds.HasError() {
		diags.Append(ds...)
	}

	r.SearchCriteria = searchCriteria

	return diags
}

type ResourceBundleCleanupPolicyV2APIModel struct {
	Key               string                                              `json:"key"`
	Description       string                                              `json:"description,omitempty"`
	CronExpression    string                                              `json:"cronExp,omitempty"`
	ItemType          string                                              `json:"itemType"`
	DurationInMinutes int64                                               `json:"durationInMinutes"`
	Enabled           bool                                                `json:"enabled,omitempty"`
	SearchCriteria    ResourceBundleCleanupPolicyV2SearchCriteriaAPIModel `json:"searchCriteria,omitempty"`
}

type ResourceBundleCleanupPolicyV2SearchCriteriaAPIModel struct {
	ReleaseBundles              *[]ResourceBundleCleanupPolicyV2ReleaseBundlesAPIModel `json:"releaseBundles,omitempty"`
	ExcludePromotedEnvironments *[]string                                              `json:"excludePromotedEnvironments,omitempty"`
	IncludeAllProjects          *bool                                                  `json:"includeAllProjects,omitempty"`
	IncludedProjects            *[]string                                              `json:"includedProjects,omitempty"`
	CreatedBeforeInMonths       *int64                                                 `json:"createdBeforeInMonths,omitempty"`
}

type ResourceBundleCleanupPolicyV2ReleaseBundlesAPIModel struct {
	Name       string `json:"name,omitempty"`
	ProjectKey string `json:"projectKey"`
}

type ResourceBundleCleanupPolicyV2EnablementAPIModel struct {
	Enabled bool `json:"enabled"`
}

func (r *ResourceBundleCleanupPolicyV2Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`), "only letters, numbers, underscore and hyphen are allowed"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "An ID that is used to identify the release bundle cleanup policy. A minimum of three characters is required and can include letters, numbers, underscore and hyphen.",
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"cron_expression": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validatorfw_string.IsCron(),
				},
				MarkdownDescription: "The cron expression determines when the policy is run. This parameter is not mandatory, however if left empty the policy will not run automatically and can only be triggered manually.",
			},
			"duration_in_minutes": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The maximum duration (in minutes) for policy execution, after which the policy will stop running even if not completed. While setting a maximum run duration for a policy is useful for adhering to a strict archive V2 schedule, it can cause the policy to stop before completion.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enables or disabled the release bundle cleanup policy. This allows the user to run the policy manually. If a policy has a valid cron expression, then it will be scheduled for execution based on it. If a policy is disabled, its future executions will be unscheduled. Defaults to `true`",
			},
			"item_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(resourceBundleItemTypeName),
				MarkdownDescription: "Needs to be set to releaseBundle.",
			},
			"search_criteria": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"release_bundles": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
									MarkdownDescription: "The name of the release bundle. Set '**' to include all bundles.",
								},
								"project_key": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "The project identifier associated with the release bundle. This key is obtained from the Project Settings screen. Leave the field blank to apply at a global level.",
								},
							},
						},
						Optional:            true,
						MarkdownDescription: "Specify the release bundles to include in the cleanup policy. The policy will only clean up the release bundles that match the specified criteria.",
					},
					"exclude_promoted_environments": schema.SetAttribute{
						Required:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "A list of environments to exclude from the cleanup process. To exclude all, set to **",
					},
					"include_all_projects": schema.BoolAttribute{
						Optional:    true,
						Description: "Set this value to `true` if you want the policy to run on all Artifactory projects. The default value is `false`.\n\n~>This attribute is relevant only on the global level, for Platform Admins.",
					},
					"included_projects": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						MarkdownDescription: "List of projects on which you want this policy to run. To include repositories that are not assigned to any project, enter the project key `default`.\n\n" +
							"~>This setting is relevant only on the global level, for Platform Admins.",
					},
					"created_before_in_months": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(24),
						MarkdownDescription: "Specifies the time frame for filtering based on item creation date (for example, 24 months).",
					},
				},
				Required: true,
			},
		},
		Description: "Provides an Artifactory Release Bundles Cleanup Policy resource. The following APIs are used to configure and maintain JFrog cleanup policies for Release Bundles V2. " +
			"See [Cleanup Policies Release Bundles](https://jfrog.com/help/r/jfrog-rest-apis/cleanup-policies-release-bundles-v2-apis) for more details.\n\n",
	}
}

func (r *ResourceBundleCleanupPolicyV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ResourceBundleCleanupPolicyV2ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy ResourceBundleCleanupPolicyV2APIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &policy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jfrogErrors util.JFrogErrors
	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", plan.Key.ValueString()).
		SetBody(policy).
		SetError(&jfrogErrors).
		Post(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, jfrogErrors.String())
		return
	}

	// if Enabled has changed then call enablement API to toggle the value
	if plan.Enabled.ValueBool() {
		policyEnablement := ResourceBundleCleanupPolicyV2EnablementAPIModel{
			Enabled: true,
		}

		enablementResp, enablementErr := r.ProviderData.Client.R().
			SetPathParam("policyKey", plan.Key.ValueString()).
			SetBody(policyEnablement).
			SetError(&jfrogErrors).
			Post(r.EnablementEndpoint)

		if enablementErr != nil {
			utilfw.UnableToCreateResourceError(resp, enablementErr.Error())
			return
		}

		if enablementResp.IsError() {
			utilfw.UnableToCreateResourceError(resp, jfrogErrors.String())
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ResourceBundleCleanupPolicyV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ResourceBundleCleanupPolicyV2ResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var policy ResourceBundleCleanupPolicyV2APIModel
	var jfrogErrors util.JFrogErrors

	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", state.Key.ValueString()).
		SetResult(&policy).
		SetError(&jfrogErrors).
		Get(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, jfrogErrors.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.fromAPIModel(ctx, policy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceBundleCleanupPolicyV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ResourceBundleCleanupPolicyV2ResourceModel
	var state ResourceBundleCleanupPolicyV2ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy ResourceBundleCleanupPolicyV2APIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &policy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// policy.Enabled can't be changed using update API so set the field to
	// the current state's value
	policy.Enabled = state.Enabled.ValueBool()

	var jfrogErrors util.JFrogErrors
	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", plan.Key.ValueString()).
		SetBody(policy).
		SetError(&jfrogErrors).
		Put(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, jfrogErrors.String())
		return
	}

	// if Enabled has changed then call enablement API to toggle the value
	enabledChanged := state.Enabled.ValueBool() != plan.Enabled.ValueBool()
	if enabledChanged {
		policyEnablement := ResourceBundleCleanupPolicyV2EnablementAPIModel{}
		if state.Enabled.ValueBool() && !plan.Enabled.ValueBool() { // if Enabled goes from true to false
			policyEnablement.Enabled = false
		} else if !state.Enabled.ValueBool() && plan.Enabled.ValueBool() { // if Enabled goes from false to true
			policyEnablement.Enabled = true
		}

		enablementResp, enablementErr := r.ProviderData.Client.R().
			SetPathParam("policyKey", plan.Key.ValueString()).
			SetBody(policyEnablement).
			SetError(&jfrogErrors).
			Post(r.EnablementEndpoint)

		if enablementErr != nil {
			utilfw.UnableToUpdateResourceError(resp, enablementErr.Error())
			return
		}

		if enablementResp.IsError() {
			utilfw.UnableToUpdateResourceError(resp, jfrogErrors.String())
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ResourceBundleCleanupPolicyV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ResourceBundleCleanupPolicyV2ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var jfrogErrors util.JFrogErrors

	response, err := r.ProviderData.Client.R().
		SetPathParam("policyKey", state.Key.ValueString()).
		SetError(&jfrogErrors).
		Delete(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToDeleteResourceError(resp, jfrogErrors.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ResourceBundleCleanupPolicyV2Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)

	if len(parts) > 0 && parts[0] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), parts[0])...)
	}
}
