package webhook

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var _ resource.Resource = &BuildWebhookResource{}

func NewBuildWebhookResource() resource.Resource {
	return &BuildWebhookResource{
		WebhookResource: WebhookResource{
			TypeName:    "artifactory_build_webhook",
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

func (r *BuildWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	criteriaBlock := schema.SetNestedBlock{
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
						Required:    true,
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

	resp.Schema = r.schema(r.Domain, criteriaBlock)
}

func (r *BuildWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.WebhookResource.Configure(ctx, req, resp)
}

func (r BuildWebhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data BuildWebhookResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	criteriaObj := data.Criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	anyBuild := criteriaAttrs["any_build"].(types.Bool).ValueBool()

	if !anyBuild && len(criteriaAttrs["selected_builds"].(types.Set).Elements()) == 0 && len(criteriaAttrs["include_patterns"].(types.Set).Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("any_build"),
			"Invalid Attribute Configuration",
			"selected_builds or include_patterns cannot be empty when any_build is false",
		)
	}
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

	r.WebhookResource.Create(ctx, webhook, req, resp)
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
	found := r.WebhookResource.Read(ctx, state.Key.ValueString(), &webhook, req, resp)
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

	r.WebhookResource.Update(ctx, plan.Key.ValueString(), webhook, req, resp)
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

	r.WebhookResource.Delete(ctx, state.Key.ValueString(), req, resp)
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

func (m BuildWebhookResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *WebhookAPIModel) (diags diag.Diagnostics) {
	critieriaObj := m.Criteria.Elements()[0].(types.Object)
	critieriaAttrs := critieriaObj.Attributes()

	baseCriteria, d := m.WebhookResourceModel.toBaseCriteriaAPIModel(ctx, critieriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	var selectedBuilds []string
	d = critieriaAttrs["selected_builds"].(types.Set).ElementsAs(ctx, &selectedBuilds, false)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaAPIModel := BuildCriteriaAPIModel{
		BaseCriteriaAPIModel: baseCriteria,
		AnyBuild:             critieriaAttrs["any_build"].(types.Bool).ValueBool(),
		SelectedBuilds:       selectedBuilds,
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

func (m *BuildWebhookResourceModel) fromAPIModel(ctx context.Context, apiModel WebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	criteriaAPIModel := apiModel.EventFilter.Criteria.(map[string]interface{})

	baseCriteriaAttrs, d := m.WebhookResourceModel.fromBaseCriteriaAPIModel(ctx, criteriaAPIModel)

	selectedBuilds := types.SetNull(types.StringType)
	if v, ok := criteriaAPIModel["selectedBuilds"]; ok && v != nil {
		sb, d := types.SetValueFrom(ctx, types.StringType, v)
		if d.HasError() {
			diags.Append(d...)
		}

		selectedBuilds = sb
	}

	criteria, d := types.ObjectValue(
		buildCriteriaSetResourceModelAttributeTypes,
		lo.Assign(
			baseCriteriaAttrs,
			map[string]attr.Value{
				"any_build":       types.BoolValue(criteriaAPIModel["anyBuild"].(bool)),
				"selected_builds": selectedBuilds,
			},
		),
	)
	if d.HasError() {
		diags.Append(d...)
	}
	criteriaSet, d := types.SetValue(
		buildCriteriaSetResourceModelElementTypes,
		[]attr.Value{criteria},
	)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.WebhookResourceModel.fromAPIModel(ctx, apiModel, stateHandlers, criteriaSet)
	if d.HasError() {
		diags.Append(d...)
	}

	return diags
}

type BuildCriteriaAPIModel struct {
	BaseCriteriaAPIModel
	AnyBuild       bool     `json:"anyBuild"`
	SelectedBuilds []string `json:"selectedBuilds"`
}

var buildWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*sdkv2_schema.Schema {
	return utilsdk.MergeMaps(getBaseSchemaByVersion(webhookType, version, isCustom), map[string]*sdkv2_schema.Schema{
		"criteria": {
			Type:     sdkv2_schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &sdkv2_schema.Resource{
				Schema: utilsdk.MergeMaps(baseCriteriaSchema, map[string]*sdkv2_schema.Schema{
					"any_build": {
						Type:        sdkv2_schema.TypeBool,
						Required:    true,
						Description: "Trigger on any builds",
					},
					"selected_builds": {
						Type:        sdkv2_schema.TypeSet,
						Required:    true,
						Elem:        &sdkv2_schema.Schema{Type: sdkv2_schema.TypeString},
						Description: "Trigger on this list of build IDs",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied on which builds.",
		},
	})
}

// var packBuildCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
// 	return map[string]interface{}{
// 		"any_build":       artifactoryCriteria["anyBuild"].(bool),
// 		"selected_builds": sdkv2_schema.NewSet(sdkv2_schema.HashString, artifactoryCriteria["selectedBuilds"].([]interface{})),
// 	}
// }
//
// var unpackBuildCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseCriteriaAPIModel) interface{} {
// 	return BuildCriteriaAPIModel{
// 		AnyBuild:             terraformCriteria["any_build"].(bool),
// 		SelectedBuilds:       utilsdk.CastToStringArr(terraformCriteria["selected_builds"].(*sdkv2_schema.Set).List()),
// 		BaseCriteriaAPIModel: baseCriteria,
// 	}
// }
//
// var buildCriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
// 	anyBuild := criteria["any_build"].(bool)
// 	selectedBuilds := criteria["selected_builds"].(*sdkv2_schema.Set).List()
// 	includePatterns := criteria["include_patterns"].(*sdkv2_schema.Set).List()
//
// 	if !anyBuild && (len(selectedBuilds) == 0 && len(includePatterns) == 0) {
// 		return fmt.Errorf("selected_builds or include_patterns cannot be empty when any_build is false")
// 	}
//
// 	return nil
// }
