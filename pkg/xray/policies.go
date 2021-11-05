package xray

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
)

// PolicyCVSSRange All of this was ripped out of it's original location and copied here.
// The original provider code used th v1 API, which is not supported in the JFROG go client
// The current objective is to rip out any dependencies not from jfrog. So, jfrog doesn't support it
// I need backward compatibility, and I can't have any other dependencies.

type PolicyCVSSRange struct {
	To   float64 `json:"to,omitempty"`
	From float64 `json:"from,omitempty"`
}

type PolicyRuleCriteria struct {
	// Security Criteria
	MinimumSeverity string           `json:"min_severity,omitempty"`
	CVSSRange       *PolicyCVSSRange `json:"cvss_range,omitempty"`

	// License Criteria
	AllowUnknown           *bool     `json:"allow_unknown,omitempty"`
	MultiLicensePermissive *bool     `json:"multi_license_permissive,omitempty"`
	BannedLicenses         *[]string `json:"banned_licenses,omitempty"`
	AllowedLicenses        *[]string `json:"allowed_licenses,omitempty"`
}

type BlockDownloadSettings struct {
	Unscanned *bool `json:"unscanned,omitempty"`
	Active    *bool `json:"active,omitempty"`
}

type PolicyRuleActions struct {
	Webhooks                *[]string              `json:"webhooks,omitempty"`
	Mails                   *[]string              `json:"mails,omitempty"`
	FailBuild               *bool                  `json:"fail_build,omitempty"`
	BlockDownload           *BlockDownloadSettings `json:"block_download,omitempty"`
	BlockReleaseBundle      *bool                  `json:"block_release_bundle_distribution,omitempty"`
	NotifyWatchRecipients   *bool                  `json:"notify_watch_recipients,omitempty"`
	NotifyDeployer          *bool                  `json:"notify_deployer,omitempty"`
	CreateJiraTicketEnabled *bool                  `json:"create_ticket_enabled,omitempty"`
	FailureGracePeriodDays  *int                   `json:"build_failure_grace_period_in_days,omitempty"`
	// License Actions
	CustomSeverity *string `json:"custom_severity,omitempty"`
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
		if *policyType == "license" {
			rule.Criteria, err = expandLicenseCriteria(data["criteria"].([]interface{}))
		}
		if *policyType == "security" {
			rule.Criteria, err = expandSecurityCriteria(data["criteria"].([]interface{}))
		}

		if v, ok := data["actions"]; ok {
			rule.Actions = expandActions(v.([]interface{}))
		}
		rules = append(rules, *rule)
	}

	return rules, err
}

func expandLicenseCriteria(l []interface{}) (*PolicyRuleCriteria, error) {
	if len(l) == 0 {
		return nil, nil
	}
	m := l[0].(map[string]interface{}) // We made this a list of one to make schema validation easier
	criteria := new(PolicyRuleCriteria)
	if (m["allow_unknown"]) != nil {
		criteria.AllowUnknown = BoolPtr(m["allow_unknown"].(bool))
	}
	if (m["banned_licenses"]) != nil {
		criteria.BannedLicenses = expandLicenses(m["banned_licenses"].([]interface{}))
	}
	if (m["allowed_licenses"]) != nil {
		criteria.AllowedLicenses = expandLicenses(m["allowed_licenses"].([]interface{}))
	}
	if (m["multi_license_permissive"]) != nil {
		criteria.MultiLicensePermissive = BoolPtr(m["multi_license_permissive"].(bool))
	}

	return criteria, nil
}

func expandSecurityCriteria(l []interface{}) (*PolicyRuleCriteria, error) {
	if len(l) == 0 {
		return nil, nil
	}
	m := l[0].(map[string]interface{}) // We made this a list of one to make schema validation easier
	criteria := new(PolicyRuleCriteria)

	if (m["min_severity"]) != nil {
		criteria.MinimumSeverity = m["min_severity"].(string)
	}
	if (m["cvss_range"]) != nil {
		criteria.CVSSRange = expandCVSSRange(m["cvss_range"].([]interface{}))
	}
	return criteria, nil
}

func expandCVSSRange(l []interface{}) *PolicyCVSSRange {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})
	cvssrange := &PolicyCVSSRange{
		From: m["from"].(float64),
		To:   m["to"].(float64),
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
	if v, ok := m["webhooks"]; ok && len(v.([]interface{})) > 0 {
		m := v.([]interface{})
		webhooks := make([]string, 0, len(m))
		for _, hook := range m {
			webhooks = append(webhooks, hook.(string))
		}
		actions.Webhooks = &webhooks // if webhook in not set - the empty value is not passing to the payload
	}
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
	if v, ok := m["block_release_bundle_distribution"]; ok {
		actions.BlockReleaseBundle = BoolPtr(v.(bool)) // TODO: why do we need to return the poiner instead of just assigning the value?
	}
	//
	if v, ok := m["notify_watch_recipients"]; ok {
		actions.NotifyWatchRecipients = BoolPtr(v.(bool))
	}
	if v, ok := m["block_release_bundle_distribution"]; ok {
		actions.BlockReleaseBundle = BoolPtr(v.(bool))
	}
	if v, ok := m["notify_deployer"]; ok {
		actions.NotifyDeployer = BoolPtr(v.(bool))
	}
	if v, ok := m["create_ticket_enabled"]; ok {
		actions.CreateJiraTicketEnabled = BoolPtr(v.(bool))
	}
	if v, ok := m["build_failure_grace_period_in_days"]; ok {
		actions.FailureGracePeriodDays = IntPtr(v.(int))
	}
	if v, ok := m["custom_severity"]; ok {
		gosucks := v.(string)
		actions.CustomSeverity = &gosucks
	}

	return actions
}

