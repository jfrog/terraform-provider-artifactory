package xray

import (
	"context"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type PolicyCVSSRange struct {
	To   *float64 `json:"to,omitempty"`
	From *float64 `json:"from,omitempty"`
}

type PolicyRuleCriteria struct {
	// Security Criteria
	MinimumSeverity string           `json:"min_severity,omitempty"` // Omitempty is used because the empty field is conflicting with CVSSRange
	CVSSRange       *PolicyCVSSRange `json:"cvss_range,omitempty"`
	// We use pointer for CVSSRange to address nil-verification for non-primitive types.
	// Unlike primitive types, when the non-primitive type in the struct is set
	// to nil, the empty key will be created in the JSON body anyway.
	// Since CVSSRange is conflicting with MinimumSeverity, Xray will throw an error if .
	// Pointer can be set to nil value, so we can remove CVSSRange entirely only
	// if it's a pointer.
	// The nil pointer is used in conjunction with the omitempty flag in the JSON tag,
	// to remove the key completely in the payload.

	// License Criteria
	AllowUnknown           *bool    `json:"allow_unknown,omitempty"`            // Omitempty is used because the empty field is conflicting with MultiLicensePermissive
	MultiLicensePermissive *bool    `json:"multi_license_permissive,omitempty"` // Omitempty is used because the empty field is conflicting with AllowUnknown
	BannedLicenses         []string `json:"banned_licenses,omitempty"`
	AllowedLicenses        []string `json:"allowed_licenses,omitempty"`
}

type BlockDownloadSettings struct {
	Unscanned bool `json:"unscanned"`
	Active    bool `json:"active"`
}

type PolicyRuleActions struct {
	Webhooks                []string              `json:"webhooks"`
	Mails                   []string              `json:"mails"`
	FailBuild               bool                  `json:"fail_build"`
	BlockDownload           BlockDownloadSettings `json:"block_download"`
	BlockReleaseBundle      bool                  `json:"block_release_bundle_distribution"`
	NotifyWatchRecipients   bool                  `json:"notify_watch_recipients"`
	NotifyDeployer          bool                  `json:"notify_deployer"`
	CreateJiraTicketEnabled bool                  `json:"create_ticket_enabled"`
	FailureGracePeriodDays  int                   `json:"build_failure_grace_period_in_days"`
	// License Actions
	CustomSeverity string `json:"custom_severity"`
}

type PolicyRule struct {
	Name     string              `json:"name"`
	Priority int                 `json:"priority"`
	Criteria *PolicyRuleCriteria `json:"criteria"`
	Actions  PolicyRuleActions   `json:"actions"`
}

type Policy struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Author      string        `json:"author,omitempty"` // Omitempty is used because the field is computed
	Description string        `json:"description"`
	Rules       *[]PolicyRule `json:"rules"`
	Created     string        `json:"created,omitempty"`  // Omitempty is used because the field is computed
	Modified    string        `json:"modified,omitempty"` // Omitempty is used because the field is computed
}

func unpackPolicy(d *schema.ResourceData) (*Policy, error) {
	policy := new(Policy)

	policy.Name = d.Get("name").(string)
	if v, ok := d.GetOk("type"); ok {
		policy.Type = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}
	if v, ok := d.GetOk("author"); ok {
		policy.Author = v.(string)
	}
	policyRules, err := unpackRules(d.Get("rule").([]interface{}), policy.Type)
	policy.Rules = &policyRules

	return policy, err
}

func unpackRules(configured []interface{}, policyType string) (policyRules []PolicyRule, err error) {
	var rules []PolicyRule

	if configured != nil {
		for _, raw := range configured {
			rule := new(PolicyRule)
			data := raw.(map[string]interface{})
			rule.Name = data["name"].(string)
			rule.Priority = data["priority"].(int)

			rule.Criteria, err = unpackCriteria(data["criteria"].(*schema.Set), policyType)
			if v, ok := data["actions"]; ok {
				rule.Actions = unpackActions(v.(*schema.Set))
			}
			rules = append(rules, *rule)
		}
	}

	return rules, err
}

