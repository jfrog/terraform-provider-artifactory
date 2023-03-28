package security

import (
	"golang.org/x/net/context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

func DataSourceArtifactoryGroup() *schema.Resource {
	dataSourceGroupRead := func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		group := security.Group{}
		name := d.Get("name").(string)
		includeUsers := d.Get("include_users").(string)
		_, err := m.(util.ProvderMetadata).Client.R().SetResult(&group).SetQueryParam("includeUsers", includeUsers).Get(security.GroupsEndpoint + name)

		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(group.Name)

		pkr := packer.Universal(predicate.SchemaHasKey(security.GroupSchema))
		return diag.FromErr(pkr(&group, d))
	}

	getDataSourceGroupSchema := func() map[string]*schema.Schema {
		includeUsersSchema := map[string]*schema.Schema{
			"include_users": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     false,
				Description: "Setting includeUsers to true will return the group with its associated user list attached.",
			},
		}

		includeGroupSchema := util.MergeMaps(security.GroupSchema, includeUsersSchema)
		delete(includeGroupSchema, "detach_all_users")
		return includeGroupSchema
	}

	return &schema.Resource{
		ReadContext: dataSourceGroupRead,
		Schema:      getDataSourceGroupSchema(),
		Description: "Provides the Artifactory Group data source. Contains information about the group configuration and optionally provides its associated user list.",
	}
}
