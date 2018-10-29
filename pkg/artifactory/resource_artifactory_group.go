package artifactory

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auto_join": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"admin_privileges": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"realm": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "internal",
				ValidateFunc: validateLowerCase,
			},
			"realm_attributes": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func unmarshalGroup(s *schema.ResourceData) (*artifactory.Group, error) {
	d := &ResourceData{s}

	group := new(artifactory.Group)

	group.Name = d.GetStringRef("name")
	group.Description = d.GetStringRef("description")
	group.AutoJoin = d.GetBoolRef("auto_join")
	group.AdminPrivileges = d.GetBoolRef("admin_privileges")
	group.Realm = d.GetStringRef("realm")
	group.RealmAttributes = d.GetStringRef("realm_attributes")

	// Validator
	if group.AdminPrivileges != nil && group.AutoJoin != nil && *group.AdminPrivileges && *group.AutoJoin {
		return nil, fmt.Errorf("error: auto_join cannot be true if admin_privileges is true")
	}

	return group, nil
}

func resourceGroupCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	group, err := unmarshalGroup(d)

	if err != nil {
		return err
	}

	_, err = c.Security.CreateOrReplaceGroup(context.Background(), *group.Name, group)

	if err != nil {
		return err
	}

	d.SetId(*group.Name)
	return resourceGroupRead(d, m)
}

func resourceGroupRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	group, resp, err := c.Security.GetGroup(context.Background(), d.Id())

	// If we 404 it is likely the resources was externally deleted
	// If the ID is updated to blank, this tells Terraform the resource no longer exist
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("auto_join", group.AutoJoin)
	d.Set("admin_privileges", group.AdminPrivileges)
	d.Set("realm", group.Realm)
	d.Set("realm_attributes", group.RealmAttributes)
	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)
	group, err := unmarshalGroup(d)
	if err != nil {
		return err
	}
	_, err = c.Security.UpdateGroup(context.Background(), d.Id(), group)
	if err != nil {
		return err
	}

	d.SetId(*group.Name)
	return resourceGroupRead(d, m)
}

func resourceGroupDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)
	group, err := unmarshalGroup(d)
	if err != nil {
		return err
	}

	_, resp, err := c.Security.DeleteGroup(context.Background(), *group.Name)

	if err != nil && resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}

func resourceGroupExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*artifactory.Client)

	groupName := d.Id()
	_, resp, err := c.Security.GetGroup(context.Background(), groupName)

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