func unpackCriteria(d *schema.Set, policyType string) (*PolicyRuleCriteria, error) {
	tfCriteria := d.List()
	if len(tfCriteria) == 0 {
		return nil, nil
	}

	m := tfCriteria[0].(map[string]interface{}) // We made this a list of one to make schema validation easier
	criteria := new(PolicyRuleCriteria)
	// The API doesn't allow both severity and license criteria to be _set_, even if they have empty values
	// So we have to figure out which group is actually empty and not even set it
	if policyType == "license" {
		if v, ok := m["allow_unknown"]; ok {
			criteria.AllowUnknown = BoolPtr(v.(bool))
		}
		if v, ok := m["banned_licenses"]; ok {
			criteria.BannedLicenses = unpackLicenses(v.(*schema.Set))
		}
		if v, ok := m["allowed_licenses"]; ok {
			criteria.AllowedLicenses = unpackLicenses(v.(*schema.Set))
		}
		if v, ok := m["multi_license_permissive"]; ok {
			criteria.MultiLicensePermissive = BoolPtr(v.(bool))
		}
	} else {
		minSev := m["min_severity"].(string)
		cvss := unpackCVSSRange(m["cvss_range"].([]interface{}))

		// This is also picky about not allowing empty values to be set
		if cvss == nil {
			criteria.MinimumSeverity = minSev
		} else {
			criteria.CVSSRange = cvss
		}
	}
	return criteria, nil
}

func unpackCVSSRange(l []interface{}) *PolicyCVSSRange {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})
	cvssrange := &PolicyCVSSRange{
		From: Float64Ptr(m["from"].(float64)),
		To:   Float64Ptr(m["to"].(float64)),
	}
	return cvssrange
}

func unpackLicenses(d *schema.Set) []string {
	var licenses []string
	for _, license := range d.List() {
		licenses = append(licenses, license.(string))
	}
	return licenses
}

func unpackActions(l *schema.Set) PolicyRuleActions {
	actions := PolicyRuleActions{}
	policyActions := l.List()
	m := policyActions[0].(map[string]interface{}) // We made this a list of one to make schema validation easier

	if v, ok := m["webhooks"]; ok {
		m := v.(*schema.Set).List()
		var webhooks []string
		for _, hook := range m {
			webhooks = append(webhooks, hook.(string))
		}
		actions.Webhooks = webhooks
	}
	if v, ok := m["mails"]; ok {
		m := v.(*schema.Set).List()
		var mails []string
		for _, mail := range m {
			mails = append(mails, mail.(string))
		}
		actions.Mails = mails
	}
	if v, ok := m["fail_build"]; ok {
		actions.FailBuild = v.(bool)
	}

	if v, ok := m["block_download"]; ok {
		if len(v.(*schema.Set).List()) > 0 {
			vList := v.(*schema.Set).List()
			vMap := vList[0].(map[string]interface{})

			actions.BlockDownload = BlockDownloadSettings{
				Unscanned: vMap["unscanned"].(bool),
				Active:    vMap["active"].(bool),
			}
		} else {
			actions.BlockDownload = BlockDownloadSettings{
				Unscanned: false,
				Active:    false,
			}
			// Setting this false/false block feels like it _should_ work, since putting a false/false block in the terraform resource works fine
			// However, it doesn't, and we end up getting this diff when running acceptance tests when this is optional in the schema
			// rule.0.actions.0.block_download.#:           "1" => "0"
			// rule.0.actions.0.block_download.0.active:    "false" => ""
			// rule.0.actions.0.block_download.0.unscanned: "false" => ""
		}
	}
	if v, ok := m["block_release_bundle_distribution"]; ok {
		actions.BlockReleaseBundle = v.(bool)
	}

	if v, ok := m["notify_watch_recipients"]; ok {
		actions.NotifyWatchRecipients = v.(bool)
	}
	if v, ok := m["block_release_bundle_distribution"]; ok {
		actions.BlockReleaseBundle = v.(bool)
	}
	if v, ok := m["notify_deployer"]; ok {
		actions.NotifyDeployer = v.(bool)
	}
	if v, ok := m["create_ticket_enabled"]; ok {
		actions.CreateJiraTicketEnabled = v.(bool)
	}
	if v, ok := m["build_failure_grace_period_in_days"]; ok {
		actions.FailureGracePeriodDays = v.(int)
	}
	if v, ok := m["custom_severity"]; ok {
		actions.CustomSeverity = v.(string)
	}

	return actions
}

