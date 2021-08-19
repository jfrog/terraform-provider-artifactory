package artifactory

import (
	"context"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jasonwbarnett/go-xray/xray"
	v2 "github.com/jasonwbarnett/go-xray/xray/v2"
)

func resourceXrayWatch() *schema.Resource {
	return &schema.Resource{
		Create: resourceXrayWatchCreate,
		Read:   resourceXrayWatchRead,
		Update: resourceXrayWatchUpdate,
		Delete: resourceXrayWatchDelete,

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
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"resources": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bin_mgr_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"repo_type": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"local",
								"remote",
							}, false),
						},
						"filters": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									// TODO this can be either a string or possibly a json blob
									// eg "value":{"ExcludePatterns":[],"IncludePatterns":["*"]}
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"assigned_policies": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"watch_recipients": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func expandWatch(d *schema.ResourceData) *v2.Watch {
	watch := new(v2.Watch)

	gd := &v2.WatchGeneralData{
		Name: xray.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		gd.Description = xray.String(v.(string))
	}
	if v, ok := d.GetOk("active"); ok {
		gd.Active = xray.Bool(v.(bool))
	}
	watch.GeneralData = gd

	pr := &v2.WatchProjectResources{}
	if v, ok := d.GetOk("resources"); ok {
		r := &[]v2.WatchProjectResource{}
		for _, res := range v.([]interface{}) {
			*r = append(*r, *expandProjectResource(res))
		}
		pr.Resources = r
	}
	watch.ProjectResources = pr

	ap := &[]v2.WatchAssignedPolicy{}
	if v, ok := d.GetOk("assigned_policies"); ok {
		for _, pol := range v.([]interface{}) {
			*ap = append(*ap, *expandAssignedPolicy(pol))
		}
	}
	watch.AssignedPolicies = ap

	return watch
}

func expandProjectResource(rawCfg interface{}) *v2.WatchProjectResource {
	resource := new(v2.WatchProjectResource)

	cfg := rawCfg.(map[string]interface{})
	resource.Type = xray.String(cfg["type"].(string))
	if v, ok := cfg["bin_mgr_id"]; ok {
		resource.BinaryManagerId = xray.String(v.(string))
	}
	if v, ok := cfg["repo_type"]; ok {
		resource.RepoType = xray.String(v.(string))
	}
	if v, ok := cfg["name"]; ok {
		resource.Name = xray.String(v.(string))
	}
	if v, ok := cfg["filters"]; ok {
		resourceFilters := expandFilters(v.([]interface{}))
		resource.Filters = &resourceFilters
	}

	return resource
}

func expandFilters(l []interface{}) []v2.WatchFilter {
	filters := make([]v2.WatchFilter, 0, len(l))

	for _, raw := range l {
		filter := new(v2.WatchFilter)
		f := raw.(map[string]interface{})
		filter.Type = xray.String(f["type"].(string))
		valueWrapper := new(v2.WatchFilterValueWrapper)
		fv := new(v2.WatchFilterValue)
		fv.Value = xray.String(f["value"].(string))
		valueWrapper.WatchFilterValue = *fv
		filter.Value = valueWrapper

		filters = append(filters, *filter)
	}

	return filters
}

func expandAssignedPolicy(rawCfg interface{}) *v2.WatchAssignedPolicy {
	policy := new(v2.WatchAssignedPolicy)

	cfg := rawCfg.(map[string]interface{})
	policy.Name = xray.String(cfg["name"].(string))
	policy.Type = xray.String(cfg["type"].(string))

	return policy
}

func flattenProjectResources(resources *v2.WatchProjectResources) []interface{} {
	if resources == nil || resources.Resources == nil {
		return []interface{}{}
	}

	var l []interface{}
	for _, res := range *resources.Resources {
		m := make(map[string]interface{})
		m["type"] = res.Type
		if res.Name != nil {
			m["name"] = res.Name
		}
		if res.BinaryManagerId != nil {
			m["bin_mgr_id"] = res.BinaryManagerId
		}
		if res.RepoType != nil {
			m["repo_type"] = res.RepoType
		}
		m["filters"] = flattenFilters(res.Filters)
		l = append(l, m)
	}

	return l
}

func flattenFilters(filters *[]v2.WatchFilter) []interface{} {
	if filters == nil {
		return []interface{}{}
	}

	var l []interface{}
	for _, f := range *filters {
		m := make(map[string]interface{})
		m["type"] = f.Type
		m["value"] = f.Value.WatchFilterValue.Value
		l = append(l, m)
	}

	return l
}

func flattenAssignedPolicies(policies *[]v2.WatchAssignedPolicy) []interface{} {
	if policies == nil {
		return []interface{}{}
	}

	var l []interface{}
	for _, p := range *policies {
		m := make(map[string]interface{})
		m["name"] = p.Name
		m["type"] = p.Type
		l = append(l, m)
	}

	return l
}

func resourceXrayWatchCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Xray

	watch := expandWatch(d)

	_, err := c.V2.Watches.CreateWatch(context.Background(), watch)
	if err != nil {
		return err
	}

	d.SetId(*watch.GeneralData.Name) // ID may be returned according to the API docs, but not in go-xray
	return resourceXrayWatchRead(d, m)
}

func resourceXrayWatchRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Xray

	watch, resp, err := c.V2.Watches.GetWatch(context.Background(), d.Id())

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Xray watch (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("description", watch.GeneralData.Description); err != nil {
		return err
	}
	if err := d.Set("active", watch.GeneralData.Active); err != nil {
		return err
	}
	if err := d.Set("resources", flattenProjectResources(watch.ProjectResources)); err != nil {
		return err
	}
	if err := d.Set("assigned_policies", flattenAssignedPolicies(watch.AssignedPolicies)); err != nil {
		return err
	}

	return nil
}

func resourceXrayWatchUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Xray

	watch := expandWatch(d)
	_, err := c.V2.Watches.UpdateWatch(context.Background(), d.Id(), watch)
	if err != nil {
		return err
	}

	d.SetId(*watch.GeneralData.Name)
	return resourceXrayWatchRead(d, m)
}

func resourceXrayWatchDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Xray

	_, err := c.V2.Watches.DeleteWatch(context.Background(), d.Id())

	return err
}
