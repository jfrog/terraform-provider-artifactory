package xray

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// PolicyCVSSRange All of this was ripped out of it's original location and copied here.
// The original provider code used th v1 API, which is not supported in the JFROG go client
// The current objective is to rip out any dependencies not from jfrog. So, jfrog doesn't support it
// I need backward compatibility, and I can't have any other dependencies.
type PolicyCVSSRange struct {
	To   *int `json:"to,omitempty"`
	From *int `json:"from,omitempty"`
}

type PolicyRuleCriteria struct {
	// Security Criteria
	MinimumSeverity *string          `json:"min_severity,omitempty"`
	CVSSRange       *PolicyCVSSRange `json:"cvss_range,omitempty"`

	// License Criteria
	AllowUnknown    *bool     `json:"allow_unknown,omitempty"`
	BannedLicenses  *[]string `json:"banned_licenses,omitempty"`
	AllowedLicenses *[]string `json:"allowed_licenses,omitempty"`
}

type BlockDownloadSettings struct {
	Unscanned *bool `json:"unscanned,omitempty"`
	Active    *bool `json:"active,omitempty"`
}

type PolicyRuleActions struct {
	Mails          *[]string              `json:"mails,omitempty"`
	FailBuild      *bool                  `json:"fail_build,omitempty"`
	BlockDownload  *BlockDownloadSettings `json:"block_download,omitempty"`
	Webhooks       *[]string              `json:"webhooks,omitempty"`
	CustomSeverity *string                `json:"custom_severity,omitempty"`
}

type PolicyRule struct {
	Name     *string             `json:"name,omitempty"`
	Priority *int                `json:"priority,omitempty"`
	Criteria *PolicyRuleCriteria `json:"criteria,omitempty"`
	Actions  *PolicyRuleActions  `json:"actions,omitempty"`
}

type Policy struct {
	Name        *string       `json:"name,omitempty"`
	Type        *string       `json:"type,omitempty"`
	Author      *string       `json:"author,omitempty"`
	Description *string       `json:"description,omitempty"`
	Rules       *[]PolicyRule `json:"rules,omitempty"`
	Created     *string       `json:"created,omitempty"`
	Modified    *string       `json:"modified,omitempty"`
}

