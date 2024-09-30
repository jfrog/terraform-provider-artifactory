package webhook

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

var _ resource.Resource = &ReleaseBundleV2WebhookResource{}

func NewReleaseBundleV2PromotionWebhookResource() resource.Resource {
	return &ReleaseBundleV2PromotionWebhookResource{
		WebhookResource: WebhookResource{
			TypeName:    fmt.Sprintf("artifactory_%s_webhook", ReleaseBundleV2PromotionDomain),
			Domain:      ReleaseBundleV2PromotionDomain,
			Description: "Provides an Artifactory webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.:",
		},
	}
}

type ReleaseBundleV2PromotionWebhookResourceModel struct {
	WebhookResourceModel
}

type ReleaseBundleV2PromotionWebhookResource struct {
	WebhookResource
}

func (r *ReleaseBundleV2PromotionWebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	r.WebhookResource.Metadata(ctx, req, resp)
}

func (r *ReleaseBundleV2PromotionWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	criteriaBlock := schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: lo.Assign(
				patternsSchemaAttributes(""),
				map[string]schema.Attribute{
					"selected_environments": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "Trigger on this list of environments",
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

	resp.Schema = r.CreateSchema(r.Domain, &criteriaBlock, handlerBlock)
}

func (r *ReleaseBundleV2PromotionWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.WebhookResource.Configure(ctx, req, resp)
}

func (r *ReleaseBundleV2PromotionWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ReleaseBundleV2PromotionWebhookResourceModel

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

func (r *ReleaseBundleV2PromotionWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleV2PromotionWebhookResourceModel

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

func (r *ReleaseBundleV2PromotionWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ReleaseBundleV2PromotionWebhookResourceModel

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

func (r *ReleaseBundleV2PromotionWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleV2PromotionWebhookResourceModel

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
func (r *ReleaseBundleV2PromotionWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.WebhookResource.ImportState(ctx, req, resp)
}

func (m ReleaseBundleV2PromotionWebhookResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *WebhookAPIModel) (diags diag.Diagnostics) {
	critieriaObj := m.Criteria.Elements()[0].(types.Object)
	critieriaAttrs := critieriaObj.Attributes()

	var environments []string
	d := critieriaAttrs["selected_environments"].(types.Set).ElementsAs(ctx, &environments, false)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaAPIModel := ReleaseBundleV2PromotionCriteriaAPIModel{
		SelectedEnvironments: environments,
	}

	d = m.WebhookResourceModel.toAPIModel(ctx, domain, criteriaAPIModel, apiModel)
	if d.HasError() {
		diags.Append(d...)
	}

	return
}

var releaseBundleV2PromotionCriteriaSetResourceModelAttributeTypes = lo.Assign(
	patternsCriteriaSetResourceModelAttributeTypes,
	map[string]attr.Type{
		"selected_environments": types.SetType{ElemType: types.StringType},
	},
)

var releaseBundleV2PromotionCriteriaSetResourceModelElementTypes = types.ObjectType{
	AttrTypes: releaseBundleV2PromotionCriteriaSetResourceModelAttributeTypes,
}

func (m *ReleaseBundleV2PromotionWebhookResourceModel) fromAPIModel(ctx context.Context, apiModel WebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	criteriaAPIModel := apiModel.EventFilter.Criteria.(map[string]interface{})

	baseCriteriaAttrs, d := m.WebhookResourceModel.fromBaseCriteriaAPIModel(ctx, criteriaAPIModel)

	releaseBundleNames := types.SetNull(types.StringType)
	if v, ok := criteriaAPIModel["selectedEnvironments"]; ok && v != nil {
		rb, d := types.SetValueFrom(ctx, types.StringType, v)
		if d.HasError() {
			diags.Append(d...)
		}

		releaseBundleNames = rb
	}

	criteria, d := types.ObjectValue(
		releaseBundleV2PromotionCriteriaSetResourceModelAttributeTypes,
		lo.Assign(
			baseCriteriaAttrs,
			map[string]attr.Value{
				"selected_environments": releaseBundleNames,
			},
		),
	)
	if d.HasError() {
		diags.Append(d...)
	}
	criteriaSet, d := types.SetValue(
		releaseBundleV2PromotionCriteriaSetResourceModelElementTypes,
		[]attr.Value{criteria},
	)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.WebhookResourceModel.fromAPIModel(ctx, apiModel, stateHandlers, &criteriaSet)
	if d.HasError() {
		diags.Append(d...)
	}

	return diags
}

type ReleaseBundleV2PromotionCriteriaAPIModel struct {
	SelectedEnvironments []string `json:"selectedEnvironments"`
}
