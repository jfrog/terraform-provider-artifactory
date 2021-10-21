package artifactory

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/jfrog/jfrog-client-go/artifactory/services"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const permissionsEndPoint = "artifactory/api/v2/security/permissions/"
const (
	PERM_READ     = "read"
	PERM_WRITE    = "write"
	PERM_ANNOTATE = "annotate"
	PERM_DELETE   = "delete"
	PERM_MANAGE   = "manage"

	PERMISSION_SCHEMA = "application/vnd.org.jfrog.artifactory.security.PermissionTargetV2+json"
)

func resourceArtifactoryPermissionTargets() *schema.Resource {
	target := resourceArtifactoryPermissionTarget()
	return target
}

func resourceArtifactoryPermissionTarget() *schema.Resource {
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
							PERM_READ,
							PERM_ANNOTATE,
							PERM_WRITE,
							PERM_DELETE,
							PERM_MANAGE,
							// v2.PERM_MANAGE_XRAY_METADATA,
							"managedXrayMeta",
						}, false),
					},
					Set:      schema.HashString,
					Required: true,
				},
			},
		},
	}

	principalSchema := schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"includes_pattern": {
					Type:        schema.TypeSet,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Set:         schema.HashString,
					Optional:    true,
					Description: `The default value will be [""] if nothing is supplied`,
				},
				"excludes_pattern": {
					Type:        schema.TypeSet,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Set:         schema.HashString,
					Optional:    true,
					Description: `The default value will be [] if nothing is supplied`,
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
	buildSchema := principalSchema
	buildSchema.Elem.(*schema.Resource).Schema["repositories"].Description = `This can only be 1 value: "artifactory-build-info", and currently, validation of sets/lists is not allowed. Artifactory will reject the request if you change this`

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
			"repo":           &principalSchema,
			"build":          &buildSchema,
			"release_bundle": &principalSchema,
		},
	}
}

func hashPrincipal(o interface{}) int {
	p := o.(map[string]interface{})
	part1 := schema.HashString(p["name"].(string)) + 31
	permissions := castToStringArr(p["permissions"].(*schema.Set).List())
	part3 := schema.HashString(strings.Join(permissions, ""))
	return part1 * part3
}

func unpackPermissionTarget(s *schema.ResourceData) *services.PermissionTargetParams {
	d := &ResourceData{s}

	unpackPermission := func(rawPermissionData interface{}) *services.PermissionTargetSection {
		unpackEntity := func(rawEntityData interface{}) *services.Actions {
			unpackPermMap := func(rawPermSet interface{}) map[string][]string {
				permList := rawPermSet.(*schema.Set).List()
				if len(permList) == 0 {
					return nil
				}

				permissions := make(map[string][]string)
				for _, v := range permList {
					id := v.(map[string]interface{})

					permissions[id["name"].(string)] = castToStringArr(id["permissions"].(*schema.Set).List())
				}
				return permissions
			}

			entityDataList := rawEntityData.([]interface{})
			if len(entityDataList) == 0 {
				return nil
			}

			entityData := entityDataList[0].(map[string]interface{})
			return &services.Actions{
				Users:  unpackPermMap(entityData["users"]),
				Groups: unpackPermMap(entityData["groups"]),
			}
		}

		if rawPermissionData == nil || rawPermissionData.([]interface{})[0] == nil {
			return nil
		}

		// It is safe to unpack the zeroth element immediately since permission targets have min size of 1
		permissionData := rawPermissionData.([]interface{})[0].(map[string]interface{})

		permission := new(services.PermissionTargetSection)

		// This will always exist
		{
			tmp := castToStringArr(permissionData["repositories"].(*schema.Set).List())
			permission.Repositories = tmp
		}

		// Handle optionals
		if v, ok := permissionData["includes_pattern"]; ok {
			// It is not possible to set default values for sets. Because the data type between moving from
			// atlassian to jfrog went from a *[]string to a []string, and both have json attributes of 'on empty omit'
			// when the * version was used, this would have cause an [] array to be sent, which artifactory would accept
			// now that the data type is changed, and [] is ommitted and so when artifactory see the key missing entirely
			// it responds with "[**]" which messes us the test. This hack seems to line them up
			tmp := castToStringArr(v.(*schema.Set).List())
			if len(tmp) == 0 {
				tmp = []string{""}
			}
			permission.IncludePatterns = tmp
		}
		if v, ok := permissionData["excludes_pattern"]; ok {
			tmp := castToStringArr(v.(*schema.Set).List())
			permission.ExcludePatterns = tmp
		}
		if v, ok := permissionData["actions"]; ok {
			permission.Actions = unpackEntity(v)
		}

		return permission
	}

	pTarget := new(services.PermissionTargetParams)

	pTarget.Name = d.getString("name", false)

	if v, ok := d.GetOk("repo"); ok {
		pTarget.Repo = unpackPermission(v)
	}

	if v, ok := d.GetOk("build"); ok {
		pTarget.Build = unpackPermission(v)
	}

	return pTarget
}

