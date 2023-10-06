package security

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func NewScopedTokenResource() resource.Resource {
	return &ScopedTokenResource{}
}

type ScopedTokenResource struct {
	ProviderData utilsdk.ProvderMetadata
}

// ScopedTokenResourceModel describes the Terraform resource data model to match the
// resource schema.
type ScopedTokenResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	GrantType             types.String `tfsdk:"grant_type"`
	Username              types.String `tfsdk:"username"`
	ProjectKey            types.String `tfsdk:"project_key"`
	Scopes                types.Set    `tfsdk:"scopes"`
	ExpiresIn             types.Int64  `tfsdk:"expires_in"`
	Refreshable           types.Bool   `tfsdk:"refreshable"`
	IncludeReferenceToken types.Bool   `tfsdk:"include_reference_token"`
	Description           types.String `tfsdk:"description"`
	Audiences             types.Set    `tfsdk:"audiences"`
	AccessToken           types.String `tfsdk:"access_token"`
	RefreshToken          types.String `tfsdk:"refresh_token"`
	ReferenceToken        types.String `tfsdk:"reference_token"`
	TokenType             types.String `tfsdk:"token_type"`
	Subject               types.String `tfsdk:"subject"`
	Expiry                types.Int64  `tfsdk:"expiry"`
	IssuedAt              types.Int64  `tfsdk:"issued_at"`
	Issuer                types.String `tfsdk:"issuer"`
}

type AccessTokenPostResponseAPIModel struct {
	TokenId        string `json:"token_id"`
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	ExpiresIn      int64  `json:"expires_in"`
	Scope          string `json:"scope"`
	TokenType      string `json:"token_type"`
	ReferenceToken string `json:"reference_token"`
}

type AccessTokenErrorResponseAPIModel struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

type AccessTokenPostRequestAPIModel struct {
	GrantType             string `json:"grant_type"`
	Username              string `json:"username,omitempty"`
	ProjectKey            string `json:"project_key"`
	Scope                 string `json:"scope,omitempty"`
	ExpiresIn             int64  `json:"expires_in"`
	Refreshable           bool   `json:"refreshable"`
	Description           string `json:"description,omitempty"`
	Audience              string `json:"audience,omitempty"`
	IncludeReferenceToken bool   `json:"include_reference_token"`
}

type AccessTokenGetAPIModel struct {
	TokenId     string `json:"token_id"`
	Subject     string `json:"subject"`
	Expiry      int64  `json:"expiry"`
	IssuedAt    int64  `json:"issued_at"`
	Issuer      string `json:"issuer"`
	Description string `json:"description"`
	Refreshable bool   `json:"refreshable"`
}

func (r *ScopedTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_scoped_token"
}

