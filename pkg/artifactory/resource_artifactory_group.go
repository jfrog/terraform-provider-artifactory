package artifactory

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

const groupsEndpoint = "artifactory/api/security/groups/"

var autoJoinAdminValidate = func() func(i interface{}, k string) ([]string, []error) {
	autoJoin, admin := false, false

	return func(value interface{}, key string) ([]string, []error) {
		switch key {
		case "auto_join":
			autoJoin = value.(bool)
		case "admin_privileges":
			admin = value.(bool)
		}
		if autoJoin && admin {
			return nil, []error{fmt.Errorf("admin privs on auto_join groups is not allowed")}
		}
		return nil, nil
	}
}()

func resourceArtifactoryGroup() *schema.Resource {
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
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				//ValidateFunc: autoJoinAdminValidate,
			},
			"admin_privileges": {
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				//ValidateFunc: autoJoinAdminValidate,
			},
			"realm": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateLowerCase,
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
		},
	}
}

func groupParams(s *schema.ResourceData) (services.GroupParams, error) {
	d := &ResourceData{s}

	group := services.Group{
		Name:            d.getString("name", false),
		Description:     d.getString("description", false),
		AutoJoin:        d.getBool("auto_join", false),
		AdminPrivileges: d.getBool("admin_privileges", false),
		Realm:           d.getString("realm", false),
		RealmAttributes: d.getString("realm_attributes", false),
	}
	if usersNames := d.getSetRef("users_names"); usersNames != nil {
		group.UsersNames = *usersNames
	}

	// Validator
	if group.AdminPrivileges && group.AutoJoin {
		return services.GroupParams{}, fmt.Errorf("error: auto_join cannot be true if admin_privileges is true")
	}

	return services.GroupParams{
			GroupDetails:    group,
			ReplaceIfExists: true,
			IncludeUsers:    true,
		},
		nil
}

func resourceGroupCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	groupParams, err := groupParams(d)
	if err != nil {
		return err
	}
	_, err = client.R().SetBody(&(groupParams.GroupDetails)).Put(groupsEndpoint + groupParams.GroupDetails.Name)

	if err != nil {
		return err
	}

	d.SetId(groupParams.GroupDetails.Name)
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

func resourceGroupGet(d *schema.ResourceData, m interface{}) (*services.Group, error) {
	client := m.(*ArtClient).Resty

	params := services.GroupParams{}
	params.GroupDetails.Name = d.Id()
	params.IncludeUsers = true

	group := services.Group{}
	url := fmt.Sprintf("%s%s?includeUsers=%t", groupsEndpoint, params.GroupDetails.Name, params.IncludeUsers)
	_, err := client.R().SetResult(&group).Get(url)
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

	setValue := mkLens(d)
	setValue("name", group.Name)
	setValue("description", group.Description)
	setValue("auto_join", group.AutoJoin)
	setValue("admin_privileges", group.AdminPrivileges)
	setValue("realm", group.Realm)
	setValue("realm_attributes", group.RealmAttributes)
	errors := setValue("users_names", schema.NewSet(schema.HashString, castToInterfaceArr(group.UsersNames)))
	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed saving state for groups %q", errors)
	}
	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, m interface{}) error {
	groupParams, err := groupParams(d)
	if err != nil {
		return err
	}
	// Create and Update uses same endpoint, create checks for ReplaceIfExists and then uses put
	// Update instead uses POST which prevents removing users. This recreates the group with the same permissions and updated users

	_, err = m.(*ArtClient).Resty.R().SetBody(&(groupParams.GroupDetails)).Put(groupsEndpoint + d.Id())
	if err != nil {
		return err
	}

	d.SetId(groupParams.GroupDetails.Name)
	return resourceGroupRead(d, m)
}

func resourceGroupDelete(d *schema.ResourceData, m interface{}) error {
	_, err := m.(*ArtClient).Resty.R().Delete(groupsEndpoint + d.Id())
	return err
}

func resourceGroupExists(d *schema.ResourceData, m interface{}) (bool, error) {
	_, err := m.(*ArtClient).Resty.R().Head(groupsEndpoint + d.Id())
	return err == nil, err
}
