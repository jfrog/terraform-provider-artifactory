package webhook

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jfrog/terraform-provider-shared/util"
)

var _ resource.Resource = &RepoCustomWebhookResource{}

func NewArtifactCustomWebhookResource() resource.Resource {
	return &RepoCustomWebhookResource{
		CustomWebhookResource: CustomWebhookResource{
			WebhookResource: WebhookResource{
				TypeName:    fmt.Sprintf("artifactory_%s_custom_webhook", ArtifactDomain),
				Domain:      ArtifactDomain,
				Description: "Provides an artifact webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.",
			},
		},
	}
}

func NewArtifactPropertyCustomWebhookResource() resource.Resource {
	return &RepoCustomWebhookResource{
		CustomWebhookResource: CustomWebhookResource{
			WebhookResource: WebhookResource{
				TypeName:    fmt.Sprintf("artifactory_%s_custom_webhook", ArtifactPropertyDomain),
				Domain:      ArtifactPropertyDomain,
				Description: "Provides an artifact property webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.",
			},
		},
	}
}

func NewDockerCustomWebhookResource() resource.Resource {
	return &RepoCustomWebhookResource{
		CustomWebhookResource: CustomWebhookResource{
			WebhookResource: WebhookResource{
				TypeName:    fmt.Sprintf("artifactory_%s_custom_webhook", DockerDomain),
				Domain:      DockerDomain,
				Description: "Provides a Docker webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.",
			},
		},
	}
}

type RepoCustomWebhookResourceModel struct {
	CustomWebhookResourceModel
}

type RepoCustomWebhookResource struct {
	CustomWebhookResource
}

func (r *RepoCustomWebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	r.WebhookResource.Metadata(ctx, req, resp)
}

func (r *RepoCustomWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.CreateSchema(r.Domain, &repoCriteriaBlock)
}

func (r *RepoCustomWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.WebhookResource.Configure(ctx, req, resp)
}

func (r RepoCustomWebhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data RepoWebhookResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Criteria.IsNull() || data.Criteria.IsUnknown() {
		return
	}

	criteriaObj := data.Criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	anyLocal := criteriaAttrs["any_local"].(types.Bool)
	anyRemote := criteriaAttrs["any_remote"].(types.Bool)
	anyFederated := criteriaAttrs["any_federated"].(types.Bool)
	repoKeys := criteriaAttrs["repo_keys"].(types.Set)

	if anyLocal.IsUnknown() || anyRemote.IsUnknown() || anyFederated.IsUnknown() || repoKeys.IsUnknown() {
		return
	}

	if (!anyLocal.ValueBool() && !anyRemote.ValueBool() && !anyFederated.ValueBool()) && len(repoKeys.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("repo_keys"),
			"Invalid Attribute Configuration",
			"repo_keys cannot be empty when any_local, any_remote, and any_federated are false",
		)
	}
}

func (r *RepoCustomWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan RepoCustomWebhookResourceModel

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

func (r *RepoCustomWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state RepoCustomWebhookResourceModel

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

func (r *RepoCustomWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan RepoCustomWebhookResourceModel

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

func (r *RepoCustomWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state RepoCustomWebhookResourceModel

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
func (r *RepoCustomWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.WebhookResource.ImportState(ctx, req, resp)
}

func (m RepoCustomWebhookResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *CustomWebhookAPIModel) (diags diag.Diagnostics) {
	criteriaObj := m.Criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	baseCriteria, d := m.CustomWebhookResourceModel.toBaseCriteriaAPIModel(ctx, criteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaAPIModel, d := toRepoCriteriaAPIModel(ctx, baseCriteria, criteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.CustomWebhookResourceModel.toAPIModel(ctx, domain, criteriaAPIModel, apiModel)
	if d.HasError() {
		diags.Append(d...)
	}

	return
}

func (m *RepoCustomWebhookResourceModel) fromAPIModel(ctx context.Context, apiModel CustomWebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	criteriaAPIModel := apiModel.EventFilter.Criteria.(map[string]interface{})

	baseCriteriaAttrs, d := m.CustomWebhookResourceModel.fromBaseCriteriaAPIModel(ctx, criteriaAPIModel)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaSet, d := fromRepoCriteriaAPIMode(ctx, criteriaAPIModel, baseCriteriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	d = m.CustomWebhookResourceModel.fromAPIModel(ctx, apiModel, stateHandlers, &criteriaSet)
	if d.HasError() {
		diags.Append(d...)
	}

	return diags
}
