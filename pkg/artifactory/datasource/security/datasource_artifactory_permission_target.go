package security

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/security"
	"net/http"
)

func DataSourceArtifactoryPermissionTarget() *schema.Resource {
	dataSourcePermissionTargetRead := func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		permissionTarget := new(security.PermissionTargetParams)
		targetName := d.Get("name").(string)
		resp, err := m.(*resty.Client).R().SetResult(permissionTarget).Get(security.PermissionsEndPoint + targetName)

		d.SetId(permissionTarget.Name)
		// TODO: We removed this error check from users and groups, but I forget why. Figure out why, then remove or keep this.
		if err != nil {
			if resp != nil && resp.StatusCode() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}

		// My intuition/memory makes me think that this call handles the error, it doesn't need to be handled twice?
		return security.PackPermissionTarget(permissionTarget, d)
	}
	return &schema.Resource{
		ReadContext: dataSourcePermissionTargetRead,
		Schema:      security.BuildPermissionTargetSchema(),
		Description: "Provides the permission target data source. Contains information about a specific permission target.",
	}
}