func flattenLicenseRules(rules []PolicyRule) []interface{} {
	l := make([]interface{}, len(rules))

	for i, rule := range rules {
		m := map[string]interface{}{
			"name":     *rule.Name,
			"priority": *rule.Priority,
			"criteria": flattenLicenseCriteria(rule.Criteria),
			"actions":  flattenActions(rule.Actions),
		}
		l[i] = m
	}

	return l
}

func flattenSecurityRules(rules []PolicyRule) []interface{} {
	l := make([]interface{}, len(rules))

	for i, rule := range rules {
		m := map[string]interface{}{
			"name":     *rule.Name,
			"priority": *rule.Priority,
			"criteria": flattenSecurityCriteria(rule.Criteria),
			"actions":  flattenActions(rule.Actions),
		}
		l[i] = m
	}

	return l
}

func flattenLicenseCriteria(criteria *PolicyRuleCriteria) []interface{} {
	if criteria == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{}

	if criteria.BannedLicenses != nil {
		m["banned_licenses"] = *criteria.BannedLicenses
	}
	if criteria.AllowedLicenses != nil {
		m["allowed_licenses"] = *criteria.AllowedLicenses
	}
	if criteria.AllowUnknown != nil {
		m["allow_unknown"] = *criteria.AllowUnknown // Same typo in the library
	}
	if criteria.MultiLicensePermissive != nil {
		m["multi_license_permissive"] = *criteria.MultiLicensePermissive
	}

	return []interface{}{m}
}

func flattenSecurityCriteria(criteria *PolicyRuleCriteria) []interface{} {
	if criteria == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{}

	if criteria.CVSSRange != nil {
		m["cvss_range"] = flattenCVSSRange(criteria.CVSSRange)
	}
	m["min_severity"] = criteria.MinimumSeverity

	return []interface{}{m}
}

func flattenCVSSRange(cvss *PolicyCVSSRange) []interface{} {
	if cvss == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"from": cvss.From,
		"to":   cvss.To,
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
	if actions.Webhooks != nil {
		m["webhooks"] = *actions.Webhooks
	}
	if actions.Mails != nil {
		m["mails"] = *actions.Mails
	}
	if actions.FailBuild != nil {
		m["fail_build"] = *actions.FailBuild
	}
	if actions.BlockReleaseBundle != nil {
		m["block_release_bundle_distribution"] = *actions.BlockReleaseBundle
	}
	if actions.NotifyWatchRecipients != nil {
		m["notify_watch_recipients"] = *actions.NotifyWatchRecipients
	}
	if actions.NotifyDeployer != nil {
		m["notify_deployer"] = *actions.NotifyDeployer
	}
	if actions.CreateJiraTicketEnabled != nil {
		m["create_ticket_enabled"] = *actions.CreateJiraTicketEnabled
	}
	if actions.FailureGracePeriodDays != nil {
		m["build_failure_grace_period_in_days"] = *actions.FailureGracePeriodDays // integer
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

// Create Xray policy
func resourceXrayPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	policy, err := expandPolicy(d)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = m.(*resty.Client).R().SetBody(policy).Post("xray/api/v2/policies")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*policy.Name)
	resourceXrayPolicyRead(ctx, d, m)
	return diags
}

// Get Xray policy by name
func getPolicy(id string, client *resty.Client) (Policy, *resty.Response, error) {
	policy := Policy{}
	resp, err := client.R().SetResult(&policy).Get("xray/api/v2/policies/" + id)
	return policy, resp, err
}
func resourceXrayPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	policy, resp, err := getPolicy(d.Id(), m.(*resty.Client))
	var diags diag.Diagnostics
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			log.Printf("[WARN] Xray policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("name", policy.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", policy.Type); err != nil {
		return diag.FromErr(err)
	}
	if policy.Description != nil {
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
	if *policy.Type == "license" {
		if err := d.Set("rules", flattenLicenseRules(*policy.Rules)); err != nil {
			return diag.FromErr(err)
		}
	}
	if *policy.Type == "security" {
		if err := d.Set("rules", flattenSecurityRules(*policy.Rules)); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceXrayPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	policy, err := expandPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = m.(*resty.Client).R().SetBody(policy).Put("xray/api/v2/policies/" + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*policy.Name)
	return resourceXrayPolicyRead(ctx, d, m)
}

func resourceXrayPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	_, err := m.(*resty.Client).R().Delete("xray/api/v2/policies/" + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
