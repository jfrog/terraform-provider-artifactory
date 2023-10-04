package security

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
	"golang.org/x/net/context"
)

func DataSourceArtifactoryGroup() *schema.Resource {
	dataSourceGroupRead := func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		group := Group{}
		name := d.Get("name").(string)
		includeUsers := d.Get("include_users").(string)
		_, err := m.(utilsdk.ProvderMetadata).Client.R().SetResult(&group).SetQueryParam("includeUsers", includeUsers).Get(security.GroupsEndpoint + name)

		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(group.Name)

		pkr := packer.Universal(predicate.SchemaHasKey(groupSchema))
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

		includeGroupSchema := utilsdk.MergeMaps(groupSchema, includeUsersSchema)
		delete(includeGroupSchema, "detach_all_users")
		return includeGroupSchema
	}

	return &schema.Resource{
		ReadContext: dataSourceGroupRead,
		Schema:      getDataSourceGroupSchema(),
		Description: "Provides the Artifactory Group data source. Contains information about the group configuration and optionally provides its associated user list.",
	}
}

type Group struct {
	Name            string   `json:"name,omitempty"`
	Description     string   `json:"description,omitempty"`
	ExternalId      string   `json:"externalId"`
	AutoJoin        bool     `json:"autoJoin,omitempty"`
	AdminPrivileges bool     `json:"adminPrivileges,omitempty"`
	Realm           string   `json:"realm,omitempty"`
	RealmAttributes string   `json:"realmAttributes,omitempty"`
	UsersNames      []string `json:"userNames"`
	WatchManager    bool     `json:"watchManager"`
	PolicyManager   bool     `json:"policyManager"`
	ReportsManager  bool     `json:"reportsManager"`
}

var groupSchema = map[string]*schema.Schema{
	"name": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringIsNotEmpty,
	},
	"description": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"external_id": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "New external group ID used to configure the corresponding group in Azure AD.",
	},
	"auto_join": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"admin_privileges": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"realm": {
		Type:             schema.TypeString,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validator.LowerCase,
	},
	"realm_attributes": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"users_names": {
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Optional: true,
	},
	"detach_all_users": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"watch_manager": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: `When this override is set,  User in the group can manage Xray Watches on any resource type. Default value is 'false'.`,
	},
	"policy_manager": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: `When this override is set,  User in the group can set Xray security and compliance policies. Default value is 'false'.`,
	},
	"reports_manager": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: `When this override is set,  User in the group can manage Xray Reports. Default value is 'false'.`,
	},
}
