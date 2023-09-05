/*
Package user supports the resource artifactory_user and artifactory_managed_user, which use the new terraform-plugin-framework

The truth table below shows how extra logic is needed to accommodate the behavior of Artifactory API for users while maintaining
backward compatibility with states created by SDKv2 provider.

Create
|   Config       |   Plan         |   PUT          |   POST  |   GET          |   State        |
|----------------|----------------|----------------|---------|----------------|----------------|
|   Not Defined  |   Null         |                |   []    |                |   Null         |
|   []           |   []           |   []           |   []    |                |   []           |
|   ["readers"]  |   ["readers"]  |   ["readers"]  |         |   ["readers"]  |   ["readers"]  |

Update
|   Config              |   Plan                |   POST                |   GET                 |   State               |
|-----------------------|-----------------------|-----------------------|-----------------------|-----------------------|
|   Not Defined         |   Null                |                       |                       |   Null                |
|   []                  |   []                  |   []                  |                       |   []                  |
|   ["readers", "foo"]  |   ["readers", "foo"]  |   ["readers", "foo"]  |   ["readers", "foo"]  |   ["readers", "foo"]  |
*/
package user

import (
	"context"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	Name                     string    `json:"name"`
	Email                    string    `json:"email"`
	Password                 string    `json:"password,omitempty"`
	Admin                    bool      `json:"admin"`
	ProfileUpdatable         bool      `json:"profileUpdatable"`
	DisableUIAccess          bool      `json:"disableUIAccess"`
	InternalPasswordDisabled bool      `json:"internalPasswordDisabled"`
	Groups                   *[]string `json:"groups,omitempty"`
}

var baseUserSchemaFramework = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		MarkdownDescription: "Username for user. May contain lowercase letters, numbers and symbols: '.-_@'",
		Required:            true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
			stringvalidator.RegexMatches(
				regexp.MustCompile(`^[a-z0-9.\-_\@]+$`),
				"may contain lowercase letters, numbers and symbols: '.-_@'",
			),
		},
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
		Default:             setdefault.StaticValue(types.SetNull(types.StringType)),
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
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
	var plan ArtifactoryUserResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	user := ArtifactoryUserResourceAPIModel{
		Name:                     plan.Name.ValueString(),
		Email:                    plan.Email.ValueString(),
		Password:                 plan.Password.ValueString(),
		Admin:                    plan.Admin.ValueBool(),
		ProfileUpdatable:         plan.ProfileUpdatable.ValueBool(),
		DisableUIAccess:          plan.DisableUIAccess.ValueBool(),
		InternalPasswordDisabled: plan.InternalPasswordDisabled.ValueBool(),
	}

	if !plan.Groups.IsNull() {
		groups := utilfw.StringSetToStrings(plan.Groups)
		user.Groups = &groups
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
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusCreated {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Artifactory PUT call for creating user with groups attribute set to empty/null always sets groups to "readers".
	// This is a bug on Artifactory. Below workaround will fix the issue and has to be removed after the artifactory bug is resolved.
	// Workaround: We use following POST call to update the user's groups config to empty group.
	// This action will match the expectation for this resource when "groups" attribute is empty or not specified in hcl.
	if plan.Groups.IsNull() || len(plan.Groups.Elements()) == 0 {
		user.Groups = &[]string{}
		_, errGroupUpdate := r.client.Client.R().SetBody(user).Post(UsersEndpointPath + user.Name)
		if errGroupUpdate != nil {
			utilfw.UnableToCreateResourceError(resp, response.String())
			return
		}

		// reset this back to nil to ensure TF state gets Null
		user.Groups = nil
	}

	// Parse user struct into the state
	resp.Diagnostics.Append(user.ToState(ctx, &plan)...) // not necessary with empty response, we only need an Id
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ArtifactoryBaseUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ArtifactoryUserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	user := ArtifactoryUserResourceAPIModel{}

	response, err := r.client.Client.R().SetResult(&user).Get(UsersEndpointPath + state.Id.ValueString())

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		if response.StatusCode() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(user.ToState(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ArtifactoryBaseUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ArtifactoryUserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var groups *[]string
	if !plan.Groups.IsNull() {
		g := utilfw.StringSetToStrings(plan.Groups)
		groups = &g
	}

	// Convert from Terraform data model into API data model
	user := ArtifactoryUserResourceAPIModel{
		Name:                     plan.Name.ValueString(),
		Email:                    plan.Email.ValueString(),
		Password:                 plan.Password.ValueString(),
		Admin:                    plan.Admin.ValueBool(),
		Groups:                   groups,
		ProfileUpdatable:         plan.ProfileUpdatable.ValueBool(),
		DisableUIAccess:          plan.DisableUIAccess.ValueBool(),
		InternalPasswordDisabled: plan.InternalPasswordDisabled.ValueBool(),
	}

	response, err := r.client.Client.R().SetBody(user).Post(UsersEndpointPath + user.Name)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	user.ToState(ctx, &plan)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ArtifactoryBaseUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ArtifactoryUserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	response, err := r.client.Client.R().Delete(UsersEndpointPath + state.Id.ValueString())

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 200 OK or 404 Not Found
	if response.StatusCode() != http.StatusNotFound && response.StatusCode() != http.StatusOK {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ArtifactoryBaseUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

}
func (u ArtifactoryUserResourceAPIModel) ToState(ctx context.Context, r *ArtifactoryUserResourceModel) diag.Diagnostics {
	r.Id = types.StringValue(u.Name)
	r.Name = types.StringValue(u.Name)
	r.Email = types.StringValue(u.Email)
	r.Admin = types.BoolValue(u.Admin)
	r.ProfileUpdatable = types.BoolValue(u.ProfileUpdatable)
	r.DisableUIAccess = types.BoolValue(u.DisableUIAccess)
	r.InternalPasswordDisabled = types.BoolValue(u.InternalPasswordDisabled)

	// if Groups attribute is set to [] and GET returns null then make sure state has empty set
	if !r.Groups.IsNull() && len(r.Groups.Elements()) == 0 && u.Groups == nil {
		r.Groups = types.SetValueMust(types.StringType, []attr.Value{})
	}

	if u.Groups != nil {
		groups, diags := types.SetValueFrom(ctx, types.StringType, u.Groups)
		if diags.HasError() {
			return diags
		}
		r.Groups = groups
	}

	return nil
}
