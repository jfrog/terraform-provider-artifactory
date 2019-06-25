package artifactory

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"net/http"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	v2 "github.com/atlassian/go-artifactory/v2/artifactory/v2"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceArtifactoryPermissionTargets() *schema.Resource {
	resource := resourceArtifactoryPermissionTarget()
	resource.DeprecationMessage = "Since v1.5. Use artifactory_permission_target"
	return resource
}

func resourceArtifactoryPermissionTarget() *schema.Resource {
	principalSchemaV1 := schema.Schema{
		Type:          schema.TypeSet,
		Optional:      true,
		Deprecated:    "Since Artifactory 6.6.0+. Use /api/v2 endpoint",
		ConflictsWith: []string{"repo", "build"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
					// Required as it is impossible to remove a principal as the absence of one does not
					// count as a deletion
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

	actionSchema := schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Set:      hashPrincipal,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"permissions": {
					Type: schema.TypeSet,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{
							v2.PERM_READ,
							v2.PERM_ANNOTATE,
							v2.PERM_WRITE,
							v2.PERM_DELETE,
							v2.PERM_MANAGE,
						}, false),
					},
					Set:      schema.HashString,
					Required: true,
				},
			},
		},
	}

	principalSchema := schema.Schema{
		Type:          schema.TypeList,
		ConflictsWith: []string{"repositories"},
		Optional:      true,
		MaxItems:      1,
		MinItems:      1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"includes_pattern": {
					Type:     schema.TypeSet,
					Elem:     &schema.Schema{Type: schema.TypeString},
					Set:      schema.HashString,
					Optional: true,
				},
				"excludes_pattern": {
					Type:     schema.TypeSet,
					Elem:     &schema.Schema{Type: schema.TypeString},
					Set:      schema.HashString,
					Optional: true,
				},
				"repositories": {
					Type:     schema.TypeSet,
					Elem:     &schema.Schema{Type: schema.TypeString},
					Set:      schema.HashString,
					Required: true,
				},
				"actions": {
					Type:     schema.TypeList,
					MaxItems: 1,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"users":  &actionSchema,
							"groups": &actionSchema,
						},
					},
					Optional: true,
				},
			},
		},
	}

	return &schema.Resource{
		Create: resourcePermissionTargetCreate,
		Read:   resourcePermissionTargetRead,
		Update: resourcePermissionTargetUpdate,
		Delete: resourcePermissionTargetDelete,
		Exists: resourcePermissionTargetExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repo":  &principalSchema,
			"build": &principalSchema,

			// Legacy V1 Fields
			"includes_pattern": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "**",
				Deprecated:    "Since Artifactory 6.6.0+ (provider 1.5). Use /api/v2 endpoint",
				ConflictsWith: []string{"repo", "build"},
			},
			"excludes_pattern": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Deprecated:    "Since Artifactory 6.6.0+ (provider 1.5). Use /api/v2 endpoint",
				ConflictsWith: []string{"repo", "build"},
			},
			"repositories": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
				Optional:      true,
				Deprecated:    "Since Artifactory 6.6.0+ (provider 1.5). Use /api/v2 endpoint",
				ConflictsWith: []string{"repo", "build"},
			},
			"users":  &principalSchemaV1,
			"groups": &principalSchemaV1,
		},
	}
}

func unpackPermissionTarget(s *schema.ResourceData) *v2.PermissionTarget {
	d := &ResourceData{s}

	unpackPermission := func(rawPermissionData interface{}) *v2.Permission {
		unpackEntity := func(rawEntityData interface{}) *v2.Entity {
			unpackPermMap := func(rawPermSet interface{}) *map[string][]string {
				permList := rawPermSet.(*schema.Set).List()
				if len(permList) == 0 {
					return nil
				}

				permissions := make(map[string][]string)
				for _, v := range permList {
					id := v.(map[string]interface{})

					permissions[id["name"].(string)] = castToStringArr(id["permissions"].(*schema.Set).List())
				}
				return &permissions
			}

			entityDataList := rawEntityData.([]interface{})
			if len(entityDataList) == 0 {
				return nil
			}

			entityData := entityDataList[0].(map[string]interface{})
			return &v2.Entity{
				Users:  unpackPermMap(entityData["users"]),
				Groups: unpackPermMap(entityData["groups"]),
			}
		}

		if rawPermissionData == nil || rawPermissionData.([]interface{})[0] == nil {
			return nil
		}

		// It is safe to unpack the zeroth element immediately since permission targets have min size of 1
		permissionData := rawPermissionData.([]interface{})[0].(map[string]interface{})

		permission := new(v2.Permission)

		// This will always exist
		{
			tmp := castToStringArr(permissionData["repositories"].(*schema.Set).List())
			permission.Repositories = &tmp
		}

		// Handle optionals
		if v, ok := permissionData["includes_pattern"]; ok {
			// It is not possible to set default values for sets. Therefore we should send the empty set,
			// so artifactory remote does not default to ** after creation.
			tmp := castToStringArr(v.(*schema.Set).List())
			permission.IncludePatterns = &tmp
		}
		if v, ok := permissionData["excludes_pattern"]; ok {
			tmp := castToStringArr(v.(*schema.Set).List())
			permission.ExcludePatterns = &tmp
		}
		if v, ok := permissionData["actions"]; ok {
			permission.Actions = unpackEntity(v)
		}

		return permission
	}

	pTarget := new(v2.PermissionTarget)

	pTarget.Name = d.getStringRef("name", false)

	if v, ok := d.GetOk("repo"); ok {
		pTarget.Repo = unpackPermission(v)
	}

	if v, ok := d.GetOk("build"); ok {
		pTarget.Build = unpackPermission(v)
	}

	return pTarget
}