func packPermissionTarget(permissionTarget *services.PermissionTargetParams, d *schema.ResourceData) error {
	packPermission := func(p *services.PermissionTargetSection) []interface{} {
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
				s["includes_pattern"] = schema.NewSet(schema.HashString, castToInterfaceArr(p.IncludePatterns))
			}

			if p.ExcludePatterns != nil {
				s["excludes_pattern"] = schema.NewSet(schema.HashString, castToInterfaceArr(p.ExcludePatterns))
			}

			if p.Repositories != nil {
				s["repositories"] = schema.NewSet(schema.HashString, castToInterfaceArr(p.Repositories))
			}

			if p.Actions != nil {
				perms := make(map[string]interface{})

				if p.Actions.Users != nil {
					perms["users"] = schema.NewSet(hashPrincipal, packPermMap(p.Actions.Users))
				}

				if p.Actions.Groups != nil {
					perms["groups"] = schema.NewSet(hashPrincipal, packPermMap(p.Actions.Groups))
				}

				if len(perms) > 0 {
					s["actions"] = []interface{}{perms}
				}
			}
		}

		return []interface{}{s}
	}

	setValue := mkLens(d)

	errors := setValue("name", permissionTarget.Name)
	if permissionTarget.Repo != nil {
		errors = setValue("repo", packPermission(permissionTarget.Repo))
	}
	if permissionTarget.Build != nil {
		errors = setValue("build", packPermission(permissionTarget.Build))
	}

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to marshal permission target %q", errors)
	}
	return nil
}

func resourcePermissionTargetCreate(d *schema.ResourceData, m interface{}) error {
	permissionTarget := unpackPermissionTarget(d)

	if _, err := m.(*resty.Client).R().SetBody(permissionTarget).Post(permissionsEndPoint + permissionTarget.Name); err != nil {
		return err
	}

	d.SetId(permissionTarget.Name)
	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		exists, err := resourcePermissionTargetExists(d, m)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error describing permssions target: %s", err))
		}

		if !exists {
			return resource.RetryableError(fmt.Errorf("expected permission target to be created, but currently not found"))
		}

		return nil
	})
}

func resourcePermissionTargetRead(d *schema.ResourceData, m interface{}) error {
	permissionTarget := new(services.PermissionTargetParams)
	resp, err := m.(*resty.Client).R().SetResult(permissionTarget).Get(permissionsEndPoint + d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return err
	}

	return packPermissionTarget(permissionTarget, d)
}

func resourcePermissionTargetUpdate(d *schema.ResourceData, m interface{}) error {
	permissionTarget := unpackPermissionTarget(d)

	if _, err := m.(*resty.Client).R().SetBody(permissionTarget).Put(permissionsEndPoint + d.Id()); err != nil {
		return err
	}

	d.SetId(permissionTarget.Name)
	return resourcePermissionTargetRead(d, m)
}

func resourcePermissionTargetDelete(d *schema.ResourceData, m interface{}) error {
	_, err := m.(*resty.Client).R().Delete(permissionsEndPoint + d.Id())

	return err
}

func resourcePermissionTargetExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return permTargetExists(d.Id(), m)
}

func permTargetExists(id string, m interface{}) (bool, error) {
	_, err := m.(*resty.Client).R().Head(permissionsEndPoint + id)

	return err == nil, err
}
