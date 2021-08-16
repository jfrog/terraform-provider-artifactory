package artifactory

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	artifactory "github.com/jfrog/jfrog-client-go/artifactory/services"
)

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

func groupParams(s *schema.ResourceData) (artifactory.GroupParams, error) {
	d := &ResourceData{s}

	group := artifactory.Group{}

	if name := d.getStringRef("name", false); name != nil {
		group.Name = *name
	}

	if description := d.getStringRef("description", false); description != nil {
		group.Description = *description
	}

	if autoJoin := d.getBoolRef("auto_join", false); autoJoin != nil {
		group.AutoJoin = *autoJoin
	}

	if adminPrivileges := d.getBoolRef("admin_privileges", false); adminPrivileges != nil {
		group.AdminPrivileges = *adminPrivileges
	}

	if realm := d.getStringRef("realm", false); realm != nil {
		group.Realm = *realm
	}

	if realmAttributes := d.getStringRef("realm_attributes", false); realmAttributes != nil {
		group.RealmAttributes = *realmAttributes
	}

	if usersNames := d.getSetRef("users_names"); usersNames != nil {
		group.UsersNames = *usersNames
	}

	// Validator
	if group.AdminPrivileges && group.AutoJoin {
		return artifactory.GroupParams{}, fmt.Errorf("error: auto_join cannot be true if admin_privileges is true")
	}

	return artifactory.GroupParams{GroupDetails: group, ReplaceIfExists: true, IncludeUsers: true}, nil
}

func resourceGroupCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtNew

	groupParams, err := groupParams(d)
	if err != nil {
		return err
	}

	err = c.CreateGroup(groupParams)
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

		return resource.NonRetryableError(resourceGroupRead(d, m))
	})
}

func resourceGroupGet(d *schema.ResourceData, m interface{}) (*artifactory.Group, error) {
	c := m.(*ArtClient).ArtNew

	params := artifactory.NewGroupParams()
	params.GroupDetails.Name = d.Id()
	params.IncludeUsers = true

	return c.GetGroup(params)
}

func resourceGroupRead(d *schema.ResourceData, m interface{}) error {
	group, err := resourceGroupGet(d, m)
	if err != nil {
		return err
	}

	hasErr := false
	logError := cascadingErr(&hasErr)
	logError(d.Set("name", group.Name))
	logError(d.Set("description", group.Description))
	logError(d.Set("auto_join", group.AutoJoin))
	logError(d.Set("admin_privileges", group.AdminPrivileges))
	logError(d.Set("realm", group.Realm))
	logError(d.Set("realm_attributes", group.RealmAttributes))
	logError(d.Set("users_names", schema.NewSet(schema.HashString, castToInterfaceArr(group.UsersNames))))
	if hasErr {
		return fmt.Errorf("failed to marshal group")
	}
	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtNew
	groupParams, err := groupParams(d)
	if err != nil {
		return err
	}
	err = c.UpdateGroup(groupParams)
	if err != nil {
		return err
	}

	d.SetId(groupParams.GroupDetails.Name)
	return resourceGroupRead(d, m)
}

func resourceGroupDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtNew
	groupParams, err := groupParams(d)
	if err != nil {
		return err
	}

	return c.DeleteGroup(groupParams.GroupDetails.Name)
}

func resourceGroupExists(d *schema.ResourceData, m interface{}) (bool, error) {
	group, err := resourceGroupGet(d, m)
	if err != nil {
		return false, err
	}

	return (group != nil), nil
}
