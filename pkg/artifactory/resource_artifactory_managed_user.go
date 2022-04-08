package artifactory

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/exp/maps"
)

func resourceArtifactoryManagedUser() *schema.Resource {
	managedUserSchema := map[string]*schema.Schema{
		"password": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Required:    true,
			Description: "Password for the user.",
		},
	}
	maps.Copy(managedUserSchema, baseUserSchema)

	return &schema.Resource{
		CreateContext: resourceManagedUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Exists:        resourceUserExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: managedUserSchema,

		Description: "Provides an Artifactory managed user resource. This can be used to create and manage Artifactory users.",
	}
}

func resourceManagedUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceBaseUserCreate(ctx, d, m, nil)
}
