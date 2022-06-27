package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/exp/maps"
)

func ResourceArtifactoryManagedUser() *schema.Resource {
	managedUserSchema := map[string]*schema.Schema{
		"password": {
			Type:             schema.TypeString,
			Sensitive:        true,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Password for the user.",
		},
	}
	maps.Copy(managedUserSchema, baseUserSchema)

	return &schema.Resource{
		CreateContext: resourceManagedUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: managedUserSchema,

		Description: "Provides an Artifactory managed user resource. This can be used to create and manage Artifactory users. For example, service account where password is known and managed externally.",
	}
}

func resourceManagedUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceBaseUserCreate(ctx, d, m, nil)
}
