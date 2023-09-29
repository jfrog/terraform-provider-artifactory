package security

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/security"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

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

func hashPrincipal(o interface{}) int {
	p := o.(map[string]interface{})
	part1 := schema.HashString(p["name"].(string)) + 31
	permissions := utilsdk.CastToStringArr(p["permissions"].(*schema.Set).List())
	part3 := schema.HashString(strings.Join(permissions, ""))
	return part1 * part3
}

func BuildPermissionTargetSchema() map[string]*schema.Schema {
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
							security.PermRead,
							security.PermAnnotate,
							security.PermWrite,
							security.PermDelete,
							security.PermManage,
							security.PermManagedXrayMeta,
							security.PermDistribute,
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
					Type: schema.TypeSet,
					Elem: &schema.Schema{Type: schema.TypeString},
					Set:  schema.HashString,
					Description: "You can specify the name `ANY` in the repositories section in order to apply to all repositories, " +
						"`ANY REMOTE` for all remote repositories and `ANY LOCAL` for all local repositories. The default value will be [] if nothing is specified.",
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
	buildSchema.Elem.(*schema.Resource).Schema["repositories"].Description = `This can only be 1 value: "artifactory-build-info", and currently, ` +
		"validation of sets/lists is not allowed. Artifactory will reject the request if you change this"

	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"repo":           &principalSchema,
		"build":          &buildSchema,
		"release_bundle": &principalSchema,
	}
}

func DataSourceArtifactoryPermissionTarget() *schema.Resource {
	dataSourcePermissionTargetRead := func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		permissionTarget := new(PermissionTargetParams)
		targetName := d.Get("name").(string)
		_, err := m.(utilsdk.ProvderMetadata).Client.R().SetResult(permissionTarget).Get(security.PermissionsEndPoint + targetName)

		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(permissionTarget.Name)

		return PackPermissionTarget(permissionTarget, d)
	}
	return &schema.Resource{
		ReadContext: dataSourcePermissionTargetRead,
		Schema:      BuildPermissionTargetSchema(),
		Description: "Provides the permission target data source. Contains information about a specific permission target.",
	}
}

func PackPermissionTarget(permissionTarget *PermissionTargetParams, d *schema.ResourceData) diag.Diagnostics {
	packPermission := func(p *PermissionTargetSection) []interface{} {
		packPermMap := func(e map[string][]string) []interface{} {
			perm := make([]interface{}, len(e))

			count := 0
			for k, v := range e {
				perm[count] = map[string]interface{}{
					"name":        k,
					"permissions": schema.NewSet(schema.HashString, utilsdk.CastToInterfaceArr(v)),
				}
				count++
			}

			return perm
		}

		s := map[string]interface{}{}

		if p != nil {
			if p.IncludePatterns != nil {
				s["includes_pattern"] = schema.NewSet(schema.HashString, utilsdk.CastToInterfaceArr(p.IncludePatterns))
			}

			if p.ExcludePatterns != nil {
				s["excludes_pattern"] = schema.NewSet(schema.HashString, utilsdk.CastToInterfaceArr(p.ExcludePatterns))
			}

			if p.Repositories != nil {
				s["repositories"] = schema.NewSet(schema.HashString, utilsdk.CastToInterfaceArr(p.Repositories))
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

	setValue := utilsdk.MkLens(d)

	errors := setValue("name", permissionTarget.Name)
	if permissionTarget.Repo != nil {
		errors = setValue("repo", packPermission(permissionTarget.Repo))
	}
	if permissionTarget.Build != nil {
		errors = setValue("build", packPermission(permissionTarget.Build))
	}

	if permissionTarget.ReleaseBundle != nil {
		errors = setValue("release_bundle", packPermission(permissionTarget.ReleaseBundle))
	}

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to marshal permission target %q", errors)
	}
	return nil
}
