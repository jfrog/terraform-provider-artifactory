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

const PasswordExpirationPolicyEndpoint = "artifactory/api/security/configuration/passwordExpirationPolicy"

func NewPasswordExpirationPolicyResource() resource.Resource {
	return &PasswordExpirationPolicyResource{
		TypeName: "artifactory_password_expiration_policy",
	}
}

type PasswordExpirationPolicyResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type PasswordExpirationPolicyResourceModel struct {
	Name           types.String `tfsdk:"name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	PasswordMaxAge types.Int64  `tfsdk:"password_max_age"`
	NotifyByEmail  types.Bool   `tfsdk:"notify_by_email"`
}

type PasswordExpirationPolicyAPIModel struct {
	Enabled        bool  `json:"enabled"`
	PasswordMaxAge int64 `json:"passwordMaxAge"`
	NotifyByEmail  bool  `json:"notifyByEmail"`
}

func (r *PasswordExpirationPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *PasswordExpirationPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				MarkdownDescription: "Enable Password Expiration Policy. This only applies to internal user passwords.",
			},
			"password_max_age": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Password expires every N days. The time interval in which users will be obligated to change their password.",
			},
			"notify_by_email": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Send mail notification before password expiration. Users will receive an email notification a few days before password will expire. Mail server must be enabled and configured correctly.",
			}},
		MarkdownDescription: "Provides an Artifactory Password Expiration Policy resource. See [JFrog documentation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/password-expiration-policy) for more details.",
	}
}

func (r *PasswordExpirationPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *PasswordExpirationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan PasswordExpirationPolicyResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	policy := &PasswordExpirationPolicyAPIModel{
		Enabled:        plan.Enabled.ValueBool(),
		PasswordMaxAge: plan.PasswordMaxAge.ValueInt64(),
		NotifyByEmail:  plan.NotifyByEmail.ValueBool(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(policy).
		Put(PasswordExpirationPolicyEndpoint)

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

func (r *PasswordExpirationPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state *PasswordExpirationPolicyResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var policy PasswordExpirationPolicyAPIModel

	response, err := r.ProviderData.Client.R().
		SetResult(&policy).
		Get(PasswordExpirationPolicyEndpoint)

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
	state.PasswordMaxAge = types.Int64Value(policy.PasswordMaxAge)
	state.NotifyByEmail = types.BoolValue(policy.NotifyByEmail)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PasswordExpirationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan PasswordExpirationPolicyResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	policy := &PasswordExpirationPolicyAPIModel{
		Enabled:        plan.Enabled.ValueBool(),
		PasswordMaxAge: plan.PasswordMaxAge.ValueInt64(),
		NotifyByEmail:  plan.NotifyByEmail.ValueBool(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(policy).
		Put(PasswordExpirationPolicyEndpoint)

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

func (r *PasswordExpirationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state PasswordExpirationPolicyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	resp.Diagnostics.AddWarning(
		"Password Expiration Policy cannot be deleted",
		"Artifactory does not support deletion of the password expiration policy. Provider will disable the policy instead.",
	)

	// Convert from Terraform data model into API data model
	policy := &PasswordExpirationPolicyAPIModel{
		Enabled:        false,
		PasswordMaxAge: state.PasswordMaxAge.ValueInt64(),
		NotifyByEmail:  state.NotifyByEmail.ValueBool(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(policy).
		Put(PasswordExpirationPolicyEndpoint)

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
func (r *PasswordExpirationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
