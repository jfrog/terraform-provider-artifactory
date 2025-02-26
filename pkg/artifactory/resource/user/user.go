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
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	"github.com/samber/lo"
	"github.com/sethvargo/go-password/password"
)

const (
	AccessAPIArtifactoryVersion = "7.84.3"
	UserGroupEndpointPath       = "access/api/v2/users/{name}/groups"
)

type ArtifactoryBaseUserResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

// ArtifactoryUserResourceModel describes the Terraform resource data model to match the
// resource schema.
type ArtifactoryUserResourceModelV0 struct {
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

type ArtifactoryUserResourceModel struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Email                    types.String `tfsdk:"email"`
	Password                 types.String `tfsdk:"password"`
	PasswordPolicy           types.Object `tfsdk:"password_policy"`
	Admin                    types.Bool   `tfsdk:"admin"`
	ProfileUpdatable         types.Bool   `tfsdk:"profile_updatable"`
	DisableUIAccess          types.Bool   `tfsdk:"disable_ui_access"`
	InternalPasswordDisabled types.Bool   `tfsdk:"internal_password_disabled"`
	Groups                   types.Set    `tfsdk:"groups"`
}

// ArtifactoryUserResourceAPIModel describes the API data model.
type ArtifactoryUserResourceAPIModel struct {
	Name                     string    `json:"username"`
	Email                    string    `json:"email"`
	Password                 string    `json:"password,omitempty"`
	Admin                    bool      `json:"admin"`
	ProfileUpdatable         bool      `json:"profile_updatable"`
	DisableUIAccess          bool      `json:"disable_ui_access"`
	InternalPasswordDisabled *bool     `json:"internal_password_disabled,omitempty"`
	Groups                   *[]string `json:"groups,omitempty"`
}

func (u ArtifactoryUserResourceAPIModel) ToState(ctx context.Context, r *ArtifactoryUserResourceModel) diag.Diagnostics {
	r.Id = types.StringValue(u.Name)
	r.Name = types.StringValue(u.Name)
	r.Email = types.StringValue(u.Email)

	if r.Password.IsUnknown() {
		r.Password = types.StringNull()
	}

	r.Admin = types.BoolValue(u.Admin)
	r.ProfileUpdatable = types.BoolValue(u.ProfileUpdatable)
	r.DisableUIAccess = types.BoolValue(u.DisableUIAccess)

	if u.InternalPasswordDisabled != nil {
		r.InternalPasswordDisabled = types.BoolPointerValue(u.InternalPasswordDisabled)
	}

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

var passwordPolicyAttributeTypes = map[string]attr.Type{
	"uppercase":    types.Int64Type,
	"lowercase":    types.Int64Type,
	"special_char": types.Int64Type,
	"digit":        types.Int64Type,
	"length":       types.Int64Type,
}

var baseUserSchemaFramework = lo.Assign(
	baseUserSchemaFrameworkV0,
	map[string]schema.Attribute{
		"password_policy": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"uppercase": schema.Int64Attribute{
					Optional:    true,
					Description: "Minimum number of uppercase letters that the password must contain",
				},
				"lowercase": schema.Int64Attribute{
					Optional:    true,
					Description: "Minimum number of lowercase letters that the password must contain",
				},
				"special_char": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "Minimum number of special char that the password must contain. Special chars list: ``!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~``",
				},
				"digit": schema.Int64Attribute{
					Optional:    true,
					Description: "Minimum number of digits that the password must contain",
				},
				"length": schema.Int64Attribute{
					Optional:    true,
					Description: "Minimum length of the password",
				},
			},
			Optional: true,
			MarkdownDescription: "Password policy to match JFrog Access to provide validation before API request.\n\n" +
				"->Due to Terraform limitation with interpolated value, we can only validate interpolated value prior to making API requests. This means `terraform validate` or `terraform plan` will not return error if `password` does not meet `password_policy` criteria.\n\n" +
				"Default values: `uppercase=1`, `lowercase=1`, `special_char=0`, `digit=1`, `length=8`. Also see [Supported Access Configurations](https://jfrog.com/help/r/jfrog-installation-setup-documentation/supported-access-configurations) for more details",
		},
	},
)

