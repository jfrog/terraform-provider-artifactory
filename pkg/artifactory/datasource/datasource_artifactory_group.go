package datasource

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
	"golang.org/x/net/context"
	"net/http"
)

func ArtifactoryGroupSchema() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceGroupRead,
		Schema:      getDataSourceGroupSchema(),
		Description: "Provides the Artifactory Group data source. Contains information about the group configuration and optionally provides its associated user list.",
	}
}

func getDataSourceGroupSchema() map[string]*schema.Schema {
	var includeUsersSchema = map[string]*schema.Schema{
		"include_users": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     false,
			Description: "Setting includeUsers to true will return the group with its associated user list attached.",
		},
	}

	baseGroupSchema := security.GroupSchema
	delete(baseGroupSchema, "detach_all_users")
	return util.MergeMaps(baseGroupSchema, includeUsersSchema)
}

func DataSourceGroupRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	group := security.Group{}
	name := d.Get("name").(string)
	includeUsers := d.Get("include_users").(string)
	url := fmt.Sprintf("%s%s?includeUsers=%s", security.GroupsEndpoint, name, includeUsers)
	resp, err := m.(*resty.Client).R().SetResult(&group).Get(url)

	if err != nil {
		if resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound) {
			// If we 404 it is likely the resources was externally deleted
			// If the ID is updated to blank, this tells Terraform the resource no longer exist
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	d.SetId(group.Name)

	pkr := packer.Universal(predicate.SchemaHasKey(security.GroupSchema))

	return diag.FromErr(pkr(&group, d))
}
