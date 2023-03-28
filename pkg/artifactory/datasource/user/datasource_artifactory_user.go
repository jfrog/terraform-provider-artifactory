package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func DataSourceArtifactoryUser() *schema.Resource {
	var userSchema = map[string]*schema.Schema{
		"name": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Username for user.",
		},
		"email": {
			Type:             schema.TypeString,
			Optional:         true,
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
			Description: "List of groups this user is a part of.",
		},
	}

	read := func(_ context.Context, rd *schema.ResourceData, m interface{}) diag.Diagnostics {
		d := &util.ResourceData{ResourceData: rd}

		userName := d.Get("name").(string)
		userObj := user.User{}
		_, err := m.(util.ProvderMetadata).Client.R().SetResult(&userObj).Get(user.UsersEndpointPath + userName)

		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(userObj.Name)

		return user.PackUser(userObj, rd)
	}

	return &schema.Resource{
		ReadContext: read,
		Schema:      userSchema, // note this does not include password of the user, don't think we should return it as a datasource
		Description: "Provides the Artifactory User data source. ",
	}
}