func (r *ScopedTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Create scoped tokens for any of the services in your JFrog Platform and to " +
			"manage user access to these services. If left at the default setting, the token will " +
			"be created with the user-identity scope, which allows users to identify themselves in " +
			"the Platform but does not grant any specific access permissions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"grant_type": schema.StringAttribute{
				MarkdownDescription: "The grant type used to authenticate the request. In this case, the only value supported is `client_credentials` which is also the default value if this parameter is not specified.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("client_credentials"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The user name for which this token is created. The username is based " +
					"on the authenticated user - either from the user of the authenticated token or based " +
					"on the username (if basic auth was used). The username is then used to set the subject " +
					"of the token: <service-id>/users/<username>. Limited to 255 characters.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			"project_key": schema.StringAttribute{
				MarkdownDescription: "The project for which this token is created. Enter the project name on which you want to apply this token.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^^[a-z][a-z0-9\-]{1,31}$`),
						"must be 2 - 32 lowercase alphanumeric and hyphen characters",
					),
				},
			},
			"scopes": schema.SetAttribute{
				MarkdownDescription: "The scope of access that the token provides. Access to the REST API is always " +
					"provided by default. Administrators can set any scope, while non-admin users can only set " +
					"the scope to a subset of the groups to which they belong.\n" +
					"The supported scopes include:\n" +
					"* `applied-permissions/user` - provides user access. If left at the default setting, the " +
					"token will be created with the user-identity scope, which allows users to identify themselves " +
					"in the Platform but does not grant any specific access permissions." +
					"* `applied-permissions/admin` - the scope assigned to admin users." +
					"* `applied-permissions/groups` - the group to which permissions are assigned by group name " +
					"(use username to inicate the group name)" +
					"* `system:metrics:r` - for getting the service metrics" +
					"* `system:livelogs:r` - for getting the service livelogsr" +
					"The scope to assign to the token should be provided as a list of scope tokens, limited to 500 characters in total.\n" +
					"Resource Permissions\n" +
					"From Artifactory 7.38.x, resource permissions scoped tokens are also supported in the REST API. " +
					"A permission can be represented as a scope token string in the following format:\n" +
					"`<resource-type>:<target>[/<sub-resource>]:<actions>`\n" +
					"Where:\n" +
					" `<resource-type>` - one of the permission resource types, from a predefined closed list. " +
					"Currently, the only resource type that is supported is the artifact resource type.\n" +
					" `<target>` - the target resource, can be exact name or a pattern" +
					" `<sub-resource>` - optional, the target sub-resource, can be exact name or a pattern" +
					" `<actions>` - comma-separated list of action acronyms." +
					"The actions allowed are <r, w, d, a, m> or any combination of these actions\n." +
					"To allow all actions - use `*`\n" +
					"Examples: " +
					" `[\"applied-permissions/user\", \"artifact:generic-local:r\"]`\n" +
					" `[\"applied-permissions/group\", \"artifact:generic-local/path:*\"]`\n" +
					" `[\"applied-permissions/admin\", \"system:metrics:r\", \"artifact:generic-local:*\"]`",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplaceIfConfigured(),
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.Any(
						stringvalidator.OneOf(
							"applied-permissions/user",
							"applied-permissions/admin",
							"system:metrics:r",
							"system:livelogs:r",
						),
						stringvalidator.RegexMatches(regexp.MustCompile(`^applied-permissions/groups:.+$`), "must be 'applied-permissions/groups:<group-name>[,<group-name>...]'"),
						stringvalidator.RegexMatches(regexp.MustCompile(`^artifact:.+:([rwdam*]|([rwdam]+(,[rwdam]+)))$`), "must be '<resource-type>:<target>[/<sub-resource>]:<actions>'"),
					),
					),
				},
			},
			"expires_in": schema.Int64Attribute{
				MarkdownDescription: "The amount of time, in seconds, it would take for the token to expire. An admin shall be able to set whether expiry is mandatory, what is the default expiry, and what is the maximum expiry allowed. Must be non-negative. Default value is based on configuration in 'access.config.yaml'. See [API documentation](https://jfrog.com/help/r/jfrog-rest-apis/revoke-token-by-id) for details. Access Token would not be saved by Artifactory if this is less than the persistence threshold value (default to 10800 seconds) set in Access configuration. See [official documentation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/using-the-revocable-and-persistency-thresholds) for details.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{int64validator.AtLeast(0)},
			},
			"refreshable": schema.BoolAttribute{
				MarkdownDescription: "Is this token refreshable? Default is `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"include_reference_token": schema.BoolAttribute{
				MarkdownDescription: "Also create a reference token which can be used like an API key. Default is `false`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Free text token description. Useful for filtering and managing tokens. Limited to 1024 characters.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{stringvalidator.LengthBetween(0, 1024)},
			},
			"audiences": schema.SetAttribute{
				MarkdownDescription: "A list of the other instances or services that should accept this " +
					"token identified by their Service-IDs. Limited to total 255 characters. " +
					"Default to '*@*' if not set. Service ID must begin with valid JFrog service type. " +
					"Options: jfrt, jfxr, jfpip, jfds, jfmc, jfac, jfevt, jfmd, jfcon, or *. For instructions to retrieve the Artifactory Service ID see this [documentation](https://jfrog.com/help/r/jfrog-rest-apis/get-service-id)",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplaceIfConfigured(),
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.All(
						stringvalidator.LengthAtLeast(1),
						stringvalidator.RegexMatches(regexp.MustCompile(fmt.Sprintf(`^(%s|\*)@.+`, strings.Join(serviceTypesScopedToken, "|"))),
							fmt.Sprintf(
								"must either begin with %s, or *",
								strings.Join(serviceTypesScopedToken, ", "),
							),
						),
					),
					),
				},
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Returns the access token to authenticate to Artifactory.",
				Sensitive:           true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"refresh_token": schema.StringAttribute{
				MarkdownDescription: "Refresh token.",
				Sensitive:           true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"reference_token": schema.StringAttribute{
				MarkdownDescription: "Reference Token (alias to Access Token).",
				Sensitive:           true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"token_type": schema.StringAttribute{
				MarkdownDescription: "Returns the token type.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "Returns the token type.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"expiry": schema.Int64Attribute{
				MarkdownDescription: "Returns the token expiry.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"issued_at": schema.Int64Attribute{
				MarkdownDescription: "Returns the token issued at date/time.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"issuer": schema.StringAttribute{
				MarkdownDescription: "Returns the token issuer.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

var serviceTypesScopedToken = []string{"jfrt", "jfxr", "jfpip", "jfds", "jfmc", "jfac", "jfevt", "jfmd", "jfcon"}

func (r *ScopedTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *ScopedTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ScopedTokenResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopes := []string{}
	if !data.Scopes.IsNull() {
		scopes = utilfw.StringSetToStrings(data.Scopes)
	}
	scopesString := strings.Join(scopes, " ") // Join slice into space-separated string
	if len(scopesString) > 500 {
		resp.Diagnostics.AddError(
			"Scopes length exceeds 500 characters",
			"total combined length of scopes field exceeds 500 characters:"+scopesString,
		)
		return
	}

	audiences := []string{}
	if !data.Audiences.IsNull() {
		audiences = utilfw.StringSetToStrings(data.Audiences)
	}
	audiencesString := strings.Join(audiences, " ") // Join slice into space-separated string
	if len(audiencesString) > 255 {
		resp.Diagnostics.AddError(
			"Audiences length exceeds 255 characters",
			"total combined length of audiences field exceeds 255 characters:"+audiencesString,
		)
		return
	}

	// Convert from Terraform data model into API data model
	accessTokenPostBody := AccessTokenPostRequestAPIModel{
		GrantType:             data.GrantType.ValueString(),
		Username:              data.Username.ValueString(),
		ProjectKey:            data.ProjectKey.ValueString(),
		Scope:                 scopesString,
		ExpiresIn:             data.ExpiresIn.ValueInt64(),
		Refreshable:           data.Refreshable.ValueBool(),
		Description:           data.Description.ValueString(),
		Audience:              audiencesString,
		IncludeReferenceToken: data.IncludeReferenceToken.ValueBool(),
	}

	postResult := AccessTokenPostResponseAPIModel{}

	response, err := r.ProviderData.Client.R().
		SetBody(accessTokenPostBody).
		SetResult(&postResult).
		Post("access/api/v1/tokens")

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	getResult := AccessTokenGetAPIModel{}
	id := types.StringValue(postResult.TokenId)

	_, err = r.ProviderData.Client.R().
		SetPathParam("id", id.ValueString()).
		SetResult(&getResult).
		Get("access/api/v1/tokens/{id}")

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Assign the attribute values for the resource in the state
	resp.Diagnostics.Append(data.PostResponseToState(ctx, &postResult, &accessTokenPostBody, &getResult)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // All attributes are assigned in data
}

func (r *ScopedTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ScopedTokenResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var accessToken AccessTokenGetAPIModel

	response, err := r.ProviderData.Client.R().
		SetPathParam("id", data.Id.ValueString()).
		SetResult(&accessToken).
		Get("access/api/v1/tokens/{id}")

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		if response.StatusCode() == http.StatusNotFound {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Scoped token %s not found or not created", data.Id.ValueString()),
				"Access Token would not be saved by Artifactory if 'expires_in' is less than the persistence threshold value (default to 10800 seconds) set in Access configuration. See https://www.jfrog.com/confluence/display/JFROG/Access+Tokens#AccessTokens-PersistencyThreshold for details."+err.Error(),
			)
			resp.State.RemoveResource(ctx)
			return
		}
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	data.GetResponseToState(ctx, &accessToken)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScopedTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Scoped tokens are not updatable
}

func (r *ScopedTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ScopedTokenResourceModel
	respError := AccessTokenErrorResponseAPIModel{}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	id := data.Id.ValueString()

	_, err := r.ProviderData.Client.R().
		SetPathParam("id", id).
		SetError(&respError).
		Delete("access/api/v1/tokens/{id}")

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to revoke scoped token %s", id),
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}
	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ScopedTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError(
		"Import is not supported",
		"resource artifactory_scoped_token doesn't support import.",
	)
}

func (r *ScopedTokenResourceModel) PostResponseToState(ctx context.Context,
	accessTokenResp *AccessTokenPostResponseAPIModel, accessTokenPostBody *AccessTokenPostRequestAPIModel, getResult *AccessTokenGetAPIModel) diag.Diagnostics {

	r.Id = types.StringValue(accessTokenResp.TokenId)

	if len(accessTokenResp.Scope) > 0 {
		scopesList := strings.Split(accessTokenResp.Scope, " ")
		scopes, diags := types.SetValueFrom(ctx, types.StringType, scopesList)
		if diags != nil {
			return diags
		}
		r.Scopes = scopes
	}

	r.ExpiresIn = types.Int64Value(accessTokenResp.ExpiresIn)

	r.AccessToken = types.StringValue(accessTokenResp.AccessToken)

	// only have refresh token if 'refreshable' is set to true in the request
	r.RefreshToken = types.StringNull()
	if accessTokenPostBody.Refreshable && len(accessTokenResp.RefreshToken) > 0 {
		r.RefreshToken = types.StringValue(accessTokenResp.RefreshToken)
	}

	// only have reference token if 'include_reference_token' is set to true in the request
	r.ReferenceToken = types.StringNull()
	if accessTokenPostBody.IncludeReferenceToken && len(accessTokenResp.ReferenceToken) > 0 {
		r.ReferenceToken = types.StringValue(accessTokenResp.ReferenceToken)
	}

	r.IncludeReferenceToken = types.BoolValue(accessTokenPostBody.IncludeReferenceToken)
	r.TokenType = types.StringValue(accessTokenResp.TokenType)
	r.Subject = types.StringValue(getResult.Subject)
	r.Expiry = types.Int64Value(getResult.Expiry) // could be absent in the get response!
	r.IssuedAt = types.Int64Value(getResult.IssuedAt)
	r.Issuer = types.StringValue(getResult.Issuer)

	return nil
}

func (r *ScopedTokenResourceModel) GetResponseToState(ctx context.Context, accessToken *AccessTokenGetAPIModel) {
	r.Id = types.StringValue(accessToken.TokenId)
	if r.GrantType.IsNull() {
		r.GrantType = types.StringValue("client_credentials")
	}
	r.Subject = types.StringValue(accessToken.Subject)
	r.Expiry = types.Int64Value(accessToken.Expiry)
	r.IssuedAt = types.Int64Value(accessToken.IssuedAt)
	r.Issuer = types.StringValue(accessToken.Issuer)

	if r.Description.IsNull() {
		r.Description = types.StringValue("")
	}
	if len(accessToken.Description) > 0 {
		r.Description = types.StringValue(accessToken.Description)
	}

	r.Refreshable = types.BoolValue(accessToken.Refreshable)

	// Need to set empty string for null state value to avoid state drift.
	// See https://discuss.hashicorp.com/t/diffsuppressfunc-alternative-in-terraform-framework/52578/2
	if r.RefreshToken.IsNull() {
		r.RefreshToken = types.StringValue("")
	}
	if r.ReferenceToken.IsNull() {
		r.ReferenceToken = types.StringValue("")
	}
}

func CheckAccessToken(id string, request *resty.Request) (*resty.Response, error) {
	return request.SetPathParam("id", id).Get("access/api/v1/tokens/{id}")
}
