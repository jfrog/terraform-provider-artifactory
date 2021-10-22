package artifactory

import (
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// WatchGeneralData this struct and all the others below it line up identically with the
// structs from the V2 go client from jfrog with one fatal exception: None of these nested types is exported
// and it's totally inconsistent with the rest of the code.
// Option are: move this code into the terraform space, as is, or beg the jfrog-go-client
// team to captial case those variables. I ticket will be filed, but I am not hopeful.
type WatchGeneralData struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}

type WatchFilterValue struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

// WatchFilterValueWrapper is a wrapper around WatchFilterValue which handles the API returning both a string and an object for the watch filter value
type WatchFilterValueWrapper struct {
	WatchFilterValue
	IsPropertyFilter bool `json:”-”`
}

type WatchFilter struct {
	Type  *string                  `json:"type,omitempty"`
	Value *WatchFilterValueWrapper `json:"value,omitempty"`
}

type WatchProjectResource struct {
	Type            *string        `json:"type,omitempty"`
	RepoType        *string        `json:"repo_type,omitempty"`
	BinaryManagerId *string        `json:"bin_mgr_id,omitempty"`
	Name            *string        `json:"name,omitempty"`
	Filters         *[]WatchFilter `json:"filters,omitempty"`
}

type WatchProjectResources struct {
	Resources *[]WatchProjectResource `json:"resources,omitempty"`
}

type WatchAssignedPolicy struct {
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
}

type Watch struct {
	GeneralData      *WatchGeneralData      `json:"general_data,omitempty"`
	ProjectResources *WatchProjectResources `json:"project_resources,omitempty"`
	AssignedPolicies *[]WatchAssignedPolicy `json:"assigned_policies,omitempty"`
}

func resourceXrayWatch() *schema.Resource {
	return &schema.Resource{
		Create: resourceXrayWatchCreate,
		Read:   resourceXrayWatchRead,
		Update: resourceXrayWatchUpdate,
		Delete: resourceXrayWatchDelete,
		DeprecationMessage: "This portion of the provider uses V1 apis and will eventually be moved " +
			"to the separate repo. The discussion is here: https://github.com/jfrog/terraform-provider-artifactory/issues/160",
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

func expandWatch(d *schema.ResourceData) *Watch {
	watch := new(Watch)

	gd := &WatchGeneralData{
		Name: StringPtr(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		gd.Description = StringPtr(v.(string))
	}
	if v, ok := d.GetOk("active"); ok {
		gd.Active = BoolPtr(v.(bool))
	}
	watch.GeneralData = gd

	pr := &WatchProjectResources{}
	if v, ok := d.GetOk("resources"); ok {
		r := &[]WatchProjectResource{}
		for _, res := range v.([]interface{}) {
			*r = append(*r, *expandProjectResource(res))
		}
		pr.Resources = r
	}
	watch.ProjectResources = pr

	ap := &[]WatchAssignedPolicy{}
	if v, ok := d.GetOk("assigned_policies"); ok {
		for _, pol := range v.([]interface{}) {
			*ap = append(*ap, *expandAssignedPolicy(pol))
		}
	}
	watch.AssignedPolicies = ap

	return watch
}

func expandProjectResource(rawCfg interface{}) *WatchProjectResource {
	resource := new(WatchProjectResource)

	cfg := rawCfg.(map[string]interface{})
	resource.Type = StringPtr(cfg["type"].(string))
	if v, ok := cfg["bin_mgr_id"]; ok {
		resource.BinaryManagerId = StringPtr(v.(string))
	}
	if v, ok := cfg["repo_type"]; ok {
		resource.RepoType = StringPtr(v.(string))
	}
	if v, ok := cfg["name"]; ok {
		resource.Name = StringPtr(v.(string))
	}
	if v, ok := cfg["filters"]; ok {
		resourceFilters := expandFilters(v.([]interface{}))
		resource.Filters = &resourceFilters
	}

	return resource
}

func expandFilters(l []interface{}) []WatchFilter {
	filters := make([]WatchFilter, 0, len(l))

	for _, raw := range l {
		filter := new(WatchFilter)
		f := raw.(map[string]interface{})
		filter.Type = StringPtr(f["type"].(string))
		valueWrapper := new(WatchFilterValueWrapper)
		fv := new(WatchFilterValue)
		fv.Value = StringPtr(f["value"].(string))
		valueWrapper.WatchFilterValue = *fv
		filter.Value = valueWrapper

		filters = append(filters, *filter)
	}

	return filters
}

func expandAssignedPolicy(rawCfg interface{}) *WatchAssignedPolicy {
	policy := new(WatchAssignedPolicy)

	cfg := rawCfg.(map[string]interface{})
	policy.Name = StringPtr(cfg["name"].(string))
	policy.Type = StringPtr(cfg["type"].(string))

	return policy
}

func flattenProjectResources(resources *WatchProjectResources) []interface{} {
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

func flattenFilters(filters *[]WatchFilter) []interface{} {
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

func flattenAssignedPolicies(policies *[]WatchAssignedPolicy) []interface{} {
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

	watch := expandWatch(d)
	_, err := m.(*resty.Client).R().SetBody(&watch).Post("xray/api/v2/watches")
	if err != nil {
		return err
	}

	d.SetId(*watch.GeneralData.Name) // ID may be returned according to the API docs, but not in go-xray
	return resourceXrayWatchRead(d, m)
}

func resourceXrayWatchRead(d *schema.ResourceData, m interface{}) error {
	watch := Watch{}
	resp, err := m.(*resty.Client).R().SetResult(&watch).Get("xray/api/v2/watches/" + d.Id())
	if err != nil {

		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			log.Printf("[WARN] Xray watch (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
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
	watch := expandWatch(d)
	_, err := m.(*resty.Client).R().SetBody(&watch).Put("xray/api/v2/watches/" + d.Id())
	if err != nil {
		return err
	}

	d.SetId(*watch.GeneralData.Name)
	return resourceXrayWatchRead(d, m)
}

func resourceXrayWatchDelete(d *schema.ResourceData, m interface{}) error {
	_, err := m.(*resty.Client).R().Delete("xray/api/v2/watches/" + d.Id())
	return err
}
