package artifactory

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArtifactoryAnonymousUser() *schema.Resource {
	managedUserSchema := map[string]*schema.Schema{
		"name": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "Username for user.",
		},
		"email": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "Email for user.",
		},
		"admin": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "When enabled, this user is an administrator with all the ensuing privileges.",
		},
		"profile_updatable": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
			Description: "When enabled, this user can update their profile details (except for the password. " +
				"Only an administrator can update the password). There may be cases in which you want to leave " +
				"this unset to prevent users from updating their profile. For example, a departmental user with " +
				"a single password shared between all department members.",
		},
		"disable_ui_access": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
			Description: "When enabled, this user can only access the system through the REST API." +
				" This option cannot be set if the user has Admin privileges.",
		},
		"internal_password_disabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
			Description: "When enabled, disables the fallback mechanism for using an internal password when " +
				"external authentication (such as LDAP) is enabled.",
		},
		"groups": {
			Type:        schema.TypeSet,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Set:         schema.HashString,
			Optional:    true,
			Computed:    true,
			Description: "List of groups this user is a part of.",
		},
		"password": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Optional:    true,
			Computed:    true,
			Description: "Password for the user.",
		},
	}

	return &schema.Resource{
		CreateContext: resourceAnonymousUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Exists:        resourceUserExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: managedUserSchema,
	}
}

func resourceAnonymousUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceBaseUserCreate(ctx, d, m, nil)
}
