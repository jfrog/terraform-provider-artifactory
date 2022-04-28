package security

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

const GroupsEndpoint = "artifactory/api/security/groups/"

func ResourceArtifactoryGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupCreate,
		Read:   resourceGroupRead,
		Update: resourceGroupUpdate,
		Delete: resourceGroupDelete,
		Exists: resourceGroupExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
		},
	}
}

func groupParams(s *schema.ResourceData) (Group, bool, error) {
	d := &util.ResourceData{s}

	group := Group{
		Name:            d.GetString("name", false),
		Description:     d.GetString("description", false),
		AutoJoin:        d.GetBool("auto_join", false),
		AdminPrivileges: d.GetBool("admin_privileges", false),
		Realm:           d.GetString("realm", false),
		RealmAttributes: d.GetString("realm_attributes", false),
		UsersNames:      d.GetSet("users_names"),
		WatchManager:    d.GetBool("watch_manager", false),
		PolicyManager:   d.GetBool("policy_manager", false),
		ReportsManager:  d.GetBool("reports_manager", false),
	}

	// Validator
	if group.AdminPrivileges && group.AutoJoin {
		return Group{}, false, fmt.Errorf("error: auto_join cannot be true if admin_privileges is true")
	}

	// includeUsers determines if tf is managing group membership
	// if not it shouldn't return users on the read since they arent in state
	// this means usersnames is always empty
	// so it also changes the update from put to post to prevent detaching all existing users
	// without an explict instruction

	includeUsers := len(group.UsersNames) > 0 || d.GetBool("detach_all_users", false)
	return group, includeUsers, nil
}

func resourceGroupCreate(d *schema.ResourceData, m interface{}) error {
	group, _, err := groupParams(d)
	if err != nil {
		return err
	}
	_, err = m.(*resty.Client).R().SetBody(group).Put(GroupsEndpoint + group.Name)

	if err != nil {
		return err
	}

	d.SetId(group.Name)
	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		exists, err := resourceGroupExists(d, m)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error describing group: %s", err))
		}

		if !exists {
			return resource.RetryableError(fmt.Errorf("expected group to be created, but currently not found"))
		}

		return nil
	})
}

func resourceGroupGet(d *schema.ResourceData, m interface{}) (*Group, error) {
	_, includeUsers, err := groupParams(d)
	if err != nil {
		return nil, err
	}

	group := Group{}
	url := fmt.Sprintf("%s%s?includeUsers=%t", GroupsEndpoint, d.Id(), includeUsers)
	_, err = m.(*resty.Client).R().SetResult(&group).Get(url)
	return &group, err
}

func resourceGroupRead(d *schema.ResourceData, m interface{}) error {
	group, err := resourceGroupGet(d, m)
	if err != nil {
		// If we 404 it is likely the resources was externally deleted
		// If the ID is updated to blank, this tells Terraform the resource no longer exist
		if group == nil {
			d.SetId("")
			return nil
		}
		return err
	}

	setValue := util.MkLens(d)
	setValue("name", group.Name)
	setValue("description", group.Description)
	setValue("auto_join", group.AutoJoin)
	setValue("admin_privileges", group.AdminPrivileges)
	setValue("realm", group.Realm)
	setValue("realm_attributes", group.RealmAttributes)
	setValue("watch_manager", group.WatchManager)
	setValue("policy_manager", group.PolicyManager)
	setValue("reports_manager", group.ReportsManager)
	errors := setValue("users_names", schema.NewSet(schema.HashString, util.CastToInterfaceArr(group.UsersNames)))
	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed saving state for groups %q", errors)
	}
	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, m interface{}) error {
	group, includeUsers, err := groupParams(d)
	if err != nil {
		return err
	}

	// Create and Update uses same endpoint, create checks for ReplaceIfExists and then uses put
	// This recreates the group with the same permissions and updated users
	// Update instead uses POST which prevents removing users and since it is only used when membership is empty
	// this results in a group where users are not managed by artifactory if users_names is not set.

	if includeUsers {
		_, err := m.(*resty.Client).R().SetBody(group).Put(GroupsEndpoint + d.Id())
		if err != nil {
			return err
		}
	} else {
		_, err = m.(*resty.Client).R().SetBody(group).Post(GroupsEndpoint + d.Id())
		if err != nil {
			return err
		}
	}

	d.SetId(group.Name)
	return resourceGroupRead(d, m)
}

func resourceGroupDelete(d *schema.ResourceData, m interface{}) error {
	_, err := m.(*resty.Client).R().Delete(GroupsEndpoint + d.Id())
	return err
}

func resourceGroupExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return groupExists(m.(*resty.Client), d.Id())
}

func groupExists(client *resty.Client, groupName string) (bool, error) {
	resp, err := client.R().Head(GroupsEndpoint + groupName)
	if err != nil && resp != nil && resp.StatusCode() == http.StatusNotFound {
		// Do not error on 404s as this causes errors when the upstream user has been manually removed
		return false, nil
	}

	return err == nil, err
}

// Group is a encoding struct to match
// https://www.jfrog.com/confluence/display/JFROG/Security+Configuration+JSON#SecurityConfigurationJSON-application/vnd.org.jfrog.artifactory.security.Group+json
type Group struct {
	Name            string   `json:"name,omitempty"`
	Description     string   `json:"description,omitempty"`
	AutoJoin        bool     `json:"autoJoin,omitempty"`
	AdminPrivileges bool     `json:"adminPrivileges,omitempty"`
	Realm           string   `json:"realm,omitempty"`
	RealmAttributes string   `json:"realmAttributes,omitempty"`
	UsersNames      []string `json:"userNames"`
	WatchManager    bool     `json:"watchManager"`
	PolicyManager   bool     `json:"policyManager"`
	ReportsManager  bool     `json:"reportsManager"`
}

func (g Group) Id() string {
	return g.Name
}