var baseUserSchemaFrameworkV0 = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		MarkdownDescription: "Username for user. May contain lowercase letters, numbers and symbols: '.-_@' for self-hosted. For SaaS, '+' is also allowed.",
		Required:            true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
			stringvalidator.RegexMatches(
				regexp.MustCompile(`^[a-z0-9.\-_\@\+]+$`),
				"may contain lowercase letters, numbers and symbols: '.-_@' for self-hosted. For SaaS, '+' is also allowed.",
			),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"email": schema.StringAttribute{
		MarkdownDescription: "Email for user.",
		Required:            true,
	},
	"password": schema.StringAttribute{
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
		MarkdownDescription: "List of groups this user is a part of. **Notes:** If this attribute is not specified then user's group membership is set to empty. User will not be part of default \"readers\" group automatically.",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
	},
}

func (r *ArtifactoryBaseUserResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 1 (Schema.Version)
		0: {
			PriorSchema: &schema.Schema{
				Attributes: baseUserSchemaFrameworkV0,
			},
			// Optionally, the PriorSchema field can be defined.
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData ArtifactoryUserResourceModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgradedStateData := ArtifactoryUserResourceModel{
					Id:                       priorStateData.Id,
					Name:                     priorStateData.Name,
					Email:                    priorStateData.Email,
					Password:                 priorStateData.Password,
					PasswordPolicy:           types.ObjectNull(passwordPolicyAttributeTypes),
					Admin:                    priorStateData.Admin,
					ProfileUpdatable:         priorStateData.ProfileUpdatable,
					DisableUIAccess:          priorStateData.DisableUIAccess,
					InternalPasswordDisabled: priorStateData.InternalPasswordDisabled,
					Groups:                   priorStateData.Groups,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
	}
}

func (r *ArtifactoryBaseUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *ArtifactoryBaseUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r ArtifactoryBaseUserResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ArtifactoryUserResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.validatePasswordByPolicy(data))
	if resp.Diagnostics.HasError() {
		return
	}
}

type GroupsAddRemove struct {
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

func (r *ArtifactoryBaseUserResource) syncReadersGroup(_ context.Context, client *resty.Client, plan ArtifactoryUserResourceAPIModel, actual ArtifactoryUserResourceAPIModel) error {
	planGroups := []string{}
	if plan.Groups != nil {
		planGroups = *plan.Groups
	}
	actualGroups := []string{}
	if actual.Groups != nil {
		actualGroups = *actual.Groups
	}
	toAdd, toRemove := lo.Difference(planGroups, actualGroups)

	if len(toAdd) == 0 && len(toRemove) == 0 {
		return nil
	}

	var artifactoryError artifactory.ArtifactoryErrorsResponse
	groupsToAddRemove := GroupsAddRemove{
		Add:    toAdd,
		Remove: toRemove,
	}
	// Access API for creating user will add any groups with "auto_join = true" to the user's groups.
	// We use following PATCH call to sync up user's groups from TF to Artifactory.
	// This action will match the expectation for this resource so "groups" attribute matches what's on Artifactory.
	resp, err := client.R().
		SetPathParam("name", actual.Name).
		SetBody(groupsToAddRemove).
		SetError(&artifactoryError).
		Patch(UserGroupEndpointPath)
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("%s", artifactoryError.String())
	}

	return nil
}

func GetUsersEndpointPath(artifactoryVersion string) string {
	if ok, err := util.CheckVersion(artifactoryVersion, AccessAPIArtifactoryVersion); err == nil && ok {
		return "access/api/v2/users"
	}

	return "artifactory/api/security/users"
}

func GetUserEndpointPath(artifactoryVersion string) string {
	if ok, err := util.CheckVersion(artifactoryVersion, AccessAPIArtifactoryVersion); err == nil && ok {
		return "access/api/v2/users/{name}"
	}

	return "artifactory/api/security/users/{name}"
}

// ArtifactoryUserAPIModel corresponds to old Artifactory user API
type ArtifactoryUserAPIModel struct {
	Name                     string    `json:"name"`
	Email                    string    `json:"email"`
	Password                 string    `json:"password,omitempty"`
	Admin                    bool      `json:"admin"`
	ProfileUpdatable         bool      `json:"profileUpdatable"`
	DisableUIAccess          bool      `json:"disableUIAccess"`
	InternalPasswordDisabled *bool     `json:"internalPasswordDisabled"`
	Groups                   *[]string `json:"groups,omitempty"`
}

