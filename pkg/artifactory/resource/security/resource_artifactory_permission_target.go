package security

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

const permissionsEndPoint = "artifactory/api/v2/security/permissions/"
const (
	PermRead     = "read"
	PermWrite    = "write"
	PermAnnotate = "annotate"
	PermDelete   = "delete"
	PermManage   = "manage"
)

// PermissionTargetParams Copy from https://github.com/jfrog/jfrog-client-go/blob/master/artifactory/services/permissiontarget.go#L116
//
// Using struct pointers to keep the fields null if they are empty.
// Artifactory evaluates inner struct typed fields if they are not null, which can lead to failures in the request.
type PermissionTargetParams struct {
	Name          string                   `json:"name"`
	Repo          *PermissionTargetSection `json:"repo,omitempty"`
	Build         *PermissionTargetSection `json:"build,omitempty"`
	ReleaseBundle *PermissionTargetSection `json:"releaseBundle,omitempty"`
}

type PermissionTargetSection struct {
	IncludePatterns []string `json:"include-patterns,omitempty"`
	ExcludePatterns []string `json:"exclude-patterns,omitempty"`
	Repositories    []string `json:"repositories"`
	Actions         *Actions `json:"actions,omitempty"`
}

type Actions struct {
	Users  map[string][]string `json:"users,omitempty"`
	Groups map[string][]string `json:"groups,omitempty"`
}

func ResourceArtifactoryPermissionTargets() *schema.Resource {
	target := ResourceArtifactoryPermissionTarget()
	target.DeprecationMessage = "This resource has been deprecated in favour of artifactory_permission_target resource."
	return target
}

func ResourceArtifactoryPermissionTarget() *schema.Resource {
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
							PermRead,
							PermAnnotate,
							PermWrite,
							PermDelete,
							PermManage,
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
		CreateContext: resourcePermissionTargetCreate,
		ReadContext:   resourcePermissionTargetRead,
		UpdateContext: resourcePermissionTargetUpdate,
		DeleteContext: resourcePermissionTargetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
	permissions := util.CastToStringArr(p["permissions"].(*schema.Set).List())
	part3 := schema.HashString(strings.Join(permissions, ""))
	return part1 * part3
}

func unpackPermissionTarget(s *schema.ResourceData) *PermissionTargetParams {
	d := &util.ResourceData{ResourceData: s}

	unpackPermission := func(rawPermissionData interface{}) *PermissionTargetSection {
		unpackEntity := func(rawEntityData interface{}) *Actions {
			unpackPermMap := func(rawPermSet interface{}) map[string][]string {
				permList := rawPermSet.(*schema.Set).List()
				if len(permList) == 0 {
					return nil
				}

				permissions := make(map[string][]string)
				for _, v := range permList {
					id := v.(map[string]interface{})

					permissions[id["name"].(string)] = util.CastToStringArr(id["permissions"].(*schema.Set).List())
				}
				return permissions
			}

			entityDataList := rawEntityData.([]interface{})
			if len(entityDataList) == 0 {
				return nil
			}

			entityData := entityDataList[0].(map[string]interface{})
			return &Actions{
				Users:  unpackPermMap(entityData["users"]),
				Groups: unpackPermMap(entityData["groups"]),
			}
		}

		if rawPermissionData == nil || rawPermissionData.([]interface{})[0] == nil {
			return nil
		}

		// It is safe to unpack the zeroth element immediately since permission targets have min size of 1
		permissionData := rawPermissionData.([]interface{})[0].(map[string]interface{})

		permission := new(PermissionTargetSection)

		// This will always exist
		{
			tmp := util.CastToStringArr(permissionData["repositories"].(*schema.Set).List())
			permission.Repositories = tmp
		}

		// Handle optionals
		if v, ok := permissionData["includes_pattern"]; ok {
			// It is not possible to set default values for sets. Because the data type between moving from
			// atlassian to jfrog went from a *[]string to a []string, and both have json attributes of 'on empty omit'
			// when the * version was used, this would have cause an [] array to be sent, which artifactory would accept
			// now that the data type is changed, and [] is ommitted and so when artifactory see the key missing entirely
			// it responds with "[**]" which messes us the test. This hack seems to line them up
			tmp := util.CastToStringArr(v.(*schema.Set).List())
			if len(tmp) == 0 {
				tmp = []string{""}
			}
			permission.IncludePatterns = tmp
		}
		if v, ok := permissionData["excludes_pattern"]; ok {
			tmp := util.CastToStringArr(v.(*schema.Set).List())
			permission.ExcludePatterns = tmp
		}
		if v, ok := permissionData["actions"]; ok {
			permission.Actions = unpackEntity(v)
		}

		return permission
	}

	pTarget := new(PermissionTargetParams)

	pTarget.Name = d.GetString("name", false)

	if v, ok := d.GetOk("repo"); ok {
		pTarget.Repo = unpackPermission(v)
	}

	if v, ok := d.GetOk("build"); ok {
		pTarget.Build = unpackPermission(v)
	}

	return pTarget
}

