package user

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
	"github.com/samber/lo"
)

type User struct {
	Name                     string   `json:"username"`
	Email                    string   `json:"email"`
	Password                 string   `json:"password,omitempty"`
	Admin                    bool     `json:"admin"`
	ProfileUpdatable         bool     `json:"profile_updatable"`
	DisableUIAccess          bool     `json:"disable_ui_access"`
	InternalPasswordDisabled bool     `json:"internal_password_disabled"`
	Groups                   []string `json:"groups,omitempty"`
}

type GroupsAddRemove struct {
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

var baseUserSchema = map[string]*schema.Schema{
	"name": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
		ValidateDiagFunc: validation.ToDiagFunc(
			validation.All(
				validation.StringIsNotEmpty,
				validation.StringMatch(
					regexp.MustCompile(`^[a-z0-9.\-_\@]+$`),
					"may contain lowercase letters, numbers and symbols: '.-_@'",
				),
			),
		),
		Description: "Username for user. May contain lowercase letters, numbers and symbols: '.-_@'",
	},
	"email": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validator.IsEmail,
		Description:      "Email for user.",
	},
	"admin": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "(Optional, Default: false) When enabled, this user is an administrator with all the ensuing privileges.",
	},
	"profile_updatable": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
		Description: "(Optional, Default: true) When enabled, this user can update their profile details (except for the password. " +
			"Only an administrator can update the password). There may be cases in which you want to leave " +
			"this unset to prevent users from updating their profile. For example, a departmental user with " +
			"a single password shared between all department members.",
	},
	"disable_ui_access": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
		Description: "(Optional, Default: true) When enabled, this user can only access the system through the REST API." +
			" This option cannot be set if the user has Admin privileges.",
	},
	"internal_password_disabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
		Description: "(Optional, Default: false) When enabled, disables the fallback mechanism for using an internal password when " +
			"external authentication (such as LDAP) is enabled.",
	},
	"groups": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Set:         schema.HashString,
		Optional:    true,
		Description: "List of groups this user is a part of. **Notes:** If this attribute is not specified then user's group membership is set to empty. User will not be part of default \"readers\" group automatically.",
	},
}

func unpackUser(s *schema.ResourceData) User {
	d := &utilsdk.ResourceData{ResourceData: s}
	return User{
		Name:                     d.GetString("name", false),
		Email:                    d.GetString("email", false),
		Password:                 d.GetString("password", false),
		Admin:                    d.GetBool("admin", false),
		ProfileUpdatable:         d.GetBool("profile_updatable", false),
		DisableUIAccess:          d.GetBool("disable_ui_access", false),
		InternalPasswordDisabled: d.GetBool("internal_password_disabled", false),
		Groups:                   d.GetSet("groups"),
	}
}

func PackUser(user User, d *schema.ResourceData) diag.Diagnostics {

	setValue := utilsdk.MkLens(d)

	setValue("name", user.Name)
	setValue("email", user.Email)
	setValue("admin", user.Admin)
	setValue("profile_updatable", user.ProfileUpdatable)
	setValue("disable_ui_access", user.DisableUIAccess)
	errors := setValue("internal_password_disabled", user.InternalPasswordDisabled)

	if user.Groups != nil {
		errors = setValue("groups", schema.NewSet(schema.HashString, utilsdk.CastToInterfaceArr(user.Groups)))
	}

	if len(errors) > 0 {
		return diag.Errorf("failed to pack user %q", errors)
	}

	return nil
}

const UsersEndpointPath = "access/api/v2/users"
const UserEndpointPath = "access/api/v2/users/{name}"
const UserGroupEndpointPath = "access/api/v2/users/{name}/groups"