func (r *ArtifactoryBaseUserResource) createUser(_ context.Context, req *resty.Request, artifactoryVersion string, user ArtifactoryUserResourceAPIModel, result *ArtifactoryUserResourceAPIModel, artifactoryError *artifactory.ArtifactoryErrorsResponse) (*resty.Response, error) {
	// 7.84.3 or later, use Access API
	if ok, err := util.CheckVersion(artifactoryVersion, AccessAPIArtifactoryVersion); err == nil && ok {
		return req.
			SetBody(user).
			SetResult(result).
			SetError(artifactoryError).
			Post(GetUsersEndpointPath(artifactoryVersion))
	}

	// else use old Artifactory API, which has a slightly differect JSON payload!
	artifactoryUser := ArtifactoryUserAPIModel(user)
	endpoint := GetUserEndpointPath(artifactoryVersion)
	resp, err := req.
		SetPathParam("name", artifactoryUser.Name).
		SetBody(artifactoryUser).
		SetError(artifactoryError).
		Put(endpoint)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return resp, nil
	}

	var artifactoryResult ArtifactoryUserAPIModel
	res, err := req.
		SetPathParam("name", artifactoryUser.Name).
		SetResult(&artifactoryResult).
		SetError(artifactoryError).
		Get(endpoint)

	*result = ArtifactoryUserResourceAPIModel{
		Name:                     artifactoryResult.Name,
		Email:                    artifactoryResult.Email,
		Password:                 user.Password,
		Admin:                    artifactoryResult.Admin,
		ProfileUpdatable:         artifactoryResult.ProfileUpdatable,
		DisableUIAccess:          artifactoryResult.DisableUIAccess,
		InternalPasswordDisabled: artifactoryResult.InternalPasswordDisabled,
		Groups:                   artifactoryResult.Groups,
	}

	return res, err
}

func (r *ArtifactoryBaseUserResource) readUser(req *resty.Request, artifactoryVersion, name string, result *ArtifactoryUserResourceAPIModel, artifactoryError *artifactory.ArtifactoryErrorsResponse) (*resty.Response, error) {
	endpoint := GetUserEndpointPath(artifactoryVersion)

	// 7.84.3 or later, use Access API
	if ok, err := util.CheckVersion(artifactoryVersion, AccessAPIArtifactoryVersion); err == nil && ok {
		return req.
			SetPathParam("name", name).
			SetResult(&result).
			SetError(&artifactoryError).
			Get(endpoint)
	}

	// else use old Artifactory API, which has a slightly differect JSON payload!
	var artifactoryResult ArtifactoryUserAPIModel
	res, err := req.
		SetPathParam("name", name).
		SetResult(&artifactoryResult).
		SetError(artifactoryError).
		Get(endpoint)

	*result = ArtifactoryUserResourceAPIModel{
		Name:                     artifactoryResult.Name,
		Email:                    artifactoryResult.Email,
		Admin:                    artifactoryResult.Admin,
		ProfileUpdatable:         artifactoryResult.ProfileUpdatable,
		DisableUIAccess:          artifactoryResult.DisableUIAccess,
		InternalPasswordDisabled: artifactoryResult.InternalPasswordDisabled,
		Groups:                   artifactoryResult.Groups,
	}

	return res, err
}