func packPermissionTarget(permissionTarget *PermissionTargetParams, d *schema.ResourceData) diag.Diagnostics {
	packPermission := func(p *PermissionTargetSection) []interface{} {
		packPermMap := func(e map[string][]string) []interface{} {
			perm := make([]interface{}, len(e))

			count := 0
			for k, v := range e {
				perm[count] = map[string]interface{}{
					"name":        k,
					"permissions": schema.NewSet(schema.HashString, util.CastToInterfaceArr(v)),
				}
				count++
			}

			return perm
		}

		s := map[string]interface{}{}

		if p != nil {
			if p.IncludePatterns != nil {
				s["includes_pattern"] = schema.NewSet(schema.HashString, util.CastToInterfaceArr(p.IncludePatterns))
			}

			if p.ExcludePatterns != nil {
				s["excludes_pattern"] = schema.NewSet(schema.HashString, util.CastToInterfaceArr(p.ExcludePatterns))
			}

			if p.Repositories != nil {
				s["repositories"] = schema.NewSet(schema.HashString, util.CastToInterfaceArr(p.Repositories))
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

	setValue := util.MkLens(d)

	errors := setValue("name", permissionTarget.Name)
	if permissionTarget.Repo != nil {
		errors = setValue("repo", packPermission(permissionTarget.Repo))
	}
	if permissionTarget.Build != nil {
		errors = setValue("build", packPermission(permissionTarget.Build))
	}

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to marshal permission target %q", errors)
	}
	return nil
}

func resourcePermissionTargetCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	permissionTarget := unpackPermissionTarget(d)

	if _, err := m.(*resty.Client).R().AddRetryCondition(repository.Retry400).SetBody(permissionTarget).Post(permissionsEndPoint + permissionTarget.Name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(permissionTarget.Name)
	return nil
}

func resourcePermissionTargetRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	permissionTarget := new(PermissionTargetParams)
	resp, err := m.(*resty.Client).R().SetResult(permissionTarget).Get(permissionsEndPoint + d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return packPermissionTarget(permissionTarget, d)
}

func resourcePermissionTargetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	permissionTarget := unpackPermissionTarget(d)

	if _, err := m.(*resty.Client).R().SetBody(permissionTarget).Put(permissionsEndPoint + d.Id()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(permissionTarget.Name)
	return resourcePermissionTargetRead(ctx, d, m)
}

func resourcePermissionTargetDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, err := m.(*resty.Client).R().Delete(permissionsEndPoint + d.Id())

	return diag.FromErr(err)
}

func PermTargetExists(id string, m interface{}) (bool, error) {
	resp, err := m.(*resty.Client).R().Head(permissionsEndPoint + id)
	if err != nil && resp != nil && resp.StatusCode() == http.StatusNotFound {
		// Do not error on 404s as this causes errors when the upstream permission has been manually removed
		return false, nil
	}

	return err == nil, err
}
