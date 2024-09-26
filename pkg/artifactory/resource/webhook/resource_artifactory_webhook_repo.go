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
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

var _ resource.Resource = &RepoWebhookResource{}

func NewArtifactWebhookResource() resource.Resource {
	return &RepoWebhookResource{
		WebhookResource: WebhookResource{
			TypeName:    "artifactory_artifact_webhook",
			Domain:      ArtifactDomain,
			Description: "Provides an artifact webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.",
		},
	}
}

func NewArtifactPropertyWebhookResource() resource.Resource {
	return &RepoWebhookResource{
		WebhookResource: WebhookResource{
			TypeName:    "artifactory_artifact_property_webhook",
			Domain:      ArtifactPropertyDomain,
			Description: "Provides an artifact property webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.",
		},
	}
}

func NewDockerWebhookResource() resource.Resource {
	return &RepoWebhookResource{
		WebhookResource: WebhookResource{
			TypeName:    "artifactory_docker_webhook",
			Domain:      DockerDomain,
			Description: "Provides a Docker webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.",
		},
	}
}

type RepoWebhookResourceModel struct {
	WebhookResourceModel
}

type RepoWebhookResource struct {
	WebhookResource
}

func (r *RepoWebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	r.WebhookResource.Metadata(ctx, req, resp)
}

func (r *RepoWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	criteriaBlock := schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: lo.Assign(
				patternsSchemaAttributes("Simple comma separated wildcard patterns for repository artifact paths (with no leading slash).\nAnt-style path expressions are supported (*, **, ?).\nFor example: `org/apache/**`"),
				map[string]schema.Attribute{
					"any_local": schema.BoolAttribute{
						Required:    true,
						Description: "Trigger on any local repositories",
					},
					"any_remote": schema.BoolAttribute{
						Required:    true,
						Description: "Trigger on any remote repositories",
					},
					"any_federated": schema.BoolAttribute{
						Required:    true,
						Description: "Trigger on any federated repositories",
					},
					"repo_keys": schema.SetAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "Trigger on this list of repository keys",
					},
				},
			),
		},
		Validators: []validator.Set{
			setvalidator.SizeBetween(1, 1),
			setvalidator.IsRequired(),
		},
		Description: "Specifies where the webhook will be applied on which repositories.",
	}

	resp.Schema = r.schema(r.Domain, criteriaBlock)
}

func (r *RepoWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.WebhookResource.Configure(ctx, req, resp)
}

func (r RepoWebhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data RepoWebhookResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	criteriaObj := data.Criteria.Elements()[0].(types.Object)
	criteriaAttrs := criteriaObj.Attributes()

	anyLocal := criteriaAttrs["any_local"].(types.Bool).ValueBool()
	anyRemote := criteriaAttrs["any_remote"].(types.Bool).ValueBool()
	anyFederated := criteriaAttrs["any_federated"].(types.Bool).ValueBool()

	if (!anyLocal && !anyRemote && !anyFederated) && len(criteriaAttrs["repo_keys"].(types.Set).Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("criteria").AtSetValue(criteriaObj).AtName("repo_keys"),
			"Invalid Attribute Configuration",
			"repo_keys cannot be empty when any_local, any_remote, and any_federated are false",
		)
	}
}

func (r *RepoWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan RepoWebhookResourceModel

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

	plan.ID = plan.Key

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepoWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state RepoWebhookResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook WebhookAPIModel
	r.WebhookResource.Read(ctx, state.Key.ValueString(), &webhook, req, resp)

	resp.Diagnostics.Append(state.fromAPIModel(ctx, webhook, state.Handlers)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RepoWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan RepoWebhookResourceModel

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

	plan.ID = plan.Key

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepoWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state RepoWebhookResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	r.WebhookResource.Delete(ctx, state.Key.ValueString(), req, resp)

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *RepoWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.WebhookResource.ImportState(ctx, req, resp)
}

func (m RepoWebhookResourceModel) toAPIModel(ctx context.Context, domain string, apiModel *WebhookAPIModel) (diags diag.Diagnostics) {
	critieriaObj := m.Criteria.Elements()[0].(types.Object)
	critieriaAttrs := critieriaObj.Attributes()

	baseCriteria, d := m.WebhookResourceModel.toBaseCriteriaAPIModel(ctx, critieriaAttrs)
	if d.HasError() {
		diags.Append(d...)
	}

	var repoKeys []string
	d = critieriaAttrs["repo_keys"].(types.Set).ElementsAs(ctx, &repoKeys, false)
	if d.HasError() {
		diags.Append(d...)
	}

	criteriaAPIModel := RepoCriteriaAPIModel{
		BaseCriteriaAPIModel: baseCriteria,
		AnyLocal:             critieriaAttrs["any_local"].(types.Bool).ValueBool(),
		AnyRemote:            critieriaAttrs["any_remote"].(types.Bool).ValueBool(),
		AnyFederated:         critieriaAttrs["any_federated"].(types.Bool).ValueBool(),
		RepoKeys:             repoKeys,
	}

	d = m.WebhookResourceModel.toAPIModel(ctx, domain, criteriaAPIModel, apiModel)
	if d.HasError() {
		diags.Append(d...)
	}

	return
}

var repoCriteriaSetResourceModelAttributeTypes = lo.Assign(
	patternsCriteriaSetResourceModelAttributeTypes,
	map[string]attr.Type{
		"any_local":     types.BoolType,
		"any_remote":    types.BoolType,
		"any_federated": types.BoolType,
		"repo_keys":     types.SetType{ElemType: types.StringType},
	},
)

var repoCriteriaSetResourceModelElementTypes = types.ObjectType{
	AttrTypes: repoCriteriaSetResourceModelAttributeTypes,
}

func (m *RepoWebhookResourceModel) fromAPIModel(ctx context.Context, apiModel WebhookAPIModel, stateHandlers basetypes.SetValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	criteriaAPIModel := apiModel.EventFilter.Criteria.(map[string]interface{})

	baseCriteriaAttrs, d := m.WebhookResourceModel.fromBaseCriteriaAPIModel(ctx, criteriaAPIModel)

	repoKeys := types.SetNull(types.StringType)
	if v, ok := criteriaAPIModel["repoKeys"]; ok && v != nil {
		ks, d := types.SetValueFrom(ctx, types.StringType, v)
		if d.HasError() {
			diags.Append(d...)
		}

		repoKeys = ks
	}

	criteria, d := types.ObjectValue(
		repoCriteriaSetResourceModelAttributeTypes,
		lo.Assign(
			baseCriteriaAttrs,
			map[string]attr.Value{
				"any_local":     types.BoolValue(criteriaAPIModel["anyLocal"].(bool)),
				"any_remote":    types.BoolValue(criteriaAPIModel["anyRemote"].(bool)),
				"any_federated": types.BoolValue(criteriaAPIModel["anyFederated"].(bool)),
				"repo_keys":     repoKeys,
			},
		),
	)
	if d.HasError() {
		diags.Append(d...)
	}
	criteriaSet, d := types.SetValue(
		repoCriteriaSetResourceModelElementTypes,
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

type RepoCriteriaAPIModel struct {
	BaseCriteriaAPIModel
	AnyLocal     bool     `json:"anyLocal"`
	AnyRemote    bool     `json:"anyRemote"`
	AnyFederated bool     `json:"anyFederated"`
	RepoKeys     []string `json:"repoKeys"`
}
