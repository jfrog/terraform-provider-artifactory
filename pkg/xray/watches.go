package xray

import (
	"context"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/go-resty/resty/v2"
)

type WatchGeneralData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

type WatchFilterValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type WatchFilter struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type WatchProjectResource struct {
	Type            string        `json:"type"`
	BinaryManagerId string        `json:"bin_mgr_id"`
	Filters         []WatchFilter `json:"filters"`
	Name            string        `json:"name"`
}

type WatchProjectResources struct {
	Resources []WatchProjectResource `json:"resources"`
}

type WatchAssignedPolicy struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Watch struct {
	GeneralData      WatchGeneralData      `json:"general_data"`
	ProjectResources WatchProjectResources `json:"project_resources"`
	AssignedPolicies []WatchAssignedPolicy `json:"assigned_policies"`
	WatchRecipients  []string              `json:"watch_recipients"`
}

func unpackWatch(d *schema.ResourceData) Watch {
	watch := Watch{}

	gd := WatchGeneralData{
		Name: d.Get("name").(string),
	}
	if v, ok := d.GetOk("description"); ok {
		gd.Description = v.(string)
	}
	if v, ok := d.GetOk("active"); ok {
		gd.Active = v.(bool)
	}
	watch.GeneralData = gd

	pr := WatchProjectResources{}
	if v, ok := d.GetOk("watch_resource"); ok {
		var r []WatchProjectResource
		for _, res := range v.(*schema.Set).List() {
			r = append(r, unpackProjectResource(res))
		}
		pr.Resources = r
	}
	watch.ProjectResources = pr

	var ap []WatchAssignedPolicy
	if v, ok := d.GetOk("assigned_policy"); ok {
		policies := v.(*schema.Set).List()
		for _, pol := range policies {
			ap = append(ap, unpackAssignedPolicy(pol))
		}
	}
	watch.AssignedPolicies = ap

	var watchRecipients []string

	if v, ok := d.GetOk("watch_recipients"); ok {
		recipients := v.(*schema.Set).List()
		for _, watchRec := range recipients {
			watchRecipients = append(watchRecipients, watchRec.(string))
		}
	}
	watch.WatchRecipients = watchRecipients

	return watch
}

func unpackProjectResource(rawCfg interface{}) WatchProjectResource {
	resource := WatchProjectResource{}

	cfg := rawCfg.(map[string]interface{})
	resource.Type = cfg["type"].(string)

	if v, ok := cfg["bin_mgr_id"]; ok {
		resource.BinaryManagerId = v.(string)
	}
	if v, ok := cfg["name"]; ok {
		resource.Name = v.(string)
	}

	if v, ok := cfg["filter"]; ok {
		resourceFilters := unpackFilters(v.([]interface{}))
		resource.Filters = resourceFilters
	}

	return resource
}

func unpackFilters(list []interface{}) []WatchFilter {
	var filters []WatchFilter

	for _, raw := range list {
		filter := WatchFilter{}
		f := raw.(map[string]interface{})
		filter.Type = f["type"].(string)
		filter.Value = f["value"].(string)
		filters = append(filters, filter)
	}

	return filters
}

func unpackAssignedPolicy(rawCfg interface{}) WatchAssignedPolicy {
	policy := WatchAssignedPolicy{}

	cfg := rawCfg.(map[string]interface{})
	policy.Name = cfg["name"].(string)
	policy.Type = cfg["type"].(string)

	return policy
}

func packProjectResources(resources WatchProjectResources) []interface{} {
	var list []interface{}
	for _, res := range resources.Resources {
		resourceMap := map[string]interface{}{}
		resourceMap["type"] = res.Type
		if len(res.Name) > 0 {
			resourceMap["name"] = res.Name
		}
		if len(res.BinaryManagerId) > 0 {
			resourceMap["bin_mgr_id"] = res.BinaryManagerId
		}
		resourceMap["filter"] = packFilters(res.Filters)
		list = append(list, resourceMap)
	}

	return list
}

func packFilters(filters []WatchFilter) []interface{} {
	var l []interface{}
	for _, f := range filters {
		m := map[string]interface{}{
			"type":  f.Type,
			"value": f.Value,
		}
		l = append(l, m)
	}

	return l
}

func packAssignedPolicies(policies []WatchAssignedPolicy) []interface{} {
	var l []interface{}
	for _, p := range policies {
		m := make(map[string]interface{})
		m["name"] = p.Name
		m["type"] = p.Type
		l = append(l, m)
	}

	return l
}

func packWatch(watch Watch, d *schema.ResourceData) diag.Diagnostics {
	if err := d.Set("description", watch.GeneralData.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", watch.GeneralData.Active); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("watch_resource", packProjectResources(watch.ProjectResources)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("assigned_policy", packAssignedPolicies(watch.AssignedPolicies)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func getWatch(id string, client *resty.Client) (Watch, *resty.Response, error) {
	watch := Watch{}
	resp, err := client.R().SetResult(&watch).Get("xray/api/v2/watches/" + id)
	return watch, resp, err
}

func resourceXrayWatchCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	watch := unpackWatch(d)
	_, err := m.(*resty.Client).R().SetBody(watch).Post("xray/api/v2/watches")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(watch.GeneralData.Name)
	return resourceXrayWatchRead(ctx, d, m)
}

func resourceXrayWatchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	watch, resp, err := getWatch(d.Id(), m.(*resty.Client))
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			log.Printf("[WARN] Xray watch (%s) not found, removing from state", d.Id())
			d.SetId("")
		}
		return diag.FromErr(err)
	}
	packWatch(watch, d)
	return nil
}

func resourceXrayWatchUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	watch := unpackWatch(d)
	resp, err := m.(*resty.Client).R().SetBody(watch).Put("xray/api/v2/watches/" + d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			log.Printf("[WARN] Xray watch (%s) not found, removing from state", d.Id())
			d.SetId("")
		}
		return diag.FromErr(err)
	}

	d.SetId(watch.GeneralData.Name)
	return resourceXrayWatchRead(ctx, d, m)
}

func resourceXrayWatchDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resp, err := m.(*resty.Client).R().Delete("xray/api/v2/watches/" + d.Id())
	if err != nil && resp.StatusCode() == http.StatusNotFound {
		d.SetId("")
		return diag.FromErr(err)
	}
	return nil
}
