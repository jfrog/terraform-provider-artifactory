package security

import (
	"context"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	validatorfw "github.com/jfrog/terraform-provider-shared/validator/fw"
)

const GroupsEndpoint = "artifactory/api/security/groups/"

func NewGroupResource() resource.Resource {
	return &ArtifactoryGroupResource{}
}

type ArtifactoryGroupResource struct {
	ProviderData utilsdk.ProvderMetadata
}

// ArtifactoryGroupResourceModel describes the Terraform resource data model to match the
// resource schema.
type ArtifactoryGroupResourceModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	ExternalId      types.String `tfsdk:"external_id"`
	AutoJoin        types.Bool   `tfsdk:"auto_join"`
	AdminPrivileges types.Bool   `tfsdk:"admin_privileges"`
	Realm           types.String `tfsdk:"realm"`
	RealmAttributes types.String `tfsdk:"realm_attributes"`
	DetachAllUsers  types.Bool   `tfsdk:"detach_all_users"`
	UsersNames      types.Set    `tfsdk:"users_names"`
	WatchManager    types.Bool   `tfsdk:"watch_manager"`
	PolicyManager   types.Bool   `tfsdk:"policy_manager"`
	ReportsManager  types.Bool   `tfsdk:"reports_manager"`
}

// ArtifactoryGroupResourceAPIModel describes the API data model.
type ArtifactoryGroupResourceAPIModel struct {
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	ExternalId      string   `json:"externalId,omitempty"`
	AutoJoin        bool     `json:"autoJoin"`
	AdminPrivileges bool     `json:"adminPrivileges"`
	Realm           string   `json:"realm"`
	RealmAttributes string   `json:"realmAttributes,omitempty"`
	UsersNames      []string `json:"userNames"`
	WatchManager    bool     `json:"watchManager"`
	PolicyManager   bool     `json:"policyManager"`
	ReportsManager  bool     `json:"reportsManager"`
}

func (r *ArtifactoryGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_group"
}