func packRules(rules []PolicyRule, policyType string) []interface{} {
	var l []interface{}

	for _, rule := range rules {
		var criteria []interface{}
		var isLicense bool

		switch policyType {
		case "license":
			criteria = packLicenseCriteria(rule.Criteria)
			isLicense = true
		case "security":
			criteria = packSecurityCriteria(rule.Criteria)
			isLicense = false
		}

		m := map[string]interface{}{
			"name":     rule.Name,
			"priority": rule.Priority,
			"criteria": criteria,
			"actions":  packActions(rule.Actions, isLicense),
		}

		l = append(l, m)
	}

	return l
}

func packLicenseCriteria(criteria *PolicyRuleCriteria) []interface{} {

	m := map[string]interface{}{}

	if criteria.BannedLicenses != nil {
		m["banned_licenses"] = criteria.BannedLicenses
	}
	if criteria.AllowedLicenses != nil {
		m["allowed_licenses"] = criteria.AllowedLicenses
	}
	m["allow_unknown"] = criteria.AllowUnknown
	m["multi_license_permissive"] = criteria.MultiLicensePermissive

	return []interface{}{m}
}

func packSecurityCriteria(criteria *PolicyRuleCriteria) []interface{} {
	m := map[string]interface{}{}
	// cvss_range and min_severity are conflicting, only one can be present in the JSON
	m["cvss_range"] = packCVSSRange(criteria.CVSSRange)
	m["min_severity"] = criteria.MinimumSeverity

	return []interface{}{m}
}

func packCVSSRange(cvss *PolicyCVSSRange) []interface{} {
	if cvss == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"from": *cvss.From,
		"to":   *cvss.To,
	}
	return []interface{}{m}
}

func packActions(actions PolicyRuleActions, license bool) []interface{} {

	m := map[string]interface{}{
		"block_download":                     packBlockDownload(actions.BlockDownload),
		"webhooks":                           actions.Webhooks,
		"mails":                              actions.Mails,
		"fail_build":                         actions.FailBuild,
		"block_release_bundle_distribution":  actions.BlockReleaseBundle,
		"notify_watch_recipients":            actions.NotifyWatchRecipients,
		"notify_deployer":                    actions.NotifyDeployer,
		"create_ticket_enabled":              actions.CreateJiraTicketEnabled,
		"build_failure_grace_period_in_days": actions.FailureGracePeriodDays,
	}

	if license {
		m["custom_severity"] = actions.CustomSeverity
	}

	return []interface{}{m}
}

func packBlockDownload(bd BlockDownloadSettings) []interface{} {

	m := map[string]interface{}{}
	m["unscanned"] = bd.Unscanned
	m["active"] = bd.Active
	return []interface{}{m}
}

func packPolicy(policy Policy, d *schema.ResourceData) diag.Diagnostics {
	if err := d.Set("name", policy.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", policy.Type); err != nil {
		return diag.FromErr(err)
	}
	if len(policy.Description) > 0 {
		if err := d.Set("description", policy.Description); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("author", policy.Author); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", policy.Created); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("modified", policy.Modified); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("rule", packRules(*policy.Rules, policy.Type)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getPolicy(id string, client *resty.Client) (Policy, *resty.Response, error) {
	policy := Policy{}
	resp, err := client.R().SetResult(&policy).Get("xray/api/v2/policies/" + id)
	return policy, resp, err
}

func resourceXrayPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	policy, err := unpackPolicy(d)
	// Warning or errors can be collected in a slice type
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = m.(*resty.Client).R().SetBody(policy).Post("xray/api/v2/policies")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Name)
	return resourceXrayPolicyRead(ctx, d, m)
}

func resourceXrayPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	policy, resp, err := getPolicy(d.Id(), m.(*resty.Client))
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			log.Printf("[WARN] Xray policy (%s) not found, removing from state", d.Id())
			d.SetId("")
		}
		return diag.FromErr(err)
	}
	packPolicy(policy, d)
	return nil
}

func resourceXrayPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	policy, err := unpackPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = m.(*resty.Client).R().SetBody(policy).Put("xray/api/v2/policies/" + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Name)
	return resourceXrayPolicyRead(ctx, d, m)
}

func resourceXrayPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	resp, err := m.(*resty.Client).R().Delete("xray/api/v2/policies/" + d.Id())
	if err != nil && resp.StatusCode() == http.StatusInternalServerError {
		d.SetId("")
		return diag.FromErr(err)
	}
	return nil
}
