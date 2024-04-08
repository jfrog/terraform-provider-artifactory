package security

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
)

const UserLockPolicyEndpoint = "artifactory/api/security/userLockPolicy"

func NewUserLockPolicyResource() resource.Resource {
	return &UserLockPolicyResource{}
}

type UserLockPolicyResource struct {
	ProviderData util.ProvderMetadata
}

type UserLockPolicyResourceModel struct {
	Name          types.String `tfsdk:"name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	LoginAttempts types.Int64  `tfsdk:"login_attempts"`
}

type UserLockPolicyAPIModel struct {
	Enabled       bool  `json:"enabled"`
	LoginAttempts int64 `json:"loginAttempts"`
}

func (r *UserLockPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_lock_policy"
}

func (r *UserLockPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "Name of the resource. Only used for importing.",
			},
			"enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Enable User Lock Policy. Lock user after exceeding max failed login attempts.",
			},
			"login_attempts": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Max failed login attempts.",
			},
		},
		MarkdownDescription: "Provides an Artifactory User Lock Policy resource.",
	}
}

func (r *UserLockPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProvderMetadata)
}

func (r *UserLockPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// go util.SendUsageResourceCreate(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var plan UserLockPolicyResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	policy := &UserLockPolicyAPIModel{
		Enabled:       plan.Enabled.ValueBool(),
		LoginAttempts: plan.LoginAttempts.ValueInt64(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(policy).
		Put(UserLockPolicyEndpoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserLockPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// go util.SendUsageResourceRead(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var state *UserLockPolicyResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var policy UserLockPolicyAPIModel

	response, err := r.ProviderData.Client.R().
		SetResult(&policy).
		Get(UserLockPolicyEndpoint)

	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	state.Enabled = types.BoolValue(policy.Enabled)
	state.LoginAttempts = types.Int64Value(policy.LoginAttempts)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *UserLockPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// go util.SendUsageResourceCreate(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var plan UserLockPolicyResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	policy := &UserLockPolicyAPIModel{
		Enabled:       plan.Enabled.ValueBool(),
		LoginAttempts: plan.LoginAttempts.ValueInt64(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(policy).
		Put(UserLockPolicyEndpoint)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserLockPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// go util.SendUsageResourceDelete(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var state UserLockPolicyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	resp.Diagnostics.AddWarning(
		"User Lock Policy cannot be deleted",
		"Artifactory does not support deletion of the user lock policy. Provider will disable the policy and set login attempts to 0 instead.",
	)

	// Convert from Terraform data model into API data model
	policy := &UserLockPolicyAPIModel{
		Enabled:       false,
		LoginAttempts: 0,
	}

	response, err := r.ProviderData.Client.R().
		SetBody(policy).
		Put(UserLockPolicyEndpoint)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *UserLockPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
