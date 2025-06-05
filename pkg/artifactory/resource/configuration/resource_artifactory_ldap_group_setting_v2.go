package configuration

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
)

const LdapGroupEndpoint = "access/api/v1/ldap/groups/"

func NewLdapGroupSettingResource() resource.Resource {
	return &ArtifactoryLdapGroupSettingResource{
		TypeName: "artifactory_ldap_group_setting_v2",
	}
}

type ArtifactoryLdapGroupSettingResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

// ArtifactoryLdapGroupSettingResourceModel describes the Terraform resource data model to match the
// resource schema.
type ArtifactoryLdapGroupSettingResourceModel struct {
	Id                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	EnabledLdap          types.String `tfsdk:"enabled_ldap"`
	GroupBaseDn          types.String `tfsdk:"group_base_dn"`
	GroupNameAttribute   types.String `tfsdk:"group_name_attribute"`
	GroupMemberAttribute types.String `tfsdk:"group_member_attribute"`
	SubTree              types.Bool   `tfsdk:"sub_tree"`
	ForceAttributeSearch types.Bool   `tfsdk:"force_attribute_search"`
	Filter               types.String `tfsdk:"filter"`
	DescriptionAttribute types.String `tfsdk:"description_attribute"`
	Strategy             types.String `tfsdk:"strategy"`
	RefreshOperation     types.String `tfsdk:"refresh_operation"`
	RefreshUsername      types.String `tfsdk:"refresh_username"`
}

// ArtifactoryLdapGroupSettingResourceAPIModel describes the API data model.
type ArtifactoryLdapGroupSettingResourceAPIModel struct {
	Name                 string `json:"name"`
	EnabledLdap          string `json:"enabled_ldap"`
	GroupBaseDn          string `json:"group_base_dn"`
	GroupNameAttribute   string `json:"group_name_attribute"`
	GroupMemberAttribute string `json:"group_member_attribute"`
	SubTree              bool   `json:"sub_tree"`
	ForceAttributeSearch bool   `json:"force_attribute_search"`
	Filter               string `json:"filter"`
	DescriptionAttribute string `json:"description_attribute"`
	Strategy             string `json:"strategy"`
	RefreshOperation     string `json:"refresh_operation"`
	RefreshUsername      string `json:"refresh_username"`
}

func (r *ArtifactoryLdapGroupSettingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *ArtifactoryLdapGroupSettingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory [ldap group setting resource](https://jfrog.com/help/r/jfrog-rest-apis/ldap-group-setting).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Ldap group setting name.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"enabled_ldap": schema.StringAttribute{
				MarkdownDescription: "The LDAP setting key you want to use for group retrieval.",
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(""),
			},
			"group_base_dn": schema.StringAttribute{
				MarkdownDescription: "A search base for group entry DNs, relative to the DN on the LDAP server’s URL (and not relative to the LDAP Setting’s “Search Base”). Used when importing groups.",
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					ldapDomainNameValidator{},
				},
			},
			"group_name_attribute": schema.StringAttribute{
				MarkdownDescription: "Attribute on the group entry denoting the group name. Used when importing groups.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"group_member_attribute": schema.StringAttribute{
				MarkdownDescription: "A multi-value attribute on the group entry containing user DNs or IDs of the group members (e.g., uniqueMember, member).",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"sub_tree": schema.BoolAttribute{
				MarkdownDescription: "When set, enables deep search through the sub-tree of the LDAP URL + Search Base. `true` by default. `sub_tree` can be set to true only with `STATIC` or `DYNAMIC` strategy.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"force_attribute_search": schema.BoolAttribute{
				MarkdownDescription: "This attribute is used in very specific cases of LDAP group settings. Don't switch it to `false`, unless instructed by the JFrog support team. Default value is `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"filter": schema.StringAttribute{
				MarkdownDescription: "The LDAP filter used to search for group entries. Used for importing groups.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					ldapSearchFilterValidator{},
				},
			},
			"description_attribute": schema.StringAttribute{
				MarkdownDescription: "An attribute on the group entry which denoting the group description. Used when importing groups.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"strategy": schema.StringAttribute{
				MarkdownDescription: "The JFrog Platform Deployment (JPD) supports three ways of mapping groups to LDAP schemas: STATIC: Group objects are aware of their members, however, the users are not aware of the groups they belong to. Each group object such as groupOfNames or groupOfUniqueNames holds its respective member attributes, typically member or uniqueMember, which is a user DN. DYNAMIC: User objects are aware of what groups they belong to, but the group objects are not aware of their members. Each user object contains a custom attribute, such as group, that holds the group DNs or group names of which the user is a member. HIERARCHICAL: The user's DN is indicative of the groups the user belongs to by using group names as part of user DN hierarchy. Each user DN contains a list of ou's or custom attributes that make up the group association. For example, `uid=user1,ou=developers,ou=uk,dc=jfrog,dc=org` indicates that `user1` belongs to two groups: `uk` and `developers`. Valid values are: `STATIC`, `DYNAMIC`, `HIERARCHICAL`, case sensitive, all caps.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("STATIC", "DYNAMIC", "HIERARCHICAL"),
				},
			},
			"refresh_operation": schema.StringAttribute{
				MarkdownDescription: "Operation to perform after updating LDAP group settings. Can be UPDATE, IMPORT, or UPDATE_AND_IMPORT.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("UPDATE_AND_IMPORT"),
				Validators: []validator.String{
					stringvalidator.OneOf("UPDATE", "IMPORT", "UPDATE_AND_IMPORT"),
				},
			},
			"refresh_username": schema.StringAttribute{
				MarkdownDescription: "Optional username to refresh group membership for a specific user.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
		},
	}
}
func refreshLdapGroup(client *resty.Client, groupName, operation, username string) error {
	groupNameEscaped := url.PathEscape(groupName)
	refreshUrl := fmt.Sprintf("%s%s/refresh?operation=%s", LdapGroupEndpoint, groupNameEscaped, operation)
	refreshUrl += "&username=" + url.QueryEscape(username)
	resp, err := client.R().Post(refreshUrl)
	if err != nil {
		return fmt.Errorf("failed to trigger LDAP group refresh: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("LDAP group refresh failed: %s", resp.String())
	}
	return nil
}

func (r *ArtifactoryLdapGroupSettingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *ArtifactoryLdapGroupSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var data *ArtifactoryLdapGroupSettingResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	ldapGroup := &ArtifactoryLdapGroupSettingResourceAPIModel{
		Name:                 data.Name.ValueString(),
		EnabledLdap:          data.EnabledLdap.ValueString(),
		GroupBaseDn:          data.GroupBaseDn.ValueString(),
		GroupNameAttribute:   data.GroupNameAttribute.ValueString(),
		GroupMemberAttribute: data.GroupMemberAttribute.ValueString(),
		SubTree:              data.SubTree.ValueBool(),
		ForceAttributeSearch: data.ForceAttributeSearch.ValueBool(),
		Filter:               data.Filter.ValueString(),
		DescriptionAttribute: data.DescriptionAttribute.ValueString(),
		Strategy:             data.Strategy.ValueString(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(ldapGroup).
		Post(LdapGroupEndpoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusCreated {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Assign the resource ID for the resource in the state
	data.Id = types.StringValue(ldapGroup.Name)

	operation := data.RefreshOperation.ValueString()
	username := data.RefreshUsername.ValueString()
	if err := refreshLdapGroup(r.ProviderData.Client, ldapGroup.Name, operation, username); err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryLdapGroupSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var data *ArtifactoryLdapGroupSettingResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	ldapGroup := ArtifactoryLdapGroupSettingResourceAPIModel{}

	response, err := r.ProviderData.Client.R().
		SetResult(&ldapGroup).
		Get(LdapGroupEndpoint + data.Id.ValueString())

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	if response.StatusCode() == http.StatusBadRequest || response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(data.ToState(ctx, ldapGroup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryLdapGroupSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var data ArtifactoryLdapGroupSettingResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Convert from Terraform data model into API data model
	ldapGroup := ArtifactoryLdapGroupSettingResourceAPIModel{
		Name:                 data.Name.ValueString(),
		EnabledLdap:          data.EnabledLdap.ValueString(),
		GroupBaseDn:          data.GroupBaseDn.ValueString(),
		GroupNameAttribute:   data.GroupNameAttribute.ValueString(),
		GroupMemberAttribute: data.GroupMemberAttribute.ValueString(),
		SubTree:              data.SubTree.ValueBool(),
		ForceAttributeSearch: data.ForceAttributeSearch.ValueBool(),
		Filter:               data.Filter.ValueString(),
		DescriptionAttribute: data.DescriptionAttribute.ValueString(),
		Strategy:             data.Strategy.ValueString(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(ldapGroup).
		Put(LdapGroupEndpoint)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	resp.Diagnostics.Append(data.ToState(ctx, ldapGroup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	operation := data.RefreshOperation.ValueString()
	username := data.RefreshUsername.ValueString()
	if err := refreshLdapGroup(r.ProviderData.Client, ldapGroup.Name, operation, username); err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryLdapGroupSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var data ArtifactoryLdapGroupSettingResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	response, err := r.ProviderData.Client.R().
		Delete(LdapGroupEndpoint + data.Id.ValueString())

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 404 Not Found or 204 No Content
	if response.StatusCode() != http.StatusNotFound && response.StatusCode() != http.StatusNoContent {
		utilfw.UnableToDeleteResourceError(resp, response.String())
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
func (r *ArtifactoryLdapGroupSettingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ArtifactoryLdapGroupSettingResourceModel) ToState(ctx context.Context, ldapGroup ArtifactoryLdapGroupSettingResourceAPIModel) diag.Diagnostics {
	r.Id = types.StringValue(ldapGroup.Name)
	r.Name = types.StringValue(ldapGroup.Name)
	r.EnabledLdap = types.StringValue(ldapGroup.EnabledLdap)
	r.GroupBaseDn = types.StringValue(ldapGroup.GroupBaseDn)
	r.GroupNameAttribute = types.StringValue(ldapGroup.GroupNameAttribute)
	r.GroupMemberAttribute = types.StringValue(ldapGroup.GroupMemberAttribute)
	r.SubTree = types.BoolValue(ldapGroup.SubTree)
	r.ForceAttributeSearch = types.BoolValue(ldapGroup.ForceAttributeSearch)
	r.Filter = types.StringValue(ldapGroup.Filter)
	r.DescriptionAttribute = types.StringValue(ldapGroup.DescriptionAttribute)
	r.Strategy = types.StringValue(ldapGroup.Strategy)

	return nil
}

func (r *ArtifactoryLdapGroupSettingResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ArtifactoryLdapGroupSettingResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Strategy.IsNull() || data.Strategy.IsUnknown() {
		return
	}

	if data.SubTree.IsNull() || data.SubTree.IsUnknown() {
		return
	}

	// Validate strategy and sub_tree
	if strings.ToUpper(data.Strategy.ValueString()) == "HIERARCHICAL" && data.SubTree.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("sub_tree"),
			"Incorrect Attribute Configuration",
			"sub_tree can be set to true only with `STATIC` or `DYNAMIC` strategy",
		)
	}
}
