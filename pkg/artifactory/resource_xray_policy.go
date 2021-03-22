package artifactory

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/xero-oss/go-xray/xray"
	v1 "github.com/xero-oss/go-xray/xray/v1"
)

func resourceXrayPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceXrayPolicyCreate,
		Read:   resourceXrayPolicyRead,
		Update: resourceXrayPolicyUpdate,
		Delete: resourceXrayPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"author": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"rules": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"criteria": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// Security criteria
									"min_severity": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"rules.criteria.0.allow_unknown", "rules.criteria.0.banned_licenses", "rules.criteria.0.allowed_licenses", "rules.criteria.0.cvss_range"},
									},
									"cvss_range": {
										Type:          schema.TypeList,
										Optional:      true,
										ConflictsWith: []string{"rules.criteria.0.allow_unknown", "rules.criteria.0.banned_licenses", "rules.criteria.0.allowed_licenses", "rules.criteria.0.min_severity"},
										MaxItems:      1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"from": {
													Type:     schema.TypeInt, // Yes, the xray web ui allows floats. The go library says ints. :(
													Required: true,
												},
												"to": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									// License Criteria
									"allow_unknown": {
										Type:          schema.TypeBool,
										Optional:      true,
										ConflictsWith: []string{"rules.criteria.0.min_severity", "rules.criteria.0.cvss_range"},
									},
									"banned_licenses": {
										Type:          schema.TypeList,
										Optional:      true,
										ConflictsWith: []string{"rules.criteria.0.min_severity", "rules.criteria.0.cvss_range"},
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"allowed_licenses": {
										Type:          schema.TypeList,
										Optional:      true,
										ConflictsWith: []string{"rules.criteria.0.min_severity", "rules.criteria.0.cvss_range"},
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"actions": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mails": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"fail_build": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"block_download": {
										Type:     schema.TypeList,
										Required: true,
										// TODO: In an ideal world, this would be optional (see note in expandActions)
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"unscanned": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"active": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
									"webhooks": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"custom_severity": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func expandPolicy(d *schema.ResourceData) *v1.Policy {
	policy := new(v1.Policy)

	policy.Name = xray.String(d.Get("name").(string))
	if v, ok := d.GetOk("type"); ok {
		policy.Type = xray.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		policy.Description = xray.String(v.(string))
	}
	if v, ok := d.GetOk("author"); ok {
		policy.Author = xray.String(v.(string))
	}
	policyRules := expandRules(d.Get("rules").([]interface{}))
	policy.Rules = &policyRules

	return policy
}

func expandRules(configured []interface{}) []v1.PolicyRule {
	rules := make([]v1.PolicyRule, 0, len(configured))

	for _, raw := range configured {
		rule := new(v1.PolicyRule)
		data := raw.(map[string]interface{})
		rule.Name = xray.String(data["name"].(string))
		rule.Priority = xray.Int(data["priority"].(int))

		rule.Criteria = expandCriteria(data["criteria"].([]interface{}))
		if v, ok := data["actions"]; ok {
			rule.Actions = expandActions(v.([]interface{}))
		}
		rules = append(rules, *rule)
	}

	return rules
}

func expandCriteria(l []interface{}) *v1.PolicyRuleCriteria {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{}) // We made this a list of one to make schema validation easier
	criteria := new(v1.PolicyRuleCriteria)

	// The API doesn't allow both severity and license criteria to be _set_, even if they have empty values
	// So we have to figure out which group is actually empty and not even set it
	minSev := xray.String(m["min_severity"].(string))
	cvss := expandCVSSRange(m["cvss_range"].([]interface{}))
	allowUnk := xray.Bool(m["allow_unknown"].(bool))
	banned := expandLicenses(m["banned_licenses"].([]interface{}))
	allowed := expandLicenses(m["allowed_licenses"].([]interface{}))

	if *minSev == "" && cvss == nil {
		// If these are both the default values, we must be using license criteria
		criteria.AllowUnkown = allowUnk // "Unkown" is a typo in xray-oss
		criteria.BannedLicenses = banned
		criteria.AllowedLicenses = allowed
	} else {
		// This is also picky about not allowing empty values to be set
		if cvss == nil {
			criteria.MinimumSeverity = minSev
		} else {
			criteria.CVSSRange = cvss
		}
	}

	return criteria
}

func expandCVSSRange(l []interface{}) *v1.PolicyCVSSRange {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})
	cvssrange := &v1.PolicyCVSSRange{
		From:   xray.Int(m["from"].(int)),
		To: xray.Int(m["to"].(int)),
	}
	return cvssrange
}

func expandLicenses(l []interface{}) *[]string {
	if len(l) == 0 {
		return nil
	}

	licenses := make([]string, 0, len(l))
	for _, license := range l {
		licenses = append(licenses, license.(string))
	}
	return &licenses
}