func packPermissionTarget(permissionTarget *v2.PermissionTarget, d *schema.ResourceData) error {
	packPermission := func(p *v2.Permission) []interface{} {
		packPermMap := func(e map[string][]string) []interface{} {
			perm := make([]interface{}, len(e))

			count := 0
			for k, v := range e {
				perm[count] = map[string]interface{}{
					"name":        k,
					"permissions": schema.NewSet(schema.HashString, castToInterfaceArr(v)),
				}
				count++
			}

			return perm
		}

		s := map[string]interface{}{}

		if p != nil {
			if p.IncludePatterns != nil {
				s["includes_pattern"] = schema.NewSet(schema.HashString, castToInterfaceArr(*p.IncludePatterns))
			}

			if p.ExcludePatterns != nil {
				s["excludes_pattern"] = schema.NewSet(schema.HashString, castToInterfaceArr(*p.ExcludePatterns))
			}

			if p.Repositories != nil {
				s["repositories"] = schema.NewSet(schema.HashString, castToInterfaceArr(*p.Repositories))
			}

			if p.Actions != nil {
				perms := make(map[string]interface{})

				if p.Actions.Users != nil {
					perms["users"] = schema.NewSet(hashPrincipal, packPermMap(*p.Actions.Users))
				}

				if p.Actions.Groups != nil {
					perms["groups"] = schema.NewSet(hashPrincipal, packPermMap(*p.Actions.Groups))
				}

				if len(perms) > 0 {
					s["actions"] = []interface{}{perms}
				}
			}
		}

		return []interface{}{s}
	}

	hasErr := false
	logErrors := cascadingErr(&hasErr)

	logErrors(d.Set("name", permissionTarget.Name))
	if permissionTarget.Repo != nil {
		logErrors(d.Set("repo", packPermission(permissionTarget.Repo)))
	}
	if permissionTarget.Build != nil {
		logErrors(d.Set("build", packPermission(permissionTarget.Build)))
	}

	if hasErr {
		return fmt.Errorf("failed to marshal permission target")
	}
	return nil
}

func resourcePermissionTargetCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	if _, ok := d.GetOk("repositories"); ok {
		return resourcePermissionTargetV1CreateOrReplace(d, m)
	}

	permissionTarget := unpackPermissionTarget(d)

	_, err := c.V2.Security.CreatePermissionTarget(context.Background(), *permissionTarget.Name, permissionTarget)
	if err != nil {
		return err
	}

	d.SetId(*permissionTarget.Name)
	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		c := m.(*artifactory.Artifactory)
		exists, err := c.V2.Security.HasPermissionTarget(context.Background(), d.Id())
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error describing permssions target: %s", err))
		}

		if !exists {
			return resource.RetryableError(fmt.Errorf("expected permission target to be created, but currently not found"))
		}

		return resource.NonRetryableError(resourcePermissionTargetRead(d, m))
	})
}

func resourcePermissionTargetRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	if _, ok := d.GetOk("repositories"); ok {
		return resourcePermissionTargetV1Read(d, m)
	}

	permissionTarget, resp, err := c.V2.Security.GetPermissionTarget(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	return packPermissionTarget(permissionTarget, d)
}

func resourcePermissionTargetUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	if _, ok := d.GetOk("repositories"); ok {
		return resourcePermissionTargetV1CreateOrReplace(d, m)
	}

	permissionTarget := unpackPermissionTarget(d)
	if _, err := c.V2.Security.UpdatePermissionTarget(context.Background(), d.Id(), permissionTarget); err != nil {
		return err
	}

	d.SetId(*permissionTarget.Name)
	return resourcePermissionTargetRead(d, m)
}

func resourcePermissionTargetDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	if _, ok := d.GetOk("repositories"); ok {
		return resourcePermissionTargetV1Delete(d, m)
	}

	permissionTarget := unpackPermissionTarget(d)
	resp, err := c.V2.Security.DeletePermissionTarget(context.Background(), *permissionTarget.Name)

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return err
}

func resourcePermissionTargetExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*artifactory.Artifactory)

	if _, ok := d.GetOk("repositories"); ok {
		_, resp, err := c.V1.Security.GetPermissionTargets(context.Background(), d.Id())

		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		} else if err != nil {
			return false, fmt.Errorf("error: Request failed: %s", err.Error())
		}
		return true, nil
	}

	return c.V2.Security.HasPermissionTarget(context.Background(), d.Id())
}
