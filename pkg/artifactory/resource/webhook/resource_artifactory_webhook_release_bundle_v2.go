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

var _ resource.Resource = &ReleaseBundleV2WebhookResource{}

func NewReleaseBundleV2WebhookResource() resource.Resource {
	return &ReleaseBundleV2WebhookResource{
		WebhookResource: WebhookResource{
			TypeName:    fmt.Sprintf("artifactory_%s_webhook", ReleaseBundleV2Domain),
			Domain:      ReleaseBundleV2Domain,
			Description: "Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.:",
		},
	}
}

type ReleaseBundleV2WebhookResourceModel struct {
	WebhookResourceModel
}

type ReleaseBundleV2WebhookResource struct {
	WebhookResource
}

func (r *ReleaseBundleV2WebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	r.WebhookResource.Metadata(ctx, req, resp)
}

var releaseBundleV2CriteriaBlock = schema.SetNestedBlock{
	NestedObject: schema.NestedBlockObject{
		Attributes: lo.Assign(
			patternsSchemaAttributes("Simple wildcard patterns for Release Bundle names.\nAnt-style path expressions are supported (*, **, ?).\nFor example: `product_*`"),
			map[string]schema.Attribute{
				"any_release_bundle": schema.BoolAttribute{
					Required:    true,
					Description: "Includes all existing release bundles and any future release bundles.",
				},
				"selected_release_bundles": schema.SetAttribute{
					ElementType: types.StringType,
					Required:    true,
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

func (r *ReleaseBundleV2WebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.CreateSchema(r.Domain, &releaseBundleV2CriteriaBlock, handlerBlock)
}

func (r *ReleaseBundleV2WebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.WebhookResource.Configure(ctx, req, resp)
}

func releaseBundleV2ValidatConfig(criteria basetypes.SetValue, resp *resource.ValidateConfigResponse) {
	if criteria.IsNull() || criteria.IsUnknown() {
		return
	}

	criteriaObj := criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	anyReleaseBundle := criteriaAttrs["any_release_bundle"].(types.Bool)
	selectedReleaseBundles := criteriaAttrs["selected_release_bundles"].(types.Set)

	if anyReleaseBundle.IsUnknown() || selectedReleaseBundles.IsUnknown() {
		return
	}

	if !anyReleaseBundle.ValueBool() && len(selectedReleaseBundles.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("any_release_bundle"),
			"Invalid Attribute Configuration",
			"selected_release_bundles cannot be empty when any_release_bundle is false",
		)
	}
}

func (r ReleaseBundleV2WebhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ReleaseBundleV2WebhookResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	releaseBundleV2ValidatConfig(data.Criteria, resp)
}

func (r *ReleaseBundleV2WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ReleaseBundleV2WebhookResourceModel

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

func (r *ReleaseBundleV2WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleV2WebhookResourceModel

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

func (r *ReleaseBundleV2WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ReleaseBundleV2WebhookResourceModel

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

func (r *ReleaseBundleV2WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleV2WebhookResourceModel

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
func (r *ReleaseBundleV2WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.WebhookResource.ImportState(ctx, req, resp)
}

func toReleaseBundleV2APIModel(ctx context.Context, baseCriteria BaseCriteriaAPIModel, criteriaAttrs map[string]attr.Value) (criteriaAPIModel ReleaseBundleV2CriteriaAPIModel, diags diag.Diagnostics) {
	anyReleaseBundle := criteriaAttrs["any_release_bundle"].(types.Bool).ValueBool()

	var releaseBundleNames []string
	if !anyReleaseBundle {
		d := criteriaAttrs["selected_release_bundles"].(types.Set).ElementsAs(ctx, &releaseBundleNames, false)
		if d.HasError() {
			diags.Append(d...)
		}
	}

	return ReleaseBundleV2CriteriaAPIModel{
		BaseCriteriaAPIModel:   baseCriteria,
		AnyReleaseBundle:       anyReleaseBundle,
		SelectedReleaseBundles: releaseBundleNames,
	}, diags
}

func (m ReleaseBundleV2WebhookResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *WebhookAPIModel) (diags diag.Diagnostics) {
	criteriaObj := m.Criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	baseCriteria, d := m.WebhookResourceModel.toBaseCriteriaAPIModel(ctx, criteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaAPIModel, d := toReleaseBundleV2APIModel(ctx, baseCriteria, criteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.WebhookResourceModel.toAPIModel(ctx, domain, criteriaAPIModel, apiModel)
	if d.HasError() {
		diags.Append(d...)
	}

	return
}

var releaseBundleV2CriteriaSetResourceModelAttributeTypes = lo.Assign(
	patternsCriteriaSetResourceModelAttributeTypes,
	map[string]attr.Type{
		"any_release_bundle":       types.BoolType,
		"selected_release_bundles": types.SetType{ElemType: types.StringType},
	},
)

var releaseBundleV2CriteriaSetResourceModelElementTypes = types.ObjectType{
	AttrTypes: releaseBundleV2CriteriaSetResourceModelAttributeTypes,
}

func fromReleaseBundleV2APIModel(ctx context.Context, criteriaAPIModel map[string]interface{}, baseCriteriaAttrs map[string]attr.Value) (criteriaSet basetypes.SetValue, diags diag.Diagnostics) {
	releaseBundleNames, d := types.SetValueFrom(ctx, types.StringType, []string{})
	if d.HasError() {
		diags.Append(d...)
	}
	if v, ok := criteriaAPIModel["selectedReleaseBundles"]; ok && v != nil {
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
		releaseBundleV2CriteriaSetResourceModelAttributeTypes,
		lo.Assign(
			baseCriteriaAttrs,
			map[string]attr.Value{
				"any_release_bundle":       types.BoolValue(anyReleaseBundle),
				"selected_release_bundles": releaseBundleNames,
			},
		),
	)
	if d.HasError() {
		diags.Append(d...)
	}
	criteriaSet, d = types.SetValue(
		releaseBundleV2CriteriaSetResourceModelElementTypes,
		[]attr.Value{criteria},
	)
	if d.HasError() {
		diags.Append(d...)
	}

	return
}

func (m *ReleaseBundleV2WebhookResourceModel) fromAPIModel(ctx context.Context, apiModel WebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	criteriaAPIModel := apiModel.EventFilter.Criteria.(map[string]interface{})

	baseCriteriaAttrs, d := m.WebhookResourceModel.fromBaseCriteriaAPIModel(ctx, criteriaAPIModel)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaSet, d := fromReleaseBundleV2APIModel(ctx, criteriaAPIModel, baseCriteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.WebhookResourceModel.fromAPIModel(ctx, apiModel, stateHandlers, &criteriaSet)
	if d.HasError() {
		diags.Append(d...)
	}

	return diags
}

type ReleaseBundleV2CriteriaAPIModel struct {
	BaseCriteriaAPIModel
	AnyReleaseBundle       bool     `json:"anyReleaseBundle"`
	SelectedReleaseBundles []string `json:"selectedReleaseBundles"`
}
