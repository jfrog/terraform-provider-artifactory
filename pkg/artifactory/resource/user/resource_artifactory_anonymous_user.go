package user

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryAnonymousUser() *schema.Resource {

	type AnonymousUser struct {
		Name string `json:"name"`
	}

	anonymousUserSchema := map[string]*schema.Schema{
		// This isn't necessary in theory but Terraform doesn't like schema with no attributes
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Username for anonymous user. This should not be set in the HCL, or change after importing into Terraform state.",
		},
	}

	packAnonymousUser := func(user AnonymousUser, d *schema.ResourceData) diag.Diagnostics {

		setValue := util.MkLens(d)

		errors := setValue("name", user.Name)

		if errors != nil && len(errors) > 0 {
			return diag.Errorf("failed to pack anonymous user %q", errors)
		}

		return nil
	}

	resourceAnonymousUserRead := func(ctx context.Context, rd *schema.ResourceData, m interface{}) diag.Diagnostics {
		d := &util.ResourceData{ResourceData: rd}

		userName := d.Id()
		user := &AnonymousUser{}
		resp, err := m.(util.ProvderMetadata).Client.R().SetResult(user).Get(UsersEndpointPath + userName)

		if err != nil {
			if resp != nil && resp.StatusCode() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}

		return packAnonymousUser(*user, rd)
	}

	resourceAnonymousUserCreate := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		return diag.Errorf("Anonymous Artifactory user cannot be created. Use `terraform import` instead.")
	}

	resourceAnonymousUserUpdate := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		return diag.Errorf("Anonymous Artifactory user cannot be updated. Use `terraform import` instead.")
	}

	resourceAnonymousUserDelete := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		return diag.Errorf("Anonymous Artifactory user cannot be deleted. Use `terraform state rm` instead.")
	}

	return &schema.Resource{
		CreateContext: resourceAnonymousUserCreate,
		ReadContext:   resourceAnonymousUserRead,
		UpdateContext: resourceAnonymousUserUpdate,
		DeleteContext: resourceAnonymousUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: anonymousUserSchema,

		Description: "Provides an Artifactory anonymous user resource. This can be used to import Artifactory 'anonymous' uer for some use cases where this is useful.\n\nThis resource is not intended for managing the 'anonymous' user in Artifactory. Use the `resource_artifactory_user` resource instead.\n\n!> Anonymous user cannot be created from scratch, nor updated/deleted once imported into Terraform state.",
	}
}
