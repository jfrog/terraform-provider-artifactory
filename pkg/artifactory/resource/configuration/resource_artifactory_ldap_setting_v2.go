package configuration

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
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
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"gopkg.in/ldap.v2"
)

const LdapEndpoint = "access/api/v1/ldap/settings/"

func NewLdapSettingResource() resource.Resource {
	return &ArtifactoryLdapSettingResource{}
}

type ArtifactoryLdapSettingResource struct {
	ProviderData utilsdk.ProvderMetadata
}

// ArtifactoryLdapSettingResourceModel describes the Terraform resource data model to match the
// resource schema.
type ArtifactoryLdapSettingResourceModel struct {
	Id                       types.String `tfsdk:"id"`
	Key                      types.String `tfsdk:"key"`
	Enabled                  types.Bool   `tfsdk:"enabled"`
	LdapUrl                  types.String `tfsdk:"ldap_url"`
	UserDnPattern            types.String `tfsdk:"user_dn_pattern"`
	EmailAttribute           types.String `tfsdk:"email_attribute"`
	AutoCreateUser           types.Bool   `tfsdk:"auto_create_user"`
	LdapPoisoningProtection  types.Bool   `tfsdk:"ldap_poisoning_protection"`
	AllowUserToAccessProfile types.Bool   `tfsdk:"allow_user_to_access_profile"`
	PagingSupportEnabled     types.Bool   `tfsdk:"paging_support_enabled"`
	SearchFilter             types.String `tfsdk:"search_filter"`
	SearchBase               types.String `tfsdk:"search_base"`
	SearchSubTree            types.Bool   `tfsdk:"search_sub_tree"`
	ManagerDn                types.String `tfsdk:"manager_dn"`
	ManagerPassword          types.String `tfsdk:"manager_password"`
}

// ArtifactoryLdapSettingResourceAPIModel describes the API data model.
type ArtifactoryLdapSettingResourceAPIModel struct {
	Key                      string             `json:"key"`
	Enabled                  bool               `json:"enabled"`
	LdapUrl                  string             `json:"ldap_url"`
	UserDnPattern            string             `json:"user_dn_pattern"`
	Search                   LdapSearchAPIModel `json:"search"`
	AutoCreateUser           bool               `json:"auto_create_user"`
	EmailAttribute           string             `json:"email_attribute"`
	LdapPoisoningProtection  bool               `json:"ldap_poisoning_protection"`
	AllowUserToAccessProfile bool               `json:"allow_user_to_access_profile"`
	PagingSupportEnabled     bool               `json:"paging_support_enabled"`
}

type LdapSearchAPIModel struct {
	SearchFilter    string `json:"search_filter,omitempty"`
	SearchBase      string `json:"search_base,omitempty"`
	SearchSubTree   bool   `json:"search_sub_tree"`
	ManagerDn       string `json:"manager_dn,omitempty"`
	ManagerPassword string `json:"manager_password,omitempty"`
}

func (r *ArtifactoryLdapSettingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_ldap_setting_v2"
}

func (r *ArtifactoryLdapSettingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory [ldap setting resource](https://jfrog.com/help/r/jfrog-rest-apis/ldap-setting).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Ldap setting name.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Flag to enable or disable the ldap setting. Default value is `true`.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			"ldap_url": schema.StringAttribute{
				MarkdownDescription: "Location of the LDAP server in the following format: `ldap://myldapserver/dc=sampledomain,dc=com`",
				Required:            true,
				Validators: []validator.String{stringvalidator.RegexMatches(
					regexp.MustCompile(`^ldap.*`), "must start with `ldap`"),
				},
			},
			"user_dn_pattern": schema.StringAttribute{
				MarkdownDescription: "A DN pattern that can be used to log users directly in to LDAP. This pattern is used to create a DN string for 'direct' user authentication where the pattern is relative to the base DN in the LDAP URL. The pattern argument {0} is replaced with the username. This only works if anonymous binding is allowed and a direct user DN can be used, which is not the default case for Active Directory (use User DN search filter instead). Example: uid={0},ou=People. Default value is blank/empty.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"auto_create_user": schema.BoolAttribute{
				MarkdownDescription: "When set, users are automatically created when using LDAP. Otherwise, users are transient and associated with auto-join groups defined in Artifactory. Default value is `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"email_attribute": schema.StringAttribute{
				MarkdownDescription: "An attribute that can be used to map a user's email address to a user created automatically in Artifactory. Default value is`mail`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("mail"),
			},
			"ldap_poisoning_protection": schema.BoolAttribute{
				MarkdownDescription: "When this is set to `true`, an empty or missing usernames array will detach all users from the group.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"allow_user_to_access_profile": schema.BoolAttribute{
				MarkdownDescription: "Auto created users will have access to their profile page and will be able to perform actions such as generating an API key. Default value is `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"paging_support_enabled": schema.BoolAttribute{
				MarkdownDescription: "When set, supports paging results for the LDAP server. This feature requires that the LDAP server supports a PagedResultsControl configuration. Default value is `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			// Search attributes must be set together, otherwise LDAP settings will be broken on the Artifactory instance.
			// Only `search_sub_tree` bool is always set by default in the UI call, which is not creating a problem.
			"search_filter": schema.StringAttribute{
				MarkdownDescription: "A filter expression used to search for the user DN used in LDAP authentication. This is an LDAP search filter (as defined in 'RFC 2254') with optional arguments. In this case, the username is the only argument, and is denoted by '{0}'. Possible examples are: (uid={0}) - This searches for a username match on the attribute. Authentication to LDAP is performed from the DN found if successful.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("search_base"),
						path.MatchRoot("search_sub_tree"),
						path.MatchRoot("manager_dn"),
						path.MatchRoot("manager_password"),
					}...),
				},
			},
			"search_base": schema.StringAttribute{
				MarkdownDescription: "A context name to search in relative to the base DN of the LDAP URL. For example, 'ou=users' With the LDAP Group Add-on enabled, it is possible to enter multiple search base entries separated by a pipe ('|') character.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("search_filter"),
						path.MatchRoot("search_sub_tree"),
						path.MatchRoot("manager_dn"),
						path.MatchRoot("manager_password"),
					}...),
				},
			},
			"search_sub_tree": schema.BoolAttribute{
				MarkdownDescription: "When set, enables deep search through the sub tree of the LDAP URL + search base. Default value is `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				Validators: []validator.Bool{
					boolvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("search_filter"),
						path.MatchRoot("search_base"),
						path.MatchRoot("manager_dn"),
						path.MatchRoot("manager_password"),
					}...),
				},
			},
			"manager_dn": schema.StringAttribute{
				MarkdownDescription: "The full DN of the user that binds to the LDAP server to perform user searches. Only used with `search` authentication.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("search_filter"),
						path.MatchRoot("search_base"),
						path.MatchRoot("search_sub_tree"),
						path.MatchRoot("manager_password"),
					}...),
				},
			},
			"manager_password": schema.StringAttribute{
				MarkdownDescription: "The password of the user that binds to the LDAP server to perform the search. Only used with `search` authentication.",
				Optional:            true,
				Sensitive:           true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("search_filter"),
						path.MatchRoot("search_base"),
						path.MatchRoot("search_sub_tree"),
						path.MatchRoot("manager_dn"),
					}...),
				},
			},
		},
	}
}

func (r *ArtifactoryLdapSettingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *ArtifactoryLdapSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ArtifactoryLdapSettingResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	ldapSearch := LdapSearchAPIModel{
		SearchFilter:  data.SearchFilter.ValueString(),
		SearchBase:    data.SearchBase.ValueString(),
		SearchSubTree: data.SearchSubTree.ValueBool(),
		ManagerDn:     data.ManagerDn.ValueString(),
	}
	if !data.ManagerPassword.IsNull() {
		ldapSearch.ManagerPassword = data.ManagerPassword.ValueString()
	}
	ldap := ArtifactoryLdapSettingResourceAPIModel{
		Key:                      data.Key.ValueString(),
		Enabled:                  data.Enabled.ValueBool(),
		LdapUrl:                  data.LdapUrl.ValueString(),
		UserDnPattern:            data.UserDnPattern.ValueString(),
		Search:                   ldapSearch,
		AutoCreateUser:           data.AutoCreateUser.ValueBool(),
		EmailAttribute:           data.EmailAttribute.ValueString(),
		LdapPoisoningProtection:  data.LdapPoisoningProtection.ValueBool(),
		AllowUserToAccessProfile: data.AllowUserToAccessProfile.ValueBool(),
		PagingSupportEnabled:     data.PagingSupportEnabled.ValueBool(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(ldap).
		Post(LdapEndpoint)

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
	data.Id = types.StringValue(ldap.Key)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryLdapSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ArtifactoryLdapSettingResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	ldap := ArtifactoryLdapSettingResourceAPIModel{}

	response, err := r.ProviderData.Client.R().
		SetResult(&ldap).
		Get(LdapEndpoint + data.Id.ValueString())

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		if response.StatusCode() == http.StatusBadRequest || response.StatusCode() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(data.ToState(ctx, ldap)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryLdapSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ArtifactoryLdapSettingResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Convert from Terraform data model into API data model
	ldapSearch := LdapSearchAPIModel{
		SearchFilter:  data.SearchFilter.ValueString(),
		SearchBase:    data.SearchBase.ValueString(),
		SearchSubTree: data.SearchSubTree.ValueBool(),
		ManagerDn:     data.ManagerDn.ValueString(),
	}
	if !data.ManagerPassword.IsNull() {
		ldapSearch.ManagerPassword = data.ManagerPassword.ValueString()
	}
	ldap := ArtifactoryLdapSettingResourceAPIModel{
		Key:                      data.Key.ValueString(),
		Enabled:                  data.Enabled.ValueBool(),
		LdapUrl:                  data.LdapUrl.ValueString(),
		UserDnPattern:            data.UserDnPattern.ValueString(),
		Search:                   ldapSearch,
		AutoCreateUser:           data.AutoCreateUser.ValueBool(),
		EmailAttribute:           data.EmailAttribute.ValueString(),
		LdapPoisoningProtection:  data.LdapPoisoningProtection.ValueBool(),
		AllowUserToAccessProfile: data.AllowUserToAccessProfile.ValueBool(),
		PagingSupportEnabled:     data.PagingSupportEnabled.ValueBool(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(ldap).
		Put(LdapEndpoint)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	resp.Diagnostics.Append(data.ToState(ctx, ldap)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ArtifactoryLdapSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ArtifactoryLdapSettingResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	response, err := r.ProviderData.Client.R().
		Delete(LdapEndpoint + data.Id.ValueString())

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 404 Not Found or 204 No Content
	if response.StatusCode() != http.StatusNotFound && response.StatusCode() != http.StatusNoContent {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ArtifactoryLdapSettingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ArtifactoryLdapSettingResourceModel) ToState(ctx context.Context, ldap ArtifactoryLdapSettingResourceAPIModel) diag.Diagnostics {
	r.Id = types.StringValue(ldap.Key)
	r.Key = types.StringValue(ldap.Key)
	r.Enabled = types.BoolValue(ldap.Enabled)
	r.LdapUrl = types.StringValue(ldap.LdapUrl)
	r.UserDnPattern = types.StringValue(ldap.UserDnPattern)
	r.EmailAttribute = types.StringValue(ldap.EmailAttribute)
	r.AutoCreateUser = types.BoolValue(ldap.AutoCreateUser)
	r.LdapPoisoningProtection = types.BoolValue(ldap.LdapPoisoningProtection)
	r.AllowUserToAccessProfile = types.BoolValue(ldap.AllowUserToAccessProfile)
	r.PagingSupportEnabled = types.BoolValue(ldap.PagingSupportEnabled)
	r.SearchSubTree = types.BoolValue(ldap.Search.SearchSubTree)

	if r.SearchFilter.IsNull() {
		r.SearchFilter = types.StringValue("")
	} else {
		r.SearchFilter = types.StringValue(ldap.Search.SearchFilter)
	}
	if ldap.Search.SearchFilter != "" {
		r.SearchFilter = types.StringValue(ldap.Search.SearchFilter)
	}

	if r.SearchBase.IsNull() {
		r.SearchBase = types.StringValue("")
	} else {
		r.SearchBase = types.StringValue(ldap.Search.SearchBase)
	}
	if ldap.Search.SearchBase != "" {
		r.SearchBase = types.StringValue(ldap.Search.SearchBase)
	}

	if r.ManagerDn.IsNull() {
		r.ManagerDn = types.StringValue("")
	}
	if ldap.Search.ManagerDn != "" {
		r.ManagerDn = types.StringValue(ldap.Search.ManagerDn)
	} else {
		r.ManagerDn = types.StringValue(ldap.Search.ManagerDn)
	}

	if r.ManagerPassword.IsNull() {
		r.ManagerPassword = types.StringValue("")
	}

	return nil
}

func (r *ArtifactoryLdapSettingResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ArtifactoryLdapSettingResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Validate search_filter
	if !data.SearchFilter.IsNull() {
		_, err := ldap.CompileFilter(data.SearchFilter.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("search_filter"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("Expected search_filter to be a valid LDAP search filter, %v", err),
			)
		}
	}
	// Validate user_dn_pattern
	if !data.UserDnPattern.IsNull() {
		_, err := ldap.ParseDN(data.UserDnPattern.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("user_dn_pattern"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("Expected user_dn_pattern to be a valid LDAP Domain Name, %v", err),
			)
		}
	}
	// Validate search_base
	if !data.SearchBase.IsNull() {
		_, err := ldap.ParseDN(data.SearchBase.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("search_base"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("Expected search_base to be a valid LDAP Domain Name, %v", err),
			)
		}
	}
	// Validate managed_dn
	if !data.ManagerDn.IsNull() {
		_, err := ldap.ParseDN(data.ManagerDn.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("manager_dn"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("Expected manager_dn to be a valid LDAP Domain Name, %v", err),
			)
		}
	}
}