func (r *ArtifactoryBaseUserResource) updateUser(req *resty.Request, artifactoryVersion string, user ArtifactoryUserResourceAPIModel, result *ArtifactoryUserResourceAPIModel, artifactoryError *artifactory.ArtifactoryErrorsResponse) (*resty.Response, error) {
	endpoint := GetUserEndpointPath(artifactoryVersion)

	// 7.84.3 or later, use Access API
	if ok, err := util.CheckVersion(artifactoryVersion, AccessAPIArtifactoryVersion); err == nil && ok {
		return req.
			SetPathParam("name", user.Name).
			SetBody(user).
			SetResult(result).
			SetError(artifactoryError).
			Patch(endpoint)
	}

	// else use old Artifactory API, which has a slightly differect JSON payload!
	artifactoryUser := ArtifactoryUserAPIModel(user)
	resp, err := req.
		SetPathParam("name", artifactoryUser.Name).
		SetBody(artifactoryUser).
		SetError(artifactoryError).
		Post(endpoint)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return resp, nil
	}

	var artifactoryResult ArtifactoryUserAPIModel
	res, err := req.
		SetPathParam("name", artifactoryUser.Name).
		SetResult(&artifactoryResult).
		SetError(artifactoryError).
		Get(endpoint)

	*result = ArtifactoryUserResourceAPIModel{
		Name:                     artifactoryResult.Name,
		Email:                    artifactoryResult.Email,
		Password:                 user.Password,
		Admin:                    artifactoryResult.Admin,
		ProfileUpdatable:         artifactoryResult.ProfileUpdatable,
		DisableUIAccess:          artifactoryResult.DisableUIAccess,
		InternalPasswordDisabled: artifactoryResult.InternalPasswordDisabled,
		Groups:                   artifactoryResult.Groups,
	}

	return res, err
}

func (r *ArtifactoryBaseUserResource) validatePasswordByPolicy(plan ArtifactoryUserResourceModel) diag.Diagnostic {
	// If password is not configured then no need to validate
	if plan.Password.IsNull() || plan.Password.IsUnknown() {
		return nil
	}

	if plan.PasswordPolicy.IsUnknown() {
		return nil
	}

	// Default password policy should match Access default configuration:
	// https://jfrog.com/help/r/jfrog-installation-setup-documentation/supported-access-configurations
	minLength := int64(8)
	lowercaseLength := int64(1)
	uppercaseLength := int64(1)
	specialCharLength := int64(0)
	digitLength := int64(1)

	// If password_policy is configured, overwrite default values
	if !plan.PasswordPolicy.IsNull() {
		attrs := plan.PasswordPolicy.Attributes()

		if v, ok := attrs["length"]; ok {
			minLength = v.(types.Int64).ValueInt64()
		}

		if v, ok := attrs["lowercase"]; ok {
			lowercaseLength = v.(types.Int64).ValueInt64()
		}

		if v, ok := attrs["uppercase"]; ok {
			uppercaseLength = v.(types.Int64).ValueInt64()
		}

		if v, ok := attrs["special_char"]; ok {
			specialCharLength = v.(types.Int64).ValueInt64()
		}

		if v, ok := attrs["digit"]; ok {
			digitLength = v.(types.Int64).ValueInt64()
		}
	}

	password := plan.Password.ValueString()

	if len(password) < int(minLength) {
		return diag.NewAttributeErrorDiagnostic(
			path.Root("password"),
			"Invalid Attribute Value Length",
			fmt.Sprintf(
				"Attribute password string length must be at least %d, got %d",
				minLength,
				len(password),
			),
		)
	}

	lowercaseRegex := regexp.MustCompile("[a-z]")
	matched := lowercaseRegex.FindAllString(password, -1)
	if len(matched) < int(lowercaseLength) {
		return diag.NewAttributeErrorDiagnostic(
			path.Root("password"),
			"Invalid Attribute Value Match",
			fmt.Sprintf(
				"Attribute password string must have at least %d lowercase letters",
				lowercaseLength,
			),
		)
	}

	uppercaseRegex := regexp.MustCompile("[A-Z]")
	matched = uppercaseRegex.FindAllString(password, -1)
	if len(matched) < int(uppercaseLength) {
		return diag.NewAttributeErrorDiagnostic(
			path.Root("password"),
			"Invalid Attribute Value Match",
			fmt.Sprintf(
				"Attribute password string must have at least %d uppercase letters",
				uppercaseLength,
			),
		)
	}

	specialCharRegex := regexp.MustCompile(`[!"#$%%&'()\*\+,\-\./:;<=>?@\[\\\]^_\x60{|}~]`)
	matched = specialCharRegex.FindAllString(password, -1)
	if len(matched) < int(specialCharLength) {
		return diag.NewAttributeErrorDiagnostic(
			path.Root("password"),
			"Invalid Attribute Value Match",
			fmt.Sprintf(
				"Attribute password string must have at least %d special characters",
				specialCharLength,
			),
		)
	}

	digitRegex := regexp.MustCompile(`\d`)
	matched = digitRegex.FindAllString(password, -1)
	if len(matched) < int(digitLength) {
		return diag.NewAttributeErrorDiagnostic(
			path.Root("password"),
			"Invalid Attribute Value Match",
			fmt.Sprintf(
				"Attribute password string must have at least %d digits",
				digitLength,
			),
		)
	}

	return nil
}

