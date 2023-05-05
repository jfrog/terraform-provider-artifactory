package security

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/go-querystring/query"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

// AccessTokenRevokeOptions jfrog client go has no v1 code and moving to v2 would be a lot of work.
// To remove the dependency, we copy and past it here
type AccessTokenRevokeOptions struct {
	Token string `url:"token,omitempty"`
}

type AccessTokenOptions struct {
	// The grant type used to authenticate the request. In this case, the only value supported is "client_credentials" which is also the default value if this parameter is not specified.
	GrantType string `url:"grant_type,omitempty"` // [Optional, default: "client_credentials"]
	// The user name for which this token is created. If the user does not exist, a transient user is created. Non-admin users can only create tokens for themselves so they must specify their own username.
	// If the user does not exist, the member-of-groups scope token must be provided (e.g. member-of-groups: g1, g2, g3...)
	Username string `url:"username,omitempty"`
	// The scope to assign to the token provided as a space-separated list of scope tokens. Currently there are three possible scope tokens:
	//     - "api:*" - indicates that the token grants access to REST API calls. This is always granted by default whether specified in the call or not.
	//     - member-of-groups:[<group-name>] - indicates the groups that the token is associated with (e.g. member-of-groups: g1, g2, g3...). The token grants access according to the permission targets specified for the groups listed.
	//       Specify "*" for group-name to indicate that the token should provide the same access privileges that are given to the group of which the logged in user is a member.
	//       A non-admin user can only provide a scope that is a subset of the groups to which he belongs
	//     - "jfrt@<instance-id>:admin" - provides admin privileges on the specified Artifactory instance. This is only available for administrators.
	// If omitted and the username specified exists, the token is granted the scope of that user.
	Scope string `url:"scope,omitempty"` // [Optional if the user specified in username exists]
	// The time in seconds for which the token will be valid. To specify a token that never expires, set to zero. Non-admin can only set a value that is equal to or less than the default 3600.
	ExpiresIn int `url:"expires_in"` // [Optional, default: 3600]
	// If true, this token is refreshable and the refresh token can be used to replace it with a new token once it expires.
	Refreshable string `url:"refreshable,omitempty"` // [Optional, default: false]
	// A space-separate list of the other Artifactory instances or services that should accept this token identified by their Artifactory Service IDs as obtained from the Get Service ID endpoint.
	// In case you want the token to be accepted by all Artifactory instances you may use the following audience parameter "audience=jfrt@*".
	Audience string `url:"audience,omitempty"` // [Optional, default: Only the Service ID of the Artifactory instance that created the token]
}

func ResourceArtifactoryAccessToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccessTokenCreate,
		ReadContext:   resourceAccessTokenRead,
		DeleteContext: resourceAccessTokenDelete,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"audience": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"end_date_relative": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"end_date"},
				ValidateFunc: func(i interface{}, k string) ([]string, []error) {
					v, ok := i.(string)
					if !ok {
						return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
					}

					if strings.TrimSpace(v) == "" {
						return nil, []error{fmt.Errorf("%q must not be empty", k)}
					}

					return nil, nil
				},
				AtLeastOneOf: []string{"end_date", "end_date_relative"},
			},
			"end_date": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"end_date_relative"},
				ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
					v, ok := i.(string)
					if !ok {
						errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
						return warnings, errors
					}

					if _, err := time.Parse(time.RFC3339, v); err != nil {
						errors = append(errors, fmt.Errorf("expected %q to be a valid RFC3339 date, got %q: %+v", k, i, err))
					}

					return warnings, errors
				},
				AtLeastOneOf: []string{"end_date", "end_date_relative"},
			},
			"refreshable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"admin_token": {
				Type:          schema.TypeSet,
				ConflictsWith: []string{"groups"},
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"access_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"refresh_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},

		DeprecationMessage: "This resource is being deprecated and replaced by artifactory_scoped_token",
	}
}

func resourceAccessTokenCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	type AccessToken struct {
		AccessToken  string `json:"access_token,omitempty"`
		ExpiresIn    int    `json:"expires_in,omitempty"`
		Scope        string `json:"scope,omitempty"`
		TokenType    string `json:"token_type,omitempty"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}

	client := m.(utilsdk.ProvderMetadata).Client
	grantType := "client_credentials" // client_credentials is the only supported type

	tokenOptions := AccessTokenOptions{}
	resourceData := &utilsdk.ResourceData{ResourceData: d}

	date, expiresIn, err := getDate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	tokenOptions.ExpiresIn = expiresIn
	err = d.Set("end_date", date.Format(time.RFC3339))
	if err != nil {
		return diag.FromErr(err)
	}

	refreshable := resourceData.Get("refreshable").(bool)
	audience := d.Get("audience").(string)

	tokenOptions.Audience = audience
	tokenOptions.GrantType = grantType
	tokenOptions.Refreshable = strconv.FormatBool(refreshable)
	tokenOptions.Username = resourceData.GetString("username", false)

	username := resourceData.Get("username").(string)
	userExists, _ := checkUserExists(client, username)

	if !userExists && len(resourceData.Get("groups").([]interface{})) == 0 {
		return diag.Errorf("you must specify at least 1 group when creating a token for a non-existant user - %s, or correct the username", username)
	}

	err = unpackGroups(d, client, &tokenOptions)
	if err != nil {
		return diag.FromErr(err)
	}

	err = unpackAdminToken(d, &tokenOptions)
	if err != nil {
		return diag.FromErr(err)
	}

	accessToken := AccessToken{}
	values, err := TokenOptsToValues(tokenOptions)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = m.(utilsdk.ProvderMetadata).Client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetResult(&accessToken).
		SetFormDataFromValues(values).Post("artifactory/api/security/token")

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(schema.HashString(accessToken.AccessToken)))

	err = d.Set("access_token", accessToken.AccessToken)
	if err != nil {
		return diag.FromErr(err)
	}

	refreshToken := ""
	if refreshable {
		refreshToken = accessToken.RefreshToken
	}

	err = d.Set("refresh_token", refreshToken)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAccessTokenRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Terraform requires that the read function is always implemented.
	// However, Artifactory does not have an API to read a token.
	return nil
}

func resourceAccessTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Artifactory only allows you to revoke a token if the there is no expiry.
	// Otherwise, Artifactory will ensure the token is revoked at the expiry time.
	// https://www.jfrog.com/confluence/display/JFROG/Access+Tokens#AccessTokens-ViewingandRevokingTokens
	// https://www.jfrog.com/jira/browse/RTFACT-15293

	// If relative end date is empty, then a fixed end date was set
	// Therefore, Artifactory will expire the token automatically
	endDateRelative := d.Get("end_date_relative").(string)
	if endDateRelative == "" {
		tflog.Debug(ctx, "AccessToken is not revoked. It will expire at "+d.Get("end_date").(string))
		return nil
	}

	// Convert end date relative to duration in seconds
	duration, err := time.ParseDuration(endDateRelative)
	if err != nil {
		return diag.Errorf("unable to parse `end_date_relative` (%s) as a duration", endDateRelative)
	}

	// If the token has no duration, it does not expire.
	// Therefore revoke the token.
	if duration.Seconds() == 0 {
		tflog.Debug(ctx, "Revoking token")
		revokeOptions := AccessTokenRevokeOptions{}
		revokeOptions.Token = d.Get("access_token").(string)
		values, err := query.Values(revokeOptions)
		resp, err := m.(utilsdk.ProvderMetadata).Client.R().
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetFormDataFromValues(values).Post("artifactory/api/security/token/revoke")
		if err != nil {
			if resp != nil {
				if resp.StatusCode() == http.StatusNotFound {
					tflog.Debug(ctx, "Access Token Revoked")
					return nil
				}
				// the original atlassian code considered any error code fine. However, expiring tokens can't be revoked
				regex := regexp.MustCompile(`.*AccessToken not revocable.*`)
				if regex.MatchString(string(resp.Body()[:])) {
					return nil
				}
			}
			return diag.FromErr(err)
		}
		return nil
	}

	// If the duration is set, Artifactory will automatically revoke the token.
	tflog.Debug(ctx, "AccessToken is not revoked. It will expire at "+d.Get("end_date").(string))

	return nil
}

func unpackGroups(d *schema.ResourceData, client *resty.Client, tokenOptions *AccessTokenOptions) error {
	if srcGroups, ok := d.GetOk("groups"); ok {
		groups := make([]string, len(srcGroups.([]interface{})))
		for i, group := range srcGroups.([]interface{}) {
			groups[i] = group.(string)

			if groups[i] != "*" {
				if exist, err := checkGroupExists(client, groups[i]); !exist {
					return err
				}
			}
		}

		scopedGroupString := strings.Join(groups[:], ",")
		scope := "member-of-groups:\"" + scopedGroupString + "\""
		tokenOptions.Scope = scope
	}

	return nil
}

func unpackAdminToken(d *schema.ResourceData, tokenOptions *AccessTokenOptions) error {
	if adminToken, ok := d.GetOk("admin_token"); ok {
		set := adminToken.(*schema.Set)
		val := set.List()[0].(map[string]interface{})

		instanceID := val["instance_id"].(string)

		scope := "jfrt@" + instanceID + ":admin"
		tokenOptions.Scope = scope
	}

	return nil
}

func checkUserExists(client *resty.Client, name string) (bool, error) {
	resp, err := client.R().Head("artifactory/api/security/users/" + name)
	if err != nil {
		// If there is an error, it is possible the user does not exist.
		if resp != nil {
			// Check if the user does not exist in artifactory
			if resp.StatusCode() == http.StatusNotFound {
				return false, errors.New("user must exist in artifactory")
			}

			// If we cannot search for Users, the current user is not an admin
			// So, we'll let this through and let the CreateToken function error if there is a misconfiguration.
			if resp.StatusCode() == http.StatusForbidden {
				return true, nil
			}
		}
		return false, err
	}

	return true, nil
}

func checkGroupExists(client *resty.Client, name string) (bool, error) {
	resp, err := client.R().Head(GroupsEndpoint + name)
	// If there is an error, it is possible the group does not exist.
	if err != nil {
		if resp != nil {
			// Check if the group does not exist in artifactory
			if resp.StatusCode() == http.StatusNotFound {
				return false, errors.New("group must exist in artifactory")
			}

			// If we cannot search for groups, the current user is not an admin and they can only specify groups they belong to.
			// Therefore, we return true and rely on Artifactory to error if the user has specified a wrong group.
			if resp.StatusCode() == http.StatusForbidden {
				return true, nil
			}
		}

		return false, err
	}

	return true, nil
}

// Inspired by azure ad implementation
func getDate(d *schema.ResourceData) (time.Time, int, error) {
	var endDate time.Time
	now := time.Now()

	if v := d.Get("end_date").(string); v != "" {
		var err error
		endDate, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return endDate, -1, fmt.Errorf("unable to parse the provided end date %q: %+v", v, err)
		}
	} else if v := d.Get("end_date_relative").(string); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return endDate, -1, fmt.Errorf("unable to parse `end_date_relative` %s as a duration", v)
		}
		// Artifactory's minimum duration is in seconds.
		// The consumer should either specify 0 for a non-expiring token, or >= 1 seconds
		// If the consumer passes in configuration that is between 0 and 1 seconds, they have used a smaller time unit that seconds.
		if d.Nanoseconds() > 0 && d.Seconds() < 1 {
			return endDate, -1, fmt.Errorf("minimum duration is 1 second, but `end_date_relative` is %s", v)
		}
		endDate = time.Now().Add(d)
	} else {
		return endDate, -1, fmt.Errorf("one of `end_date` or `end_date_relative` must be specified")
	}

	differenceInSeconds := int(endDate.Sub(now).Seconds())

	if differenceInSeconds < 0 {
		return endDate, -1, fmt.Errorf("end date must be in the future, but is %s", endDate.String())
	}
	return endDate, differenceInSeconds, nil
}

func TokenOptsToValues(t AccessTokenOptions) (url.Values, error) {
	return query.Values(t)
}