func resourceXrayPolicy() *schema.Resource {
	return &schema.Resource{
		SchemaVersion:      1,
		Create:             resourceXrayPolicyCreate,
		Read:               resourceXrayPolicyRead,
		Update:             resourceXrayPolicyUpdate,
		Delete:             resourceXrayPolicyDelete,
		DeprecationMessage: "This portion of the provider uses V1 apis and will eventually be removed",
		Description: "Creates an xray policy using V1 of the underlying APIs. Please note: " +
			"It's only compatible with Bearer token auth method (Identity and Access => Access Tokens",

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
										Type:     schema.TypeString,
										Optional: true,
									},
									"cvss_range": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
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
										Type:     schema.TypeBool,
										Optional: true,
									},
									"banned_licenses": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"allowed_licenses": {
										Type:     schema.TypeList,
										Optional: true,
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

func expandPolicy(d *schema.ResourceData) (*Policy, error) {
	policy := new(Policy)

	policy.Name = StringPtr(d.Get("name").(string))
	if v, ok := d.GetOk("type"); ok {
		policy.Type = StringPtr(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		policy.Description = StringPtr(v.(string))
	}
	if v, ok := d.GetOk("author"); ok {
		policy.Author = StringPtr(v.(string))
	}
	policyRules, err := expandRules(d.Get("rules").([]interface{}), policy.Type)
	policy.Rules = &policyRules

	return policy, err
}

func expandRules(configured []interface{}, policyType *string) (policyRules []PolicyRule, err error) {
	rules := make([]PolicyRule, 0, len(configured))

	for _, raw := range configured {
		rule := new(PolicyRule)
		data := raw.(map[string]interface{})
		rule.Name = StringPtr(data["name"].(string))
		rule.Priority = IntPtr(data["priority"].(int))

		rule.Criteria, err = expandCriteria(data["criteria"].([]interface{}), policyType)
		if v, ok := data["actions"]; ok {
			rule.Actions = expandActions(v.([]interface{}))
		}
		rules = append(rules, *rule)
	}

	return rules, err
}

func expandCriteria(l []interface{}, policyType *string) (*PolicyRuleCriteria, error) {
	if len(l) == 0 {
		return nil, nil
	}

	m := l[0].(map[string]interface{}) // We made this a list of one to make schema validation easier
	criteria := new(PolicyRuleCriteria)

	// The API doesn't allow both severity and license criteria to be _set_, even if they have empty values
	// So we have to figure out which group is actually empty and not even set it
	minSev := StringPtr(m["min_severity"].(string))
	cvss := expandCVSSRange(m["cvss_range"].([]interface{}))
	allowUnk := BoolPtr(m["allow_unknown"].(bool))
	banned := expandLicenses(m["banned_licenses"].([]interface{}))
	allowed := expandLicenses(m["allowed_licenses"].([]interface{}))

	licenseType := "license"
	securityType := "security"

	if *minSev == "" && cvss == nil {
		if *policyType == securityType {
			return nil, fmt.Errorf("allow_unknown, banned_licenses, and allowed_licenses are not supported with security policies")
		}

		// If these are both the default values, we must be using license criteria
		criteria.AllowUnknown = allowUnk
		criteria.BannedLicenses = banned
		criteria.AllowedLicenses = allowed
	} else {
		if *policyType == licenseType {
			return nil, fmt.Errorf("min_severity and cvvs_range are not supported with license policies")
		}

		// This is also picky about not allowing empty values to be set
		if cvss == nil {
			criteria.MinimumSeverity = minSev
		} else {
			criteria.CVSSRange = cvss
		}
	}

	return criteria, nil
}

func expandCVSSRange(l []interface{}) *PolicyCVSSRange {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})
	cvssrange := &PolicyCVSSRange{
		From: IntPtr(m["from"].(int)),
		To:   IntPtr(m["to"].(int)),
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

func expandActions(l []interface{}) *PolicyRuleActions {
	if len(l) == 0 {
		return nil
	}

	actions := new(PolicyRuleActions)
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
		actions.FailBuild = BoolPtr(v.(bool))
	}

	if v, ok := m["block_download"]; ok {
		if len(v.([]interface{})) > 0 {
			vList := v.([]interface{})
			vMap := vList[0].(map[string]interface{})

			actions.BlockDownload = &BlockDownloadSettings{
				Unscanned: BoolPtr(vMap["unscanned"].(bool)),
				Active:    BoolPtr(vMap["active"].(bool)),
			}
		} else {
			actions.BlockDownload = &BlockDownloadSettings{
				Unscanned: BoolPtr(false),
				Active:    BoolPtr(false),
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
		gosucks := v.(string)
		actions.CustomSeverity = &gosucks
	}

	return actions
}

func flattenRules(rules []PolicyRule) []interface{} {
	l := make([]interface{}, len(rules))

	for i, rule := range rules {
		m := map[string]interface{}{
			"name":     *rule.Name,
			"priority": *rule.Priority,
			"criteria": flattenCriteria(rule.Criteria),
			"actions":  flattenActions(rule.Actions),
		}
		l[i] = m
	}

	return l
}

func flattenCriteria(criteria *PolicyRuleCriteria) []interface{} {
	if criteria == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cvss_range": flattenCVSSRange(criteria.CVSSRange),
	}

	if criteria.MinimumSeverity != nil {
		m["min_severity"] = *criteria.MinimumSeverity
	}
	if criteria.AllowUnknown != nil {
		m["allow_unknown"] = *criteria.AllowUnknown // Same typo in the library
	}
	if criteria.BannedLicenses != nil {
		m["banned_licenses"] = *criteria.BannedLicenses
	}
	if criteria.AllowedLicenses != nil {
		m["allowed_licenses"] = *criteria.AllowedLicenses
	}

	return []interface{}{m}
}

func flattenCVSSRange(cvss *PolicyCVSSRange) []interface{} {
	if cvss == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"from": *cvss.From,
		"to":   *cvss.To,
	}
	return []interface{}{m}
}

func flattenActions(actions *PolicyRuleActions) []interface{} {
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

func flattenBlockDownload(bd *BlockDownloadSettings) []interface{} {
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

	policy, err := expandPolicy(d)
	if err != nil {
		return err
	}
	_, err = m.(*resty.Client).R().SetBody(policy).Post("xray/api/v1/policies")
	if err != nil {
		return err
	}

	d.SetId(*policy.Name)
	return resourceXrayPolicyRead(d, m)
}

func getPolicy(id string, client *resty.Client) (Policy, *resty.Response, error) {
	policy := Policy{}
	resp, err := client.R().SetResult(&policy).Put("xray/api/v1/policies/" + id)
	return policy, resp, err
}
func resourceXrayPolicyRead(d *schema.ResourceData, m interface{}) error {
	policy, resp, err := getPolicy(d.Id(), m.(*resty.Client))
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			log.Printf("[WARN] Xray policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if err := d.Set("name", *policy.Name); err != nil {
		return err
	}
	if err := d.Set("type", *policy.Type); err != nil {
		return err
	}
	if policy.Description != nil {
		if err := d.Set("description", *policy.Description); err != nil {
			return err
		}
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

	policy, err := expandPolicy(d)
	if err != nil {
		return err
	}
	_, err = m.(*resty.Client).R().SetBody(policy).Put("xray/api/v1/policies/" + d.Id())
	if err != nil {
		return err
	}

	d.SetId(*policy.Name)
	return resourceXrayPolicyRead(d, m)
}

func resourceXrayPolicyDelete(d *schema.ResourceData, m interface{}) error {
	_, err := m.(*resty.Client).R().Delete("xray/api/v1/policies/" + d.Id())
	return err
}