func expandActions(l []interface{}) *v1.PolicyRuleActions {
	if len(l) == 0 {
		return nil
	}
	
	actions := new(v1.PolicyRuleActions)
	m := l[0].(map[string]interface{}) // We made this a list of one to make schema validation easier

	if v, ok := m["mails"]; ok && len(v.([]interface{})) > 0 {
		m := v.([]interface{})
		mails := make([]string, 0, len(m))
		for _, mail := range m {
			mails = append(mails, mail.(string))
		}
		actions.Mails = &mails
	}
	if v, ok := m["fail_build"]; ok {
		actions.FailBuild = xray.Bool(v.(bool))
	}

	if v, ok := m["block_download"]; ok {
		if len(v.([]interface{})) > 0 {
			vList := v.([]interface{})
			vMap := vList[0].(map[string]interface{})

			actions.BlockDownload = &v1.BlockDownloadSettings{
				Unscanned: xray.Bool(vMap["unscanned"].(bool)),
				Active:    xray.Bool(vMap["active"].(bool)),
			}
		} else {
			actions.BlockDownload = &v1.BlockDownloadSettings{
				Unscanned: xray.Bool(false),
				Active:    xray.Bool(false),
			}
			// Setting this false/false block feels like it _should_ work, since putting a false/false block in the terraform resource works fine
			// However, it doesn't, and we end up getting this diff when running acceptance tests when this is optional in the schema
			// rules.0.actions.0.block_download.#:           "1" => "0"
			// rules.0.actions.0.block_download.0.active:    "false" => ""
			// rules.0.actions.0.block_download.0.unscanned: "false" => ""
		}
	}

	if v, ok := m["webhooks"]; ok && len(v.([]interface{})) > 0 {
		m := v.([]interface{})
		webhooks := make([]string, 0, len(m))
		for _, hook := range m {
			webhooks = append(webhooks, hook.(string))
		}
		actions.Webhooks = &webhooks
	}
	if v, ok := m["custom_severity"]; ok {
		actions.CustomSeverity = xray.String(v.(string))
	}

	return actions
}

func flattenRules(rules []v1.PolicyRule) []interface{} {
	l := make([]interface{}, len(rules))

	for i, rule := range rules {
		m := map[string]interface{}{
			"name": *rule.Name,
			"priority": *rule.Priority,
			"criteria": flattenCriteria(rule.Criteria),
			"actions": flattenActions(rule.Actions),
		}
		l[i] = m
	}

	return l
}

func flattenCriteria(criteria *v1.PolicyRuleCriteria) []interface{} {
	if criteria == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cvss_range": flattenCVSSRange(criteria.CVSSRange),
	}

	if criteria.MinimumSeverity != nil {
		m["min_severity"] = *criteria.MinimumSeverity
	}
	if criteria.AllowUnkown != nil {
		m["allow_unknown"] = *criteria.AllowUnkown // Same typo in the library
	}
	if criteria.BannedLicenses != nil {
		m["banned_licenses"] = *criteria.BannedLicenses
	}
	if criteria.AllowedLicenses != nil {
		m["allowed_licenses"] = *criteria.AllowedLicenses
	}

	return []interface{}{m}
}

func flattenCVSSRange(cvss *v1.PolicyCVSSRange) []interface{} {
	if cvss == nil {
		return []interface{}{}
	}
	
	m := map[string]interface{}{
		"from": *cvss.From,
		"to": *cvss.To,
	}
	return []interface{}{m}
}

func flattenActions(actions *v1.PolicyRuleActions) []interface{} {
	if actions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"block_download": flattenBlockDownload(actions.BlockDownload),
	}

	if actions.Mails != nil {
		m["mails"] = *actions.Mails
	}
	if actions.FailBuild != nil {
		m["fail_build"] = *actions.FailBuild
	}
	if actions.Webhooks != nil {
		m["webhooks"] = *actions.Webhooks
	}
	if actions.CustomSeverity != nil {
		m["custom_severity"] = *actions.CustomSeverity
	}

	return []interface{}{m}
}

func flattenBlockDownload(bd *v1.BlockDownloadSettings) []interface{} {
	if bd == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}
	if bd.Unscanned != nil {
		m["unscanned"] = *bd.Unscanned
	}
	if bd.Active != nil {
		m["active"] = *bd.Active
	}

	return []interface{}{m}
}

func resourceXrayPolicyCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Xray

	policy := expandPolicy(d)
	resp, err := c.V1.Policies.CreatePolicy(context.Background(), policy)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status code when creating resource: %d", resp.StatusCode)
	}

	d.SetId(*policy.Name)
	return resourceXrayPolicyRead(d, m)
}

func resourceXrayPolicyRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Xray

	policy, resp, err := c.V1.Policies.GetPolicy(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Xray policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	if err := d.Set("name", *policy.Name); err != nil {
		return err
	}
	if err := d.Set("type", *policy.Type); err != nil {
		return err
	}
	if err := d.Set("description", *policy.Description); err != nil {
		return err
	}
	if err := d.Set("author", *policy.Author); err != nil {
		return err
	}
	if err := d.Set("created", *policy.Created); err != nil {
		return err
	}
	if err := d.Set("modified", *policy.Modified); err != nil {
		return err
	}
	if err := d.Set("rules", flattenRules(*policy.Rules)); err != nil {
		return err
	}
	return nil
}

func resourceXrayPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Xray

	policy := expandPolicy(d)
	_, err := c.V1.Policies.UpdatePolicy(context.Background(), d.Id(), policy)
	if err != nil {
		return err
	}

	d.SetId(*policy.Name)
	return resourceXrayPolicyRead(d, m)
}

func resourceXrayPolicyDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Xray

	resp, err := c.V1.Policies.DeletePolicy(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}
