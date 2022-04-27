package user

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
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

		setValue := utils.MkLens(d)

		errors := setValue("name", user.Name)

		if errors != nil && len(errors) > 0 {
			return diag.Errorf("failed to pack anonymous user %q", errors)
		}

		return nil
	}

	resourceAnonymousUserRead := func(ctx context.Context, rd *schema.ResourceData, m interface{}) diag.Diagnostics {
		d := &utils.ResourceData{rd}

		userName := d.Id()
		user := &AnonymousUser{}
		resp, err := m.(*resty.Client).R().SetResult(user).Get(usersEndpointPath + userName)

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
		Exists:        resourceUserExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: anonymousUserSchema,

		Description: "Provides an Artifactory anonymouse user resource. This only supports importing from Artifactory through `terraform import` command. This cannot be created from scratch, nor updated/deleted once imported into Terraform state.",
	}
}
