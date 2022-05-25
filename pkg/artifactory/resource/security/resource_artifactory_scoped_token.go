package security

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

type AccessTokenPostRequest struct {
	GrantType   string `json:"grant_type"`
	Username    string `json:"username,omitempty"`
	Scope       string `json:"scope,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	Refreshable bool   `json:"refreshable"`
	Description string `json:"description"`
	Audience    string `json:"audience,omitempty"`
}

type AccessTokenPostResponse struct {
	TokenId     string `json:"token_id"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type AccessTokenGet struct {
	TokenId     string `json:"token_id"`
	Subject     string `json:"subject"`
	Expiry      int    `json:"expiry"`
	IssuedAt    int    `json:"issued_at"`
	Issuer      string `json:"issuer"`
	Description string `json:"description"`
	Refreshable bool   `json:"refreshable"`
}

func (a AccessTokenPostResponse) Id() string {
	return a.TokenId
}

const accessTokensUrl = "access/api/v1/tokens"
const accessTokenUrl = accessTokensUrl + "/{id}"

func ResourceArtifactoryScopedToken() *schema.Resource {

	var unpackAccessTokenPostRequest = func(data *schema.ResourceData) (*AccessTokenPostRequest, error) {
		d := &util.ResourceData{data}

		scopes := d.GetSet("scopes")
		scopesString := strings.Join(scopes, " ") // Join slice into space-separated string
		if len(scopesString) > 500 {
			return nil, fmt.Errorf("Total combined length of scopes field exceeds 500 characters: %s", scopesString)
		}

		audiences := d.GetSet("audiences")
		audiencesString := strings.Join(audiences, " ") // Join slice into space-separated string
		if len(audiencesString) > 255 {
			return nil, fmt.Errorf("Total combined length of audences field exceeds 255 characters: %s", audiencesString)
		}

		accessToken := AccessTokenPostRequest{
			Username:    d.GetString("username", false),
			Scope:       scopesString,
			ExpiresIn:   d.GetInt("expires_in", false),
			Refreshable: d.GetBool("refreshable", false),
			Description: d.GetString("description", false),
			Audience:    audiencesString,
		}

		return &accessToken, nil
	}

	var packAccessToken = func(d *schema.ResourceData, accessToken AccessTokenGet) diag.Diagnostics {
		setValue := util.MkLens(d)

		var errors []error

		errors = append(errors, setValue("subject", accessToken.Subject)...)
		errors = append(errors, setValue("expiry", accessToken.Expiry)...)
		errors = append(errors, setValue("issued_at", accessToken.IssuedAt)...)
		errors = append(errors, setValue("issuer", accessToken.Issuer)...)
		errors = append(errors, setValue("description", accessToken.Description)...)
		errors = append(errors, setValue("refreshable", accessToken.Refreshable)...)

		if len(errors) > 0 {
			return diag.Errorf("failed to pack access token %q", errors)
		}

		return nil
	}

	var accessTokenRead = func(_ context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		accessToken := AccessTokenGet{}

		_, err := m.(*resty.Client).R().
			SetPathParam("id", data.Id()).
			SetResult(&accessToken).
			Get(accessTokenUrl)

		if err != nil {
			return diag.FromErr(err)
		}

		return packAccessToken(data, accessToken)
	}

	var packAccessTokenPostResponse = func(d *schema.ResourceData, accessToken AccessTokenPostResponse) diag.Diagnostics {
		setValue := util.MkLens(d)

		var errors []error
		errors = append(errors, setValue("scopes", strings.Split(accessToken.Scope, " "))...)
		errors = append(errors, setValue("expires_in", accessToken.ExpiresIn)...)
		errors = append(errors, setValue("access_token", accessToken.AccessToken)...)
		errors = append(errors, setValue("token_type", accessToken.TokenType)...)

		if len(errors) > 0 {
			return diag.Errorf("failed to pack access token from POST response %q", errors)
		}

		return nil
	}

	var accessTokenCreate = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		accessToken, err := unpackAccessTokenPostRequest(data)
		if err != nil {
			return diag.FromErr(err)
		}

		accessToken.GrantType = "client_credentials"

		result := AccessTokenPostResponse{}
		_, err = m.(*resty.Client).R().
			SetBody(accessToken).
			SetResult(&result).
			Post(accessTokensUrl)
		if err != nil {
			return diag.FromErr(err)
		}

		data.SetId(result.Id())

		diags := packAccessTokenPostResponse(data, result)
		if diags != nil {
			return diags
		}

		return accessTokenRead(ctx, data, m)
	}

	var accessTokenDelete = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {

		_, err := m.(*resty.Client).R().
			SetPathParam("id", data.Id()).
			Delete(accessTokenUrl)

		if err != nil {
			d := &util.ResourceData{data}

			expiry := time.Unix(int64(d.GetInt("expiry", false)), 0)
			issuedAt := time.Unix(int64(d.GetInt("issued_at", false)), 0)

			return diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to revoke scoped token %s", data.Id()),
				Detail:   fmt.Sprintf("Token expiration time: %s, issued at: %s", expiry, issuedAt),
			}}
		}

		data.SetId("")

		return nil
	}

	return &schema.Resource{
		CreateContext: accessTokenCreate,
		ReadContext:   accessTokenRead,
		DeleteContext: accessTokenDelete,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 255)),
				Description:      "The user name for which this token is created. The username is based on the authenticated user - either from the user of the authenticated token or based on the username (if basic auth was used). The username is then used to set the subject of the token: <service-id>/users/<username>. Limited to 255 characters.",
			},
			"scopes": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The scope of access that the token provides. Access to the REST API is always provided by default. Administrators can set any scope, while non-admin users can only set the scope to a subset of the groups to which they belong.\n" +
					"The supported scopes include:\n" +
					"* applied-permissions/user - provides user access. If left at the default setting, the token will be created with the user-identity scope, which allows users to identify themselves in the Platform but does not grant any specific access permissions." +
					"* applied-permissions/admin - the scope assigned to admin users." +
					"* applied-permissions/groups - the group to which permissions are assigned by group name (use username to inicate the group name)" +
					"* system:metrics:r - for getting the service metrics" +
					"* system:livelogs:r - for getting the service livelogsr" +
					"The scope to assign to the token should be provided as a list of scope tokens, limited to 500 characters in total.\n" +
					"Resource Permissions\n" +
					"From Artifactory 7.38.x, resource permissions scoped tokens are also supported in the REST API. A permission can be represented as a scope token string in the following format:\n" +
					"`<resource-type>:<target>[/<sub-resource>]:<actions>`\n" +
					"Where:\n" +
					"* `<resource-type>` - one of the permission resource types, from a predefined closed list. Currently, the only resource type that is supported is the artifact resource type.\n" +
					"* `<target>` - the target resource, can be exact name or a pattern" +
					"* `<sub-resource>` - optional, the target sub-resource, can be exact name or a pattern" +
					"* `<actions>` - comma-separated list of action acronyms." +
					"The actions allowed are <r, w, d, a, m> or any combination of these actions\n." +
					"To allow all actions - use `*`",
			},
			"expires_in": {
				Type:             schema.TypeInt,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: validator.IntAtLeast(0),
				Description:      "The amount of time, in seconds, it would take for the token to expire. An admin shall be able to set whether expiry is mandatory, what is the default expiry, and what is the maximum expiry allowed. Must be non-negative. Default value is based on configuration in 'access.config.yaml'. See [API documentation](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-RevokeTokenbyIDrevoketokenbyid) for details.",
			},
			"refreshable": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "The token is not refreshable by default.",
			},
			"description": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 1024)),
				Description:      "Free text token description. Useful for filtering and managing tokens. Limited to 1024 characters.",
			},
			"audiences": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of the other instances or services that should accept this token identified by their Service-IDs. Limited to total 255 characters. Default to '*@*' if not set. For instructions to retrieve the Artifactory Service ID see this [documentation](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-GetServiceID).",
			},
			"access_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"token_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subject": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiry": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"issued_at": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"issuer": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Description: "Create scoped tokens for any of the services in your JFrog Platform and to manage user access to these services. If left at the default setting, the token will be created with the user-identity scope, which allows users to identify themselves in the Platform but does not grant any specific access permissions.",
	}
}

func CheckAccessToken(id string, request *resty.Request) (*resty.Response, error) {
	return request.SetPathParam("id", id).Get(accessTokenUrl)
}