func resourceUserRead(_ context.Context, rd *schema.ResourceData, m interface{}) diag.Diagnostics {
	d := &utilsdk.ResourceData{ResourceData: rd}

	var user User
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	resp, err := m.(util.ProvderMetadata).Client.R().
		SetPathParam("name", d.Id()).
		SetResult(&user).
		SetError(&artifactoryError).
		Get(UserEndpointPath)

	if err != nil {
		return diag.FromErr(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		d.SetId("")
		return nil
	}
	if resp.IsError() {
		return diag.Errorf("%s", artifactoryError.String())
	}

	return PackUser(user, rd)
}

func syncReadersGroup(ctx context.Context, client *resty.Client, plan User, actual User) error {
	toAdd, toRemove := lo.Difference(plan.Groups, actual.Groups)
	tflog.Debug(ctx, "syncReadersGroup", map[string]any{
		"toAdd":    toAdd,
		"toRemove": toRemove,
	})

	if len(toAdd) == 0 && len(toRemove) == 0 {
		return nil
	}

	groupsToAddRemove := GroupsAddRemove{
		Add:    toAdd,
		Remove: toRemove,
	}
	// Access API for creating user will add any groups with "auto_join = true" to the user's groups.
	// We use following PATCH call to sync up user's groups from TF to Artifactory.
	// This action will match the expectation for this resource so "groups" attribute matches what's on Artifactory.
	_, err := client.R().
		SetPathParam("name", plan.Name).
		SetBody(groupsToAddRemove).
		Patch(UserGroupEndpointPath)
	if err != nil {
		return err
	}

	return nil
}

func resourceBaseUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}, passwordGenerator func(*User) diag.Diagnostics) diag.Diagnostics {
	user := unpackUser(d)

	var diags diag.Diagnostics

	if passwordGenerator != nil && !user.InternalPasswordDisabled {
		diags = passwordGenerator(&user)
	}

	var result User
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	resp, err := m.(util.ProvderMetadata).Client.R().
		SetBody(user).
		SetResult(&result).
		SetError(&artifactoryError).
		Post(UsersEndpointPath)
	if err != nil {
		return diag.FromErr(err)
	}
	if resp.IsError() {
		return diag.Errorf("%s", artifactoryError.String())
	}

	d.SetId(user.Name)

	err = syncReadersGroup(ctx, m.(util.ProvderMetadata).Client, user, result)
	if err != nil {
		return diag.FromErr(err)
	}

	retryError := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var result User
		resp, e := m.(util.ProvderMetadata).Client.R().
			SetPathParam("name", user.Name).
			SetResult(&result).
			Get(UserEndpointPath)

		if e != nil {
			return retry.NonRetryableError(fmt.Errorf("error describing user: %s", err))
		}
		if resp.StatusCode() == http.StatusNotFound {
			return retry.RetryableError(fmt.Errorf("expected user to be created, but currently not found"))
		}

		PackUser(result, d)

		return nil
	})

	if retryError != nil {
		return diag.FromErr(retryError)
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	user := unpackUser(d)

	var result User
	var artifactoryError artifactory.ArtifactoryErrorsResponse
	resp, err := m.(util.ProvderMetadata).Client.R().
		SetPathParam("name", user.Name).
		SetBody(&user).
		SetResult(&result).
		SetError(&artifactoryError).
		Patch(UserEndpointPath)

	if err != nil {
		return diag.FromErr(err)
	}
	if resp.IsError() {
		return diag.Errorf("%s", artifactoryError.String())
	}

	err = syncReadersGroup(ctx, m.(util.ProvderMetadata).Client, user, result)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.Name)
	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(_ context.Context, rd *schema.ResourceData, m interface{}) diag.Diagnostics {
	d := &utilsdk.ResourceData{ResourceData: rd}
	userName := d.GetString("name", false)

	var artifactoryError artifactory.ArtifactoryErrorsResponse
	resp, err := m.(util.ProvderMetadata).Client.R().
		SetPathParam("name", userName).
		SetError(&artifactoryError).
		Delete(UserEndpointPath)
	if err != nil {
		return diag.Errorf("user %s not deleted. %s", userName, err)
	}
	if resp.IsError() {
		return diag.Errorf("%s", artifactoryError.String())
	}

	d.SetId("")

	return nil
}
