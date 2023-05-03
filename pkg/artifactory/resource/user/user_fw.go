package user

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/sethvargo/go-password/password"
)

type ArtifactoryBaseUserResource struct {
	client utilsdk.ProvderMetadata
}

// ArtifactoryUserResourceModel describes the Terraform resource data model to match the
// resource schema.
type ArtifactoryUserResourceModel struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Email                    types.String `tfsdk:"email"`
	Password                 types.String `tfsdk:"password"`
	Admin                    types.Bool   `tfsdk:"admin"`
	ProfileUpdatable         types.Bool   `tfsdk:"profile_updatable"`
	DisableUIAccess          types.Bool   `tfsdk:"disable_ui_access"`
	InternalPasswordDisabled types.Bool   `tfsdk:"internal_password_disabled"`
	Groups                   types.Set    `tfsdk:"groups"`
}

// ArtifactoryUserResourceAPIModel describes the API data model.
type ArtifactoryUserResourceAPIModel struct {
	Name                     string   `json:"name"`
	Email                    string   `json:"email"`
	Password                 string   `json:"password,omitempty"`
	Admin                    bool     `json:"admin"`
	ProfileUpdatable         bool     `json:"profileUpdatable"`
	DisableUIAccess          bool     `json:"disableUIAccess"`
	InternalPasswordDisabled bool     `json:"internalPasswordDisabled"`
	Groups                   []string `json:"groups"`
}

var baseUserSchemaFramework = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
	},
	"name": schema.StringAttribute{
		MarkdownDescription: "Username for user.",
		Required:            true,
	},
	"email": schema.StringAttribute{
		MarkdownDescription: "Email for user.",
		Required:            true,
	},
	"password": schema.StringAttribute{
		MarkdownDescription: "(Optional, Sensitive) Password for the user. When omitted, a random password is generated using the following password policy: " +
			"12 characters with 1 digit, 1 symbol, with upper and lower case letters",
		Optional:  true,
		Sensitive: true,
	},
	"admin": schema.BoolAttribute{
		MarkdownDescription: "(Optional, Default: false) When enabled, this user is an administrator with all the ensuing privileges.",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	},
	"profile_updatable": schema.BoolAttribute{
		MarkdownDescription: "(Optional, Default: true) When enabled, this user can update their profile details (except for the password. " +
			"Only an administrator can update the password). There may be cases in which you want to leave " +
			"this unset to prevent users from updating their profile. For example, a departmental user with " +
			"a single password shared between all department members.",
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(true),
	},
	"disable_ui_access": schema.BoolAttribute{
		MarkdownDescription: "(Optional, Default: true) When enabled, this user can only access the system through the REST API." +
			" This option cannot be set if the user has Admin privileges.",
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(true),
	},
	"internal_password_disabled": schema.BoolAttribute{
		MarkdownDescription: "(Optional, Default: false) When enabled, disables the fallback mechanism for using an internal password when " +
			"external authentication (such as LDAP) is enabled.",
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(false),
	},
	"groups": schema.SetAttribute{
		MarkdownDescription: "List of groups this user is a part of. If no groups set, `readers` group will be added by default. If other groups are assigned, `readers` must be added to the list manually to avoid state drift.",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Default: setdefault.StaticValue(
			types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("readers"),
				},
			),
		),
	},
}

