package artifactory

import (
	"context"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
)

func resourceArtifactoryPermissionTargets() *schema.Resource {
	principalSchema := schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"permissions": {
					Type:     schema.TypeSet,
					Elem:     &schema.Schema{Type: schema.TypeString},
					Set:      schema.HashString,
					Required: true,
				},
			},
		},
	}

	return &schema.Resource{
		Create: resourcePermissionTargetsCreateOrReplace,
		Read:   resourcePermissionTargetsRead,
		Update: resourcePermissionTargetsCreateOrReplace,
		Delete: resourcePermissionTargetsDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"includes_pattern": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "**",
			},
			"excludes_pattern": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"repositories": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},
			"users":  &principalSchema,
			"groups": &principalSchema,
		},
	}
}

func unmarshalPermissionTargets(s *schema.ResourceData) *artifactory.PermissionTargets {
	d := &ResourceData{s}

	pt := new(artifactory.PermissionTargets)

	pt.Name = d.GetStringRef("name")
	pt.IncludesPattern = d.GetStringRef("includes_pattern")
	pt.ExcludesPattern = d.GetStringRef("excludes_pattern")
	pt.Repositories = d.GetSetRef("repositories")

	if v, ok := d.GetOkExists("users"); ok {
		if pt.Principals == nil {
			pt.Principals = new(artifactory.Principals)
		}
		users := expandPrincipal(v.(*schema.Set))
		pt.Principals.Users = &users
	}

	if v, ok := d.GetOkExists("groups"); ok {
		if pt.Principals == nil {
			pt.Principals = new(artifactory.Principals)
		}
		groups := expandPrincipal(v.(*schema.Set))
		pt.Principals.Groups = &groups
	}

	return pt
}

func marshalPermissionTargets(permissionTargets *artifactory.PermissionTargets, d *schema.ResourceData) {
	d.Set("name", permissionTargets.Name)
	d.Set("includes_pattern", permissionTargets.IncludesPattern)
	d.Set("excludes_pattern", permissionTargets.ExcludesPattern)

	if permissionTargets.Repositories != nil {
		d.Set("repositories", schema.NewSet(schema.HashString, CastToInterfaceArr(*permissionTargets.Repositories)))
	}

	if permissionTargets.Principals.Users != nil {
		d.Set("users", flattenPrincipal(*permissionTargets.Principals.Users))
	}

	if permissionTargets.Principals.Groups != nil {
		d.Set("groups", flattenPrincipal(*permissionTargets.Principals.Groups))
	}
}

func expandPrincipal(s *schema.Set) map[string][]string {
	principal := map[string][]string{}

	for _, v := range s.List() {
		o := v.(map[string]interface{})
		principal[o["name"].(string)] = CastToStringArr(o["permissions"].(*schema.Set).List())
	}

	return principal
}

func flattenPrincipal(principal map[string][]string) *schema.Set {
	s := make([]interface{}, 0)

	for name, perms := range principal {
		user := map[string]interface{}{
			"name":        name,
			"permissions": schema.NewSet(schema.HashString, CastToInterfaceArr(perms)),
		}

		s = append(s, user)
	}

	return schema.NewSet(hashPrincipal, s)
}

func hashPrincipal(o interface{}) int {
	p := o.(map[string]interface{})
	return hashcode.String(p["name"].(string)) + 31*hashcode.String(hashcode.Strings(CastToStringArr(p["permissions"].(*schema.Set).List())))
}

func resourcePermissionTargetsCreateOrReplace(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	permissionTarget := unmarshalPermissionTargets(d)
	_, err := c.Security.CreateOrReplacePermissionTargets(context.Background(), *permissionTarget.Name, permissionTarget)
	if err != nil {
		return err
	}

	d.SetId(*permissionTarget.Name)
	return resourcePermissionTargetsRead(d, m)
}

func resourcePermissionTargetsRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	permissionTarget, resp, err := c.Security.GetPermissionTargets(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	marshalPermissionTargets(permissionTarget, d)
	return nil
}

func resourcePermissionTargetsDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)
	permissionTarget := unmarshalPermissionTargets(d)
	_, resp, err := c.Security.DeletePermissionTargets(context.Background(), *permissionTarget.Name)

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}
