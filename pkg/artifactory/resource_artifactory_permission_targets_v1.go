package artifactory

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
)

func unpackPermissionTargetV1(s *schema.ResourceData) *v1.PermissionTargets {
	d := &ResourceData{s}

	pt := new(v1.PermissionTargets)

	pt.Name = d.getStringRef("name")
	pt.IncludesPattern = d.getStringRef("includes_pattern")
	pt.ExcludesPattern = d.getStringRef("excludes_pattern")
	pt.Repositories = d.getSetRef("repositories")

	if v, ok := d.GetOkExists("users"); ok {
		if pt.Principals == nil {
			pt.Principals = new(v1.Principals)
		}
		users := expandPrincipal(v.(*schema.Set))
		pt.Principals.Users = &users
	}

	if v, ok := d.GetOkExists("groups"); ok {
		if pt.Principals == nil {
			pt.Principals = new(v1.Principals)
		}
		groups := expandPrincipal(v.(*schema.Set))
		pt.Principals.Groups = &groups
	}

	return pt
}

func packPermissionTargetV1(permissionTargets *v1.PermissionTargets, d *schema.ResourceData) error {
	hasErr := false
	logError := cascadingErr(&hasErr)

	logError(d.Set("name", permissionTargets.Name))
	logError(d.Set("includes_pattern", permissionTargets.IncludesPattern))
	logError(d.Set("excludes_pattern", permissionTargets.ExcludesPattern))

	if permissionTargets.Repositories != nil {
		logError(d.Set("repositories", schema.NewSet(schema.HashString, castToInterfaceArr(*permissionTargets.Repositories))))
	}

	if permissionTargets.Principals.Users != nil {
		logError(d.Set("users", flattenPrincipal(*permissionTargets.Principals.Users)))
	}

	if permissionTargets.Principals.Groups != nil {
		logError(d.Set("groups", flattenPrincipal(*permissionTargets.Principals.Groups)))
	}

	if hasErr {
		return fmt.Errorf("failed to marshal permission target")
	}
	return nil
}

func expandPrincipal(s *schema.Set) map[string][]string {
	principal := map[string][]string{}

	for _, v := range s.List() {
		o := v.(map[string]interface{})
		principal[o["name"].(string)] = castToStringArr(o["permissions"].(*schema.Set).List())
	}

	return principal
}

func flattenPrincipal(principal map[string][]string) *schema.Set {
	s := make([]interface{}, 0)

	for name, perms := range principal {
		user := map[string]interface{}{
			"name":        name,
			"permissions": schema.NewSet(schema.HashString, castToInterfaceArr(perms)),
		}

		s = append(s, user)
	}

	return schema.NewSet(hashPrincipal, s)
}

func hashPrincipal(o interface{}) int {
	p := o.(map[string]interface{})
	return hashcode.String(p["name"].(string)) + 31*hashcode.String(hashcode.Strings(castToStringArr(p["permissions"].(*schema.Set).List())))
}

func resourcePermissionTargetV1CreateOrReplace(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	permissionTarget := unpackPermissionTargetV1(d)
	_, err := c.V1.Security.CreateOrReplacePermissionTargets(context.Background(), *permissionTarget.Name, permissionTarget)
	if err != nil {
		return err
	}

	d.SetId(*permissionTarget.Name)
	return resourcePermissionTargetV1Read(d, m)
}

func resourcePermissionTargetV1Read(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	permissionTarget, resp, err := c.V1.Security.GetPermissionTargets(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	return packPermissionTargetV1(permissionTarget, d)
}

func resourcePermissionTargetV1Delete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)
	permissionTarget := unpackPermissionTargetV1(d)
	_, resp, err := c.V1.Security.DeletePermissionTargets(context.Background(), *permissionTarget.Name)

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}