func (r *ArtifactoryBaseUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *ArtifactoryBaseUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ArtifactoryUserResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groups := utilfw.StringSetToStrings(data.Groups)
	// Convert from Terraform data model into API data model
	user := &ArtifactoryUserResourceAPIModel{
		Name:                     data.Name.ValueString(),
		Email:                    data.Email.ValueString(),
		Password:                 data.Password.ValueString(),
		Admin:                    data.Admin.ValueBool(),
		ProfileUpdatable:         data.ProfileUpdatable.ValueBool(),
		DisableUIAccess:          data.DisableUIAccess.ValueBool(),
		InternalPasswordDisabled: data.InternalPasswordDisabled.ValueBool(),
		Groups:                   groups,
	}

	if user.Password == "" {
		resp.Diagnostics.AddWarning(
			"No password supplied",
			"One will be generated (12 characters with 1 digit, 1 symbol, with upper and lower case letters) and this may fail as your Artifactory password policy can't be determined by the provider.",
		)
		// Generate a password that is 12 characters long with 1 digit, 1 symbol,
		// allowing upper and lower case letters, disallowing repeat characters.
		randomPassword, err := password.Generate(12, 1, 1, false, false)
		if err != nil {
			resp.Diagnostics.AddError(
				"failed to generate password",
				"Error: "+err.Error(),
			)

			return
		}

		user.Password = randomPassword
	}

	response, err := r.client.Client.R().SetBody(user).Put(UsersEndpointPath + user.Name)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Status: "+response.Status(),
		)

		return
	}

	// Artifactory PUT call for creating user with groups attribute set to empty/null always sets groups to "readers".
	// This is a bug on Artifactory. Below workaround will fix the issue and has to be removed after the artifactory bug is resolved.
	// Workaround: We use following POST call to update the user's groups config to empty group.
	// This action will match the expectation for this resource when "groups" attribute is empty or not specified in hcl.
	if len(user.Groups) == 0 {
		_, errGroupUpdate := r.client.Client.R().SetBody(user).Post(UsersEndpointPath + user.Name)
		if errGroupUpdate != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while attempting to create the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"HTTP Status: "+response.Status(),
			)

			return
		}
	}

	// Parse user struct into the state
	data.ToState(ctx, user) // not necessary with empty response, we only need an Id
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *ArtifactoryBaseUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ArtifactoryUserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	user := &ArtifactoryUserResourceAPIModel{}

	response, err := r.client.Client.R().SetResult(user).Get(UsersEndpointPath + data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Refresh Resource",
			"An unexpected error occurred while attempting to refresh resource state. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	data.ToState(ctx, user)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryBaseUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ArtifactoryUserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	groups := utilfw.StringSetToStrings(data.Groups)
	// Convert from Terraform data model into API data model
	user := &ArtifactoryUserResourceAPIModel{
		Name:                     data.Name.ValueString(),
		Email:                    data.Email.ValueString(),
		Password:                 data.Password.ValueString(),
		Admin:                    data.Admin.ValueBool(),
		ProfileUpdatable:         data.ProfileUpdatable.ValueBool(),
		DisableUIAccess:          data.DisableUIAccess.ValueBool(),
		InternalPasswordDisabled: data.InternalPasswordDisabled.ValueBool(),
		Groups:                   groups,
	}

	response, err := r.client.Client.R().SetBody(user).Post(UsersEndpointPath + user.Name)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while creating the resource update request. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Status: "+response.Status(),
		)

		return
	}

	data.ToState(ctx, user)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryBaseUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ArtifactoryUserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	response, err := r.client.Client.R().Delete(UsersEndpointPath + data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}

	// Return error if the HTTP status code is not 200 OK or 404 Not Found
	if response.StatusCode() != http.StatusNotFound && response.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Status: "+response.Status(),
		)

		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ArtifactoryBaseUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

}
func (r *ArtifactoryUserResourceModel) ToState(ctx context.Context, user *ArtifactoryUserResourceAPIModel) {
	r.Id = types.StringValue(user.Name)
	r.Name = types.StringValue(user.Name)
	r.Email = types.StringValue(user.Email)
	r.Admin = types.BoolValue(user.Admin)
	r.ProfileUpdatable = types.BoolValue(user.ProfileUpdatable)
	r.DisableUIAccess = types.BoolValue(user.DisableUIAccess)
	r.InternalPasswordDisabled = types.BoolValue(user.InternalPasswordDisabled)
	if user.Groups != nil {
		r.Groups, _ = types.SetValueFrom(ctx, types.StringType, user.Groups)
	}
}