func (r *ArtifactoryGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory group resource. This can be used to create and manage Artifactory groups. A group represents a role in the system and is assigned a set of permissions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the group.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description for the group.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_id": schema.StringAttribute{
				MarkdownDescription: "New external group ID used to configure the corresponding group in Azure AD.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_join": schema.BoolAttribute{
				MarkdownDescription: "When this parameter is set, any new users defined in the system are automatically assigned to this group.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				Validators: []validator.Bool{
					validatorfw.BoolConflict(true, path.Expressions{
						path.MatchRelative().AtParent().AtName("admin_privileges"),
					}...),
				},
			},
			"admin_privileges": schema.BoolAttribute{
				MarkdownDescription: "Any users added to this group will automatically be assigned with admin privileges in the system.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"realm": schema.StringAttribute{
				MarkdownDescription: "The realm for the group.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("internal"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"realm_attributes": schema.StringAttribute{
				MarkdownDescription: "The realm attributes for the group.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"users_names": schema.SetAttribute{
				MarkdownDescription: "List of users assigned to the group. If not set or empty, Terraform will not manage group membership.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(types.SetNull(types.StringType)),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"detach_all_users": schema.BoolAttribute{
				MarkdownDescription: "When this is set to `true`, an empty or missing usernames array will detach all users from the group.",
				Optional:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"watch_manager": schema.BoolAttribute{
				MarkdownDescription: "When this override is set, User in the group can manage Xray Watches on any resource type. Default value is `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"policy_manager": schema.BoolAttribute{
				MarkdownDescription: "When this override is set, User in the group can set Xray security and compliance policies. Default value is `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"reports_manager": schema.BoolAttribute{
				MarkdownDescription: "When this override is set, User in the group can manage Xray Reports on any resource type. Default value is `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *ArtifactoryGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *ArtifactoryGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ArtifactoryGroupResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	group := &ArtifactoryGroupResourceAPIModel{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		ExternalId:      data.ExternalId.ValueString(),
		AutoJoin:        data.AutoJoin.ValueBool(),
		AdminPrivileges: data.AdminPrivileges.ValueBool(),
		Realm:           data.Realm.ValueString(),
		RealmAttributes: data.RealmAttributes.ValueString(),
		WatchManager:    data.WatchManager.ValueBool(),
		PolicyManager:   data.PolicyManager.ValueBool(),
		ReportsManager:  data.ReportsManager.ValueBool(),
	}
	if !data.UsersNames.IsNull() {
		usersNames := utilfw.StringSetToStrings(data.UsersNames)
		group.UsersNames = usersNames
	}

	response, err := r.ProviderData.Client.R().
		SetBody(group).
		Put(GroupsEndpoint + group.Name)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusCreated {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Assign the resource ID for the resource in the state
	data.Id = types.StringValue(group.Name)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getDetachUsersValue(resource *ArtifactoryGroupResourceModel) bool {
	detachAllUsers := false
	if !resource.DetachAllUsers.IsNull() && !resource.DetachAllUsers.IsUnknown() {
		detachAllUsers = resource.DetachAllUsers.ValueBool()
	}

	return detachAllUsers
}

func (r *ArtifactoryGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ArtifactoryGroupResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	group := ArtifactoryGroupResourceAPIModel{}

	includeUsers := len(data.UsersNames.Elements()) > 0 || getDetachUsersValue(data)

	response, err := r.ProviderData.Client.R().
		SetQueryParam("includeUsers", strconv.FormatBool(includeUsers)).
		SetResult(&group).
		Get(GroupsEndpoint + data.Id.ValueString())

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
	resp.Diagnostics.Append(data.ToState(ctx, group)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ArtifactoryGroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Convert from Terraform data model into API data model
	usersNames := utilfw.StringSetToStrings(data.UsersNames)
	group := ArtifactoryGroupResourceAPIModel{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		ExternalId:      data.ExternalId.ValueString(),
		AutoJoin:        data.AutoJoin.ValueBool(),
		AdminPrivileges: data.AdminPrivileges.ValueBool(),
		Realm:           data.Realm.ValueString(),
		RealmAttributes: data.RealmAttributes.ValueString(),
		UsersNames:      usersNames,
		WatchManager:    data.WatchManager.ValueBool(),
		PolicyManager:   data.PolicyManager.ValueBool(),
		ReportsManager:  data.ReportsManager.ValueBool(),
	}

	// Create and Update uses same endpoint, create checks for ReplaceIfExists and then uses PUT
	// This recreates the group with the same permissions and updated users.
	// Update instead uses POST which prevents removing users and since it is only used when membership is empty
	// this results in a group where users are not managed by artifactory if users_names is not set.
	includeUsers := len(group.UsersNames) > 0 || getDetachUsersValue(data)
	if includeUsers {
		// Create call
		response, err := r.ProviderData.Client.R().
			SetBody(&group).
			Put(GroupsEndpoint + group.Name)
		if err != nil {
			utilfw.UnableToUpdateResourceError(resp, err.Error())
			return
		}

		// Return error if the HTTP status code is not 200 OK
		if response.StatusCode() != http.StatusCreated {
			utilfw.UnableToUpdateResourceError(resp, response.String())
			return
		}
	} else {
		// Update call
		response, err := r.ProviderData.Client.R().
			SetBody(group).
			Post(GroupsEndpoint + group.Name)
		if err != nil {
			utilfw.UnableToUpdateResourceError(resp, err.Error())
			return
		}

		// Return error if the HTTP status code is not 200 OK
		if response.StatusCode() != http.StatusOK {
			utilfw.UnableToUpdateResourceError(resp, response.String())
			return
		}
	}

	resp.Diagnostics.Append(data.ToState(ctx, group)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ArtifactoryGroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	response, err := r.ProviderData.Client.R().
		Delete(GroupsEndpoint + data.Id.ValueString())

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
func (r *ArtifactoryGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ArtifactoryGroupResourceModel) ToState(ctx context.Context, group ArtifactoryGroupResourceAPIModel) diag.Diagnostics {
	r.Id = types.StringValue(group.Name)
	r.Name = types.StringValue(group.Name)

	if r.Description.IsNull() {
		r.Description = types.StringValue("")
	}
	if group.Description != "" {
		r.Description = types.StringValue(group.Description)
	}

	if r.ExternalId.IsNull() {
		r.ExternalId = types.StringValue("")
	}
	if group.ExternalId != "" {
		r.ExternalId = types.StringValue(group.ExternalId)
	}
	r.AutoJoin = types.BoolValue(group.AutoJoin)
	r.AdminPrivileges = types.BoolValue(group.AdminPrivileges)
	r.Realm = types.StringValue(group.Realm)

	// Need to set empty string for null state value to avoid state drift.
	// See https://discuss.hashicorp.com/t/diffsuppressfunc-alternative-in-terraform-framework/52578/2?u=alexhung
	if r.RealmAttributes.IsNull() {
		r.RealmAttributes = types.StringValue("")
	}
	if group.RealmAttributes != "" {
		r.RealmAttributes = types.StringValue(group.RealmAttributes)
	}

	// We have to check if the value is not null to prevent an error "...unexpected new value: .users_names: was null, but now cty.SetValEmpty(cty.String)."
	if !r.UsersNames.IsNull() {
		usersNames, diags := types.SetValueFrom(ctx, types.StringType, group.UsersNames)
		if diags != nil {
			return diags
		}
		r.UsersNames = usersNames
	}

	r.WatchManager = types.BoolValue(group.WatchManager)
	r.PolicyManager = types.BoolValue(group.PolicyManager)
	r.ReportsManager = types.BoolValue(group.ReportsManager)

	return nil
}