func (r *ArtifactoryBaseUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ArtifactoryUserResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.InternalPasswordDisabled.ValueBool() {
		resp.Diagnostics.Append(r.validatePasswordByPolicy(plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Convert from Terraform data model into API data model
	user := ArtifactoryUserResourceAPIModel{
		Name:                     plan.Name.ValueString(),
		Email:                    plan.Email.ValueString(),
		Password:                 plan.Password.ValueString(),
		Admin:                    plan.Admin.ValueBool(),
		ProfileUpdatable:         plan.ProfileUpdatable.ValueBool(),
		DisableUIAccess:          plan.DisableUIAccess.ValueBool(),
		InternalPasswordDisabled: plan.InternalPasswordDisabled.ValueBoolPointer(),
	}

	if !plan.Groups.IsNull() && len(plan.Groups.Elements()) > 0 {
		groups := utilfw.StringSetToStrings(plan.Groups)
		user.Groups = &groups
	}

	if user.Password == "" && (user.InternalPasswordDisabled == nil || !*user.InternalPasswordDisabled) {
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

		// DO NOT store the generated password in the TF state
	}

	var result ArtifactoryUserResourceAPIModel
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.createUser(
		ctx,
		r.ProviderData.Client.R(),
		r.ProviderData.ArtifactoryVersion,
		user,
		&result,
		&artifactoryError)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, artifactoryError.String())
		return
	}

	err = r.syncReadersGroup(ctx, r.ProviderData.Client, user, result)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
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
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ArtifactoryUserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var user ArtifactoryUserResourceAPIModel
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.readUser(
		r.ProviderData.Client.R(),
		r.ProviderData.ArtifactoryVersion,
		state.Name.ValueString(),
		&user,
		&artifactoryError)

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, artifactoryError.String())
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
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ArtifactoryUserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state ArtifactoryUserResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// Set internalPasswordDisabled pointer to non-nil value if it's been changed
	var internalPasswordDisabled *bool
	if !plan.InternalPasswordDisabled.Equal(state.InternalPasswordDisabled) {
		internalPasswordDisabled = plan.InternalPasswordDisabled.ValueBoolPointer()
	}

	// If 'internal_password_disabled' changes to 'false' AND 'password' is not set,
	// error out
	if (internalPasswordDisabled != nil && !*internalPasswordDisabled) &&
		(plan.Password.IsNull() || plan.Password.IsUnknown()) {
		resp.Diagnostics.AddError(
			"Password must be set",
			"Password must be set when internal_password_disabled is changed to 'false'",
		)
		return
	}

	if !plan.InternalPasswordDisabled.ValueBool() {
		resp.Diagnostics.Append(r.validatePasswordByPolicy(plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var groups *[]string
	if !plan.Groups.IsNull() && len(plan.Groups.Elements()) > 0 {
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
		InternalPasswordDisabled: internalPasswordDisabled,
	}

	var result ArtifactoryUserResourceAPIModel
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.updateUser(
		r.ProviderData.Client.R(),
		r.ProviderData.ArtifactoryVersion,
		user,
		&result,
		&artifactoryError)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, artifactoryError.String())
		return
	}

	err = r.syncReadersGroup(ctx, r.ProviderData.Client, user, result)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	user.ToState(ctx, &plan)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ArtifactoryBaseUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ArtifactoryUserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var artifactoryError artifactory.ArtifactoryErrorsResponse
	response, err := r.ProviderData.Client.R().
		SetPathParam("name", state.Name.ValueString()).
		SetError(&artifactoryError).
		Delete(GetUserEndpointPath(r.ProviderData.ArtifactoryVersion))

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK, 204 No Content, or 404 Not Found
	if !(response.StatusCode() == http.StatusNotFound ||
		response.StatusCode() == http.StatusOK ||
		response.StatusCode() == http.StatusNoContent) {
		utilfw.UnableToDeleteResourceError(resp, artifactoryError.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ArtifactoryBaseUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
